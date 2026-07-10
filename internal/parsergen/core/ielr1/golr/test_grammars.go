package golr

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var (

	// UnambiguousTestGrammarFig1 is the unambiguous grammar from the IELR(1) paper in Fig. 1 on page 3 (or 945).
	//
	//   1. S -> aAa
	//   2. S -> bAb
	//   3. A -> a
	//   4. A -> aa
	//
	UnambiguousTestGrammarFig1 = frontend.Grammar{
		Terminals: []frontend.Symbol{
			{
				Name: "a", // 0
			},
			{
				Name: "b", // 1
			},
		},
		Nonterminals: []frontend.Symbol{
			{
				Name: "S", // 0
			},
			{
				Name: "A", // 1
			},
		},
		Productions: []frontend.Production{
			// 1. S -> aAa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(0),    // a
				},
			},
			// 2. S -> bAb
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(1),    // b
				},
			},
			// 3. A -> a
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
				},
			},
			// 4. A -> aa
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(0), // a
				},
			},
		},
		StartNonterminalIdx: 0, // "S"
	}
	UnambiguousTestGrammarFig1Augmented   = frontend.AugmentGrammar(UnambiguousTestGrammarFig1)
	UnambiguousTestGrammarFig1LALR1Parser = backend.Parser{
		Grammar: UnambiguousTestGrammarFig1Augmented,
		// NOTE: The order of states is the same order as Bison would create them. The IELR(1) paper depicts the
		// states in a different order. We note down the state index from the paper for ease of reference.
		States: []backend.State{
			// state 0: Table 1: State 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),    // S3 on a
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),    // S4 on b
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3), // G1 on S
				),
			},
			// state 1: Table 1: State 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),    // S9 on a
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5), // G5 on A
				),
			},
			// state 2: Table 1: State 4
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),    // S9 on a
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 6), // G6 on A
				),
			},
			// state 3: Table 1: State 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 7), // S2 on EOF
				),
			},
			// state 4: Table 1: State 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 1),
					backend.NewCore(4, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 8), // S10 on a
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 3),
				),
			},
			// state 5: Table 1: State 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 9), // S7 on a
				),
			},
			// state 6: Table 1: State 6
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 10), // S8 on b
				),
			},
			// state 7: Table 1: State 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 8: Table 1: State 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 4),
				),
			},
			// state 9: Table 1: State 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
				),
			},
			// state 10: Table 1: State 8
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2),
				),
			},
		},
	}
	UnambiguousTestGrammarFig1LR1Parser = backend.Parser{
		Grammar: UnambiguousTestGrammarFig1Augmented,
		States: []backend.State{
			// state 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),    // shift a
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),    // shift b
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3), // goto S
				),
			},
			// state 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),    // shift a
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5), // goto A
				),
			},
			// state 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 6),    // shift a
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 7), // goto A
				),
			},
			// state 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 8), // shift EOF
				),
			},
			// state 4: reached through "a", so "A -> a ." reduces on "a" only
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 1),
					backend.NewCore(4, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 9), // shift a
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1), 3), // A -> a on {a}
				),
			},
			// state 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 10), // shift a
				),
			},
			// state 6: reached through "b", so "A -> a ." reduces on "b" only
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 1),
					backend.NewCore(4, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 11), // shift a
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(2), 3), // A -> a on {b}
				),
			},
			// state 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 12), // shift b
				),
			},
			// state 8: the accepting state. Nothing can follow the end of input, so the lookahead set is empty.
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1), 4), // A -> a on {a}
				),
			},
			// state 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1), // S -> a A a on {EOF}
				),
			},
			// state 11
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(2), 4), // A -> a on {b}
				),
			},
			// state 12
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2), // S -> b A b on {EOF}
				),
			},
		},
	}

	// AmbiguousTestGrammarFig2 is the ambiguous grammar from the IELR(1) paper in Fig. 2 on page 3 (or 945).
	//
	//   1. S -> aAa
	//   2. S -> aBb
	//   3. S -> aCc
	//   4. S -> bAb
	//   5. S -> bBa
	//   6. S -> bCa
	//   7. A -> aa
	//   8. B -> aa
	//   9. C -> aa
	//
	AmbiguousTestGrammarFig2 = frontend.Grammar{
		Terminals: []frontend.Symbol{
			{
				Name: "a", // 0
			},
			{
				Name: "b", // 1
			},
			{
				Name: "c", // 2
			},
		},
		Nonterminals: []frontend.Symbol{
			{
				Name: "S", // 0
			},
			{
				Name: "A", // 1
			},
			{
				Name: "B", // 2
			},
			{
				Name: "C", // 3
			},
		},
		Productions: []frontend.Production{
			// 1. S -> aAa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(0),    // a
				},
			},
			// 2. S -> aBb
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(1),    // b
				},
			},
			// 3. S -> aCc
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(3), // C
					frontend.NewTerminalRef(2),    // c
				},
			},
			// 4. S -> bAb
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(1),    // b
				},
			},
			// 5. S -> bBa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(0),    // a
				},
			},
			// 6. S -> bCa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(3), // C
					frontend.NewTerminalRef(0),    // a
				},
			},
			// 7. A -> aa
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(0), // a
				},
			},
			// 8. B -> aa
			{
				NonterminalIdx: 2, // B
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(0), // a
				},
			},
			// 9. C -> aa
			{
				NonterminalIdx: 3, // C
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(0), // a
				},
			},
		},
		StartNonterminalIdx: 0, // S
	}
	AmbiguousTestGrammarFig2Augmented   = frontend.AugmentGrammar(AmbiguousTestGrammarFig2)
	AmbiguousTestGrammarFig2LALR1Parser = backend.Parser{
		Grammar: AmbiguousTestGrammarFig2Augmented,
		// NOTE: The order of states is the same order as Bison would create them. The IELR(1) paper depicts the
		// states in a different order. We note down the state index from the paper for ease of reference.
		States: []backend.State{
			// state 0: Table 3: State 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3),
				),
			},
			// state 1: Table 3: State 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
					backend.NewCore(2, 1),
					backend.NewCore(3, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 6),
					backend.NewTransitionAction(frontend.NewNonterminalRef(4), 7),
				),
			},
			// state 2: Table 3: State 4
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 1),
					backend.NewCore(5, 1),
					backend.NewCore(6, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 8),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 9),
					backend.NewTransitionAction(frontend.NewNonterminalRef(4), 10),
				),
			},
			// state 3: Table 3: State 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 11),
				),
			},
			// state 4: Table 3: State 17
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(7, 1),
					backend.NewCore(8, 1),
					backend.NewCore(9, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 12),
				),
			},
			// state 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 13),
				),
			},
			// state 6
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 14),
				),
			},
			// state 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(3), 15),
				),
			},
			// state 8
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 16),
				),
			},
			// state 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(5, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 17),
				),
			},
			// state 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(6, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 18),
				),
			},
			// state 11: Table 3: State 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 12: Table 3: State 18
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(7, 2),
					backend.NewCore(8, 2),
					backend.NewCore(9, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 7),
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 8),
					backend.NewReduceAction(backend.NewLookaheadSet(1, 3), 9),
				),
			},
			// state 13
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
				),
			},
			// state 14
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2),
				),
			},
			// state 15
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 3),
				),
			},
			// state 16
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 4),
				),
			},
			// state 17
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(5, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 5),
				),
			},
			// state 18
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(6, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 6),
				),
			},
		},
	}

	// GotoFollowsTestGrammarFig5 is the grammar demonstrating goto follows from the IELR(1) paper in Fig. 5 on page 13
	// (or 955).
	//
	//   1. S -> aABa
	//   2. S -> bABb
	//   3. A -> aCDE
	//   4. B -> c
	//   5. B ->
	//   6. C -> D
	//   7. D -> a
	//   8. E -> a
	//   9. E ->
	//
	GotoFollowsTestGrammarFig5 = frontend.Grammar{
		Terminals: []frontend.Symbol{
			{
				Name: "a", // 0
			},
			{
				Name: "b", // 1
			},
			{
				Name: "c", // 2
			},
		},
		Nonterminals: []frontend.Symbol{
			{
				Name: "S", // 0
			},
			{
				Name: "A", // 1
			},
			{
				Name: "B", // 2
			},
			{
				Name: "C", // 3
			},
			{
				Name: "D", // 4
			},
			{
				Name: "E", // 5
			},
		},
		Productions: []frontend.Production{
			//   1. S -> aABa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(1), // A
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(0),    // a
				},
			},
			//   2. S -> bABb
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(1), // A
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(1),    // b
				},
			},
			//   3. A -> aCDE
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(3), // C
					frontend.NewNonterminalRef(4), // D
					frontend.NewNonterminalRef(5), // E
				},
			},
			//   4. B -> c
			{
				NonterminalIdx: 2, // B
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(2), // c
				},
			},
			//   5. B ->
			{
				NonterminalIdx: 2, // B
			},
			//   6. C -> D
			{
				NonterminalIdx: 3, // C
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewNonterminalRef(4), // D
				},
			},
			//   7. D -> a
			{
				NonterminalIdx: 4, // D
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
				},
			},
			//   8. E -> a
			{
				NonterminalIdx: 5, // E
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
				},
			},
			//   9. E ->
			{
				NonterminalIdx: 5, // E
			},
		},
		StartNonterminalIdx: 0, // S
	}
	GotoFollowsTestGrammarFig5Augmented   = frontend.AugmentGrammar(GotoFollowsTestGrammarFig5)
	GotoFollowsTestGrammarFig5LALR1Parser = backend.Parser{
		Grammar: GotoFollowsTestGrammarFig5Augmented,
		// NOTE: The order of states is the same order as Bison would create them. The IELR(1) paper depicts the
		// states in a different order. We note down the state index from the paper for ease of reference.
		States: []backend.State{
			// state 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3),
				),
			},
			// state 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5),
				),
			},
			// state 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 6),
				),
			},
			// state 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 7),
				),
			},
			// state 4
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 8),
					backend.NewTransitionAction(frontend.NewNonterminalRef(4), 9),
					backend.NewTransitionAction(frontend.NewNonterminalRef(5), 10),
				),
			},
			// state 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(3), 11),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 12),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1), 5),
				),
			},
			// state 6
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(3), 11),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 13),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(2), 5),
				),
			},
			// state 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 8
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(7, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2, 3), 7),
				),
			},
			// state 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 8),
					backend.NewTransitionAction(frontend.NewNonterminalRef(5), 14),
				),
			},
			// state 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(6, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1), 6),
				),
			},
			// state 11
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 4),
				),
			},
			// state 12
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 15),
				),
			},
			// state 13
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 16),
				),
			},
			// state 14
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 3),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 17),
					backend.NewTransitionAction(frontend.NewNonterminalRef(6), 18),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2, 3), 9),
				),
			},
			// state 15
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 4),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
				),
			},
			// state 16
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 4),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2),
				),
			},
			// state 17
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(8, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2, 3), 8),
				),
			},
			// state 18
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 4),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2, 3), 3),
				),
			},
		},
	}

	// GotoFollowsCaveatsTestGrammarFig6 is the grammar demonstrating goto follows caveats from the IELR(1) paper in
	// Fig. 6 on page 17 (or 959).
	//
	//   1. S -> aAa
	//   2. S -> aab
	//   3. S -> bAb
	//   4. A -> BC
	//   5. B -> a
	//   6. C -> D
	//   7. D ->
	//
	GotoFollowsCaveatsTestGrammarFig6 = frontend.Grammar{
		Terminals: []frontend.Symbol{
			{
				Name: "a", // 0
			},
			{
				Name: "b", // 1
			},
		},
		Nonterminals: []frontend.Symbol{
			{
				Name: "S", // 0
			},
			{
				Name: "A", // 1
			},
			{
				Name: "B", // 2
			},
			{
				Name: "C", // 3
			},
			{
				Name: "D", // 4
			},
		},
		Productions: []frontend.Production{
			//   1. S -> aAa
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(0),    // a
				},
			},
			//   2. S -> aab
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(0), // a
					frontend.NewTerminalRef(1), // b
				},
			},
			//   3. S -> bAb
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(1),    // b
				},
			},
			//   4. A -> BC
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewNonterminalRef(2), // B
					frontend.NewNonterminalRef(3), // C
				},
			},
			//   5. B -> a
			{
				NonterminalIdx: 2, // B
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0), // a
				},
			},
			//   6. C -> D
			{
				NonterminalIdx: 3, // C
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewNonterminalRef(4), // D
				},
			},
			//   7. D ->
			{
				NonterminalIdx: 4, // D
			},
		},
		StartNonterminalIdx: 0, // S
	}
	GotoFollowsCaveatsTestGrammarFig6Augmented  = frontend.AugmentGrammar(GotoFollowsCaveatsTestGrammarFig6)
	GotoFollowsCaveatsTestGrammarFig6LALRParser = backend.Parser{
		Grammar: GotoFollowsCaveatsTestGrammarFig6Augmented,
		// NOTE: The order of states is the same order as Bison would create them. The IELR(1) paper depicts the
		// states in a different order. We note down the state index from the paper for ease of reference.
		States: []backend.State{
			// state 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3),
				),
			},
			// state 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
					backend.NewCore(2, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 4),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 6),
				),
			},
			// state 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 7),
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 8),
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 6),
				),
			},
			// state 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 9),
				),
			},
			// state 4
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
					backend.NewCore(5, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 10),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1), 5),
				),
			},
			// state 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 11),
				),
			},
			// state 6
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewNonterminalRef(4), 12),
					backend.NewTransitionAction(frontend.NewNonterminalRef(5), 13),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 7),
				),
			},
			// state 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(5, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(2), 5),
				),
			},
			// state 8
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 14),
				),
			},
			// state 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2),
				),
			},
			// state 11
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
				),
			},
			// state 12
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 4),
				),
			},
			// state 13
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(6, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(1, 2), 6),
				),
			},
			// state 14
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 3),
				),
			},
		},
	}

	// ReduceReduceConflictTestGrammar is the classic grammar which is LR(1) but not LALR(1). It is used to verify that
	// the LALR(1) builder faithfully produces overlapping reduction lookahead sets when the grammar has a genuine
	// LALR(1) conflict (the builder does not resolve conflicts; that is a later phase). In canonical LR(1) the two "c"
	// states stay separate, but LALR(1) merges them because they share the same core, which merges their lookahead sets
	// and creates a reduce/reduce conflict on both "d" and "e".
	//
	//   1. S -> aAd
	//   2. S -> bBd
	//   3. S -> aBe
	//   4. S -> bAe
	//   5. A -> c
	//   6. B -> c
	//
	// This behavior was cross-checked against GNU Bison 3.8.2 (--define=lr.type=lalr), which reports the same two
	// reduce/reduce conflicts in the merged "c" state.
	ReduceReduceConflictTestGrammar = frontend.Grammar{
		Terminals: []frontend.Symbol{
			{
				Name: "a", // 0
			},
			{
				Name: "b", // 1
			},
			{
				Name: "c", // 2
			},
			{
				Name: "d", // 3
			},
			{
				Name: "e", // 4
			},
		},
		Nonterminals: []frontend.Symbol{
			{
				Name: "S", // 0
			},
			{
				Name: "A", // 1
			},
			{
				Name: "B", // 2
			},
		},
		Productions: []frontend.Production{
			//   1. S -> aAd
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(3),    // d
				},
			},
			//   2. S -> bBd
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(3),    // d
				},
			},
			//   3. S -> aBe
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(0),    // a
					frontend.NewNonterminalRef(2), // B
					frontend.NewTerminalRef(4),    // e
				},
			},
			//   4. S -> bAe
			{
				NonterminalIdx: 0, // S
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(1),    // b
					frontend.NewNonterminalRef(1), // A
					frontend.NewTerminalRef(4),    // e
				},
			},
			//   5. A -> c
			{
				NonterminalIdx: 1, // A
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(2), // c
				},
			},
			//   6. B -> c
			{
				NonterminalIdx: 2, // B
				SymbolRefs: []frontend.SymbolRef{
					frontend.NewTerminalRef(2), // c
				},
			},
		},
		StartNonterminalIdx: 0, // S
	}
	ReduceReduceConflictTestGrammarAugmented   = frontend.AugmentGrammar(ReduceReduceConflictTestGrammar)
	ReduceReduceConflictTestGrammarLALR1Parser = backend.Parser{
		Grammar: ReduceReduceConflictTestGrammarAugmented,
		// NOTE: The order of states is the same order as Bison would create them.
		States: []backend.State{
			// state 0
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 0),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(1), 1),    // shift a
					backend.NewTransitionAction(frontend.NewTerminalRef(2), 2),    // shift b
					backend.NewTransitionAction(frontend.NewNonterminalRef(1), 3), // goto S
				),
			},
			// state 1
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 1),
					backend.NewCore(3, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(3), 4),    // shift c
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 5), // goto A
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 6), // goto B
				),
			},
			// state 2
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 1),
					backend.NewCore(4, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(3), 4),    // shift c
					backend.NewTransitionAction(frontend.NewNonterminalRef(2), 7), // goto A
					backend.NewTransitionAction(frontend.NewNonterminalRef(3), 8), // goto B
				),
			},
			// state 3
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 1),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(0), 9), // shift EOF
				),
			},
			// state 4: the merged "c" state which carries the reduce/reduce conflict. A -> c . and B -> c . both reduce
			// on the merged lookahead set {d, e}.
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(5, 1),
					backend.NewCore(6, 1),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(4, 5), 5), // A -> c on {d, e}
					backend.NewReduceAction(backend.NewLookaheadSet(4, 5), 6), // B -> c on {d, e}
				),
			},
			// state 5
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(4), 10), // shift d
				),
			},
			// state 6
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(5), 11), // shift e
				),
			},
			// state 7
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(5), 12), // shift e
				),
			},
			// state 8
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 2),
				),
				TransitionActions: backend.NewTransitionActionSet(
					backend.NewTransitionAction(frontend.NewTerminalRef(4), 13), // shift d
				),
			},
			// state 9
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(0, 2),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.LookaheadSet{}, 0),
				),
			},
			// state 10
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(1, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 1),
				),
			},
			// state 11
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(3, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 3),
				),
			},
			// state 12
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(4, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 4),
				),
			},
			// state 13
			{
				KernelItems: backend.NewCoreSet(
					backend.NewCore(2, 3),
				),
				ReduceActions: backend.NewReduceActionSet(
					backend.NewReduceAction(backend.NewLookaheadSet(0), 2),
				),
			},
		},
	}
)
