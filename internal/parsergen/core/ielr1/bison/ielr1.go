package bison

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/trace"
	"strconv"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	bisonutils "github.com/backbone81/golr/internal/utils/bison"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// The policy factory is ignored. GNU Bison resolves the conflicts itself, with its own precedence and associativity
// rules, and this core only reads the tables it reports back. The parameter is there so that this core has the same
// signature as the GoLR one and a caller can switch between them.
func GrammarToParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: IELR1: Bison: GrammarToParser").End()

	builder := NewIELR1(grammar)
	parser, err := builder.BuildParser()
	// Note that we currently do not capture reported conflicts from GNU Bison. Therefore, we return no conflicts.
	return parser, nil, err
}

type IELR1 struct {
	// grammar is the grammar as the caller handed it in. It is what the GNU Bison grammar file is written from, and it
	// must not be augmented for that: GNU Bison introduces the new start symbol and the end of input marker itself.
	grammar frontend.Grammar

	// augmentedGrammar is that grammar with the new start symbol and the end of input marker, which is what every
	// index of the resulting parser refers to. The GoLR cores build exactly this and hand it on unchanged, so taking
	// the numbering from here rather than from the report is what makes the two cores agree on it.
	augmentedGrammar frontend.Grammar

	// terminalIdxByName and nonterminalIdxByName translate a symbol of the report into an index of the augmented
	// grammar. They are keyed on the name GNU Bison knows the symbol under, because that is the name the report uses.
	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int

	// productionIdxByRuleNumber translates a rule number of the report into a production index of the augmented
	// grammar, which are not the same numbering. See buildProductionMapping.
	productionIdxByRuleNumber map[int]int
}

func NewIELR1(grammar frontend.Grammar) *IELR1 {
	augmentedGrammar := frontend.AugmentGrammar(grammar)
	result := &IELR1{
		grammar:                   grammar,
		augmentedGrammar:          augmentedGrammar,
		terminalIdxByName:         make(map[string]int, len(augmentedGrammar.Terminals)),
		nonterminalIdxByName:      make(map[string]int, len(augmentedGrammar.Nonterminals)),
		productionIdxByRuleNumber: make(map[int]int, len(augmentedGrammar.Productions)),
	}

	for terminalIdx, terminal := range augmentedGrammar.Terminals {
		result.terminalIdxByName[bisonfrontend.ToBisonSymbolName(terminal.Name)] = terminalIdx
	}
	for nonterminalIdx, nonterminal := range augmentedGrammar.Nonterminals {
		result.nonterminalIdxByName[nonterminal.Name] = nonterminalIdx
	}
	return result
}

func (i *IELR1) BuildParser() (parser backend.Parser, err error) { //nolint:nonamedreturns // Required for defer
	bisonGrammarFile, err := os.CreateTemp("", "golr-ielr1-*.y")
	if err != nil {
		return backend.Parser{}, err
	}
	defer func() {
		if removeErr := os.Remove(bisonGrammarFile.Name()); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("removing file: %w", removeErr))
		}
	}()

	if err := bisonfrontend.FromGrammar(bisonGrammarFile, i.grammar); err != nil {
		return backend.Parser{}, err
	}

	bisonXmlFile, err := os.CreateTemp("", "golr-ielr1-*.xml")
	if err != nil {
		return backend.Parser{}, err
	}
	defer func() {
		if removeErr := os.Remove(bisonXmlFile.Name()); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("removing file: %w", removeErr))
		}
	}()

	if err := bisonutils.BuildIELR1(bisonGrammarFile.Name(), bisonXmlFile.Name()); err != nil {
		return backend.Parser{}, err
	}

	report, err := bisonutils.LoadBisonXMLReportFromFile(bisonXmlFile.Name())
	if err != nil {
		return backend.Parser{}, err
	}

	// The grammar of the parser is the augmented grammar and never the one rebuilt from the report, so that the symbols
	// and productions are numbered the way the GoLR cores number them.
	parser.Grammar = i.augmentedGrammar

	if err := i.buildProductionMapping(report); err != nil {
		return backend.Parser{}, err
	}
	if err := i.buildStateList(report, &parser); err != nil {
		return backend.Parser{}, err
	}
	return parser, nil
}

// buildProductionMapping maps every rule number of the report onto the index of the production it was written from.
//
// The two numberings differ. GNU Bison moves the nonterminals which cannot be reached from the start symbol, and their
// rules, to the end of its numbering so that it can truncate them away afterwards (nonterminals_reduce in
// src/reduce.c). GoLR keeps those symbols, so the relocation must not be copied, and a rule has to be found by what it
// is rather than by where it stands.
//
// What a rule is, is its left hand side and the symbols on its right, because that is all the report carries. Two
// productions which agree on both are indistinguishable here and are matched in the order they appear. They derive the
// same sentences, so which of the two a state reduces by cannot change the language the parser accepts.
func (i *IELR1) buildProductionMapping(report bisonutils.BisonXMLReport) error {
	productionIdxsBySignature := make(map[string][]int, len(i.augmentedGrammar.Productions))
	for productionIdx, production := range i.augmentedGrammar.Productions {
		signature := i.productionSignature(production)
		productionIdxsBySignature[signature] = append(productionIdxsBySignature[signature], productionIdx)
	}

	for _, rule := range report.Grammar.Rules {
		signature := ruleSignature(rule)
		productionIdxs := productionIdxsBySignature[signature]
		if len(productionIdxs) == 0 {
			return fmt.Errorf("rule %d of the GNU Bison report matches no production of %q", rule.Number, rule.Lhs)
		}
		i.productionIdxByRuleNumber[rule.Number] = productionIdxs[0]
		productionIdxsBySignature[signature] = productionIdxs[1:]
	}
	return nil
}

