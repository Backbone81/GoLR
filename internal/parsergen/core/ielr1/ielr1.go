package ielr1

import (
	"context"
	"fmt"
	"runtime/trace"
	"strconv"

	"golr/internal/parsergen/backend"
	"golr/internal/parsergen/frontend"
	"golr/internal/utils/bison"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(augmentedGrammar frontend.Grammar) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: IELR1: GrammarToParser").End()

	builder := NewIELR1(augmentedGrammar)
	return builder.BuildParser()
}

type IELR1 struct {
	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int
}

func NewIELR1(augmentedGrammar frontend.Grammar) *IELR1 {
	return &IELR1{
		terminalIdxByName:    make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
	}
}

func (i *IELR1) BuildParser() (backend.Parser, error) {
	// TODO: Output bison grammar file to temporary file.

	// TODO: Run bison against generated grammar file and produce XML report.

	if err := bison.BuildIELR1("examples/bison/spec/bison-3.8.2.y", "tmp/bison-3.8.2.xml"); err != nil {
		return backend.Parser{}, err
	}

	report, err := bison.LoadBisonXMLReportFromFile("tmp/bison-3.8.2.xml")
	if err != nil {
		return backend.Parser{}, err
	}

	var parser backend.Parser
	i.buildTerminalList(report, &parser)
	i.buildNonterminalList(report, &parser)
	i.buildProductionList(report, &parser)
	if err := i.buildStateList(report, &parser); err != nil {
		return backend.Parser{}, err
	}
	return parser, nil
}

func (i *IELR1) buildTerminalList(report bison.BisonXMLReport, parser *backend.Parser) {
	for _, terminal := range report.Grammar.Terminals {
		i.terminalIdxByName[terminal.Name] = len(parser.Grammar.Terminals)
		parser.Grammar.Terminals = append(parser.Grammar.Terminals, frontend.Symbol{
			Name: terminal.Name,
		})
	}
}

func (i *IELR1) buildNonterminalList(report bison.BisonXMLReport, parser *backend.Parser) {
	for _, nonterminal := range report.Grammar.Nonterminals {
		i.nonterminalIdxByName[nonterminal.Name] = len(parser.Grammar.Nonterminals)
		parser.Grammar.Nonterminals = append(parser.Grammar.Nonterminals, frontend.Symbol{
			Name: nonterminal.Name,
		})
	}
}

func (i *IELR1) buildProductionList(report bison.BisonXMLReport, parser *backend.Parser) {
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

func (i *IELR1) buildStateList(report bison.BisonXMLReport, parser *backend.Parser) error {
	for _, state := range report.Automaton.States {
		var newState backend.State

		for _, item := range state.ItemSet {
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

		for _, reduction := range state.Reductions {
			productionIdx, err := strconv.Atoi(reduction.Rule)
			if err != nil {
				if reduction.Rule == "accept" {
					// The accept rule is always the first production
					productionIdx = 0
				} else {
					return err
				}
			}

			var lookaheadSet backend.LookaheadSet
			if reduction.Symbol == "$default" {
				newState.DefaultReduceProductionIdx = &productionIdx
				// The default reduce action should not show up as a standard reduce. Therefore skip to the next.
				continue
			} else {
				terminalIdx, ok := i.terminalIdxByName[reduction.Symbol]
				if !ok {
					return fmt.Errorf("unknown terminal %q", reduction.Symbol)
				}
				lookaheadSet.Add(terminalIdx)
			}

			newState.ReduceActions.Add(backend.NewReduceAction(lookaheadSet, productionIdx))
		}
		parser.States = append(parser.States, newState)
	}
	return nil
}
