package ielr1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/trace"
	"strconv"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	bisonutils "github.com/backbone81/golr/internal/utils/bison"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(augmentedGrammar frontend.Grammar) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: IELR1: GrammarToParser").End()

	builder := NewIELR1(augmentedGrammar)
	return builder.BuildParser()
}

type IELR1 struct {
	grammar              frontend.Grammar
	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int
}

func NewIELR1(augmentedGrammar frontend.Grammar) *IELR1 {
	return &IELR1{
		grammar:              augmentedGrammar,
		terminalIdxByName:    make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
	}
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

	i.buildTerminalList(report, &parser)
	i.buildNonterminalList(report, &parser)
	i.buildProductionList(report, &parser)
	if err := i.buildStateList(report, &parser); err != nil {
		return backend.Parser{}, err
	}
	return parser, nil
}

func (i *IELR1) buildTerminalList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
	for _, terminal := range report.Grammar.Terminals {
		i.terminalIdxByName[terminal.Name] = len(parser.Grammar.Terminals)
		parser.Grammar.Terminals = append(parser.Grammar.Terminals, frontend.Symbol{
			Name: terminal.Name,
		})
	}
}

func (i *IELR1) buildNonterminalList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
	for _, nonterminal := range report.Grammar.Nonterminals {
		i.nonterminalIdxByName[nonterminal.Name] = len(parser.Grammar.Nonterminals)
		parser.Grammar.Nonterminals = append(parser.Grammar.Nonterminals, frontend.Symbol{
			Name: nonterminal.Name,
		})
	}
}

func (i *IELR1) buildProductionList(report bisonutils.BisonXMLReport, parser *backend.Parser) {
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

func (i *IELR1) buildStateList(report bisonutils.BisonXMLReport, parser *backend.Parser) error {
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