// productionSignature renders a production the way the report names the same rule, so that the two can be matched up.
// The symbols are separated by a space, which no symbol name can contain, and the signature therefore reads like the
// rule it stands for.
func (i *IELR1) productionSignature(production frontend.Production) string {
	symbolNames := make([]string, 0, len(production.SymbolRefs)+1)
	symbolNames = append(symbolNames, i.augmentedGrammar.Nonterminals[production.NonterminalIdx].Name)

	for _, symbolRef := range production.SymbolRefs {
		if symbolRef.IsTerminal() {
			name := i.augmentedGrammar.Terminals[symbolRef.Idx()].Name
			symbolNames = append(symbolNames, bisonfrontend.ToBisonSymbolName(name))
		} else {
			symbolNames = append(symbolNames, i.augmentedGrammar.Nonterminals[symbolRef.Idx()].Name)
		}
	}
	return strings.Join(symbolNames, " ")
}

// ruleSignature renders a rule of the report the way productionSignature renders the production it came from.
func ruleSignature(rule bisonutils.Rule) string {
	return strings.Join(append([]string{rule.Lhs}, rule.Rhs...), " ")
}

//nolint:gocognit,funlen,cyclop // The state construction loop is branchy; splitting it would obscure the flow.
func (i *IELR1) buildStateList(report bisonutils.BisonXMLReport, parser *backend.Parser) error {
	for _, state := range report.Automaton.States {
		var newState backend.State

		for _, item := range state.ItemSet {
			if !item.IsKernelItem() {
				// The XML report of GNU Bison lists the full closure of a state. We keep only the kernel items, as
				// the closure can always be recalculated from them and the GoLR cores provide the kernel items only.
				continue
			}
			productionIdx, ok := i.productionIdxByRuleNumber[item.RuleNumber]
			if !ok {
				return fmt.Errorf("unknown rule number %d", item.RuleNumber)
			}
			newState.KernelItems.Add(backend.NewCore(productionIdx, item.Dot))
		}

		for _, transition := range state.Transitions {
			var symbolRef frontend.SymbolRef
			if idx, ok := i.terminalIdxByName[transition.Symbol]; ok {
				symbolRef = frontend.NewTerminalRef(idx)
			} else if idx, ok := i.nonterminalIdxByName[transition.Symbol]; ok {
				symbolRef = frontend.NewNonterminalRef(idx)
			} else {
				return fmt.Errorf("unknown transition on %q", transition.Symbol)
			}
			newState.TransitionActions.Add(backend.NewTransitionAction(symbolRef, transition.State))
		}

		// GNU Bison reports the terminals which the state rejects on purpose, which are the ones a nonassociative
		// declaration removed every action for. They have to be kept, because a state which has a default reduce action
		// would otherwise take that action for them, see backend.State.RejectedTerminals.
		for _, stateError := range state.Errors {
			idx, ok := i.terminalIdxByName[stateError.Symbol]
			if !ok {
				return fmt.Errorf("unknown error on %q", stateError.Symbol)
			}
			newState.RejectedTerminals.Add(idx)
		}

		// A single production can reduce on multiple lookahead terminals. Bison reports those as separate reduction
		// entries, but they must collapse into one reduce action whose lookahead set is the union of all terminals.
		// We accumulate the lookaheads per production (in first-seen order for deterministic output) and emit one
		// reduce action per production afterwards.
		lookaheadByProduction := map[int]*backend.LookaheadSet{}
		var productionOrder []int
		for _, reduction := range state.Reductions {
			if !reduction.Enabled {
				// Reductions are disabled to resolve shift reduce conflicts. We ignore disabled reductions.
				continue
			}

			var productionIdx int
			if reduction.Rule == "accept" {
				// The accept rule is the augmented production, which is always the first one.
				productionIdx = 0
			} else {
				ruleNumber, err := strconv.Atoi(reduction.Rule)
				if err != nil {
					return err
				}
				var ok bool
				productionIdx, ok = i.productionIdxByRuleNumber[ruleNumber]
				if !ok {
					return fmt.Errorf("unknown rule number %d", ruleNumber)
				}
			}

			if reduction.Symbol == "$default" {
				newState.DefaultReduceProductionIdx = &productionIdx
				// The default reduce action should not show up as a standard reduce. Therefore skip to the next.
				continue
			}

			terminalIdx, ok := i.terminalIdxByName[reduction.Symbol]
			if !ok {
				return fmt.Errorf("unknown terminal %q", reduction.Symbol)
			}

			lookaheadSet, ok := lookaheadByProduction[productionIdx]
			if !ok {
				lookaheadSet = &backend.LookaheadSet{}
				lookaheadByProduction[productionIdx] = lookaheadSet
				productionOrder = append(productionOrder, productionIdx)
			}
			lookaheadSet.Add(terminalIdx)
		}
		for _, productionIdx := range productionOrder {
			newState.ReduceActions.Add(backend.NewReduceAction(*lookaheadByProduction[productionIdx], productionIdx))
		}
		parser.States = append(parser.States, newState)
	}
	return nil
}
