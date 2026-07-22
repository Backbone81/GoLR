package bison

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/trace"
	"strconv"

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
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: LALR1: Bison: GrammarToParser").End()

	builder := NewLALR1(grammar)
	parser, err := builder.BuildParser()
	// Note that we currently do not capture reported conflicts from GNU Bison. Therefore, we return no conflicts.
	return parser, nil, err
}

type LALR1 struct {
	grammar              frontend.Grammar
	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int
}

func NewLALR1(augmentedGrammar frontend.Grammar) *LALR1 {
	return &LALR1{
		grammar:              augmentedGrammar,
		terminalIdxByName:    make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
	}
}

func (i *LALR1) BuildParser() (parser backend.Parser, err error) { //nolint:nonamedreturns // Required for defer
	bisonGrammarFile, err := os.CreateTemp("", "golr-lalr1-*.y")
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

	bisonXmlFile, err := os.CreateTemp("", "golr-lalr1-*.xml")
	if err != nil {
		return backend.Parser{}, err
	}
	defer func() {
		if removeErr := os.Remove(bisonXmlFile.Name()); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("removing file: %w", removeErr))
		}
	}()

	if err := bisonutils.BuildLALR1(bisonGrammarFile.Name(), bisonXmlFile.Name()); err != nil {
		return backend.Parser{}, err
	}

	report, err := bisonutils.LoadBisonXMLReportFromFile(bisonXmlFile.Name())
	if err != nil {
		return backend.Parser{}, err
	}

	i.buildTerminalList(report, &parser)
	i.buildNonterminalList(report, &parser)
	i.buildProductionList(report, &parser)
	if err := i.buildStateList(report, &parser); err != nil {
		return backend.Parser{}, err
	}
	return parser, nil
}

func (i *LALR1) buildTerminalList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
	for _, terminal := range report.Grammar.Terminals {
		i.terminalIdxByName[terminal.Name] = len(parser.Grammar.Terminals)
		parser.Grammar.Terminals = append(parser.Grammar.Terminals, frontend.Symbol{
			Name: terminal.Name,
		})
	}
}

func (i *LALR1) buildNonterminalList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
	for _, nonterminal := range report.Grammar.Nonterminals {
		i.nonterminalIdxByName[nonterminal.Name] = len(parser.Grammar.Nonterminals)
		parser.Grammar.Nonterminals = append(parser.Grammar.Nonterminals, frontend.Symbol{
			Name: nonterminal.Name,
		})
	}
}

func (i *LALR1) buildProductionList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
	for _, rule := range report.Grammar.Rules {
		var symbolRefs []frontend.SymbolRef
		for _, rhs := range rule.Rhs {
			if idx, ok := i.terminalIdxByName[rhs]; ok {
				symbolRefs = append(symbolRefs, frontend.NewTerminalRef(idx))
			} else {
				symbolRefs = append(symbolRefs, frontend.NewNonterminalRef(i.nonterminalIdxByName[rhs]))
			}
		}
		parser.Grammar.Productions = append(parser.Grammar.Productions, frontend.Production{
			NonterminalIdx: i.nonterminalIdxByName[rule.Lhs],
			SymbolRefs:     symbolRefs,
		})
	}
}

//nolint:gocognit,funlen // The state construction loop is inherently branchy; splitting it would obscure the flow.
func (i *LALR1) buildStateList(report bisonutils.BisonXMLReport, parser *backend.Parser) error {
	for _, state := range report.Automaton.States {
		var newState backend.State

		for _, item := range state.ItemSet {
			if !item.IsKernelItem() {
				// The XML report of GNU Bison lists the full closure of a state. We keep only the kernel items, as
				// the closure can always be recalculated from them and the GoLR cores provide the kernel items only.
				continue
			}
			newState.KernelItems.Add(backend.NewCore(item.RuleNumber, item.Dot))
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

			productionIdx, err := strconv.Atoi(reduction.Rule)
			if err != nil {
				if reduction.Rule == "accept" {
					// The accept rule is always the first production
					productionIdx = 0
				} else {
					return err
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
