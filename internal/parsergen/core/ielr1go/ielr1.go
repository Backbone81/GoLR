package ielr1go

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarToParser calculates a parser from the context free grammar.
func GrammarToParser(augmentedGrammar frontend.Grammar) (backend.Parser, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Cores: IELR1: GrammarToParser").End()

	builder := NewIELR1(augmentedGrammar)
	return builder.BuildParser(), nil
}

// IELR1 provides an implementation of the IELR(1) algorithm as described by Denny and Malloy in "The IELR(1) algorithm
// for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" at
// https://doi.org/10.1016/j.scico.2009.08.001.
type IELR1 struct {
	grammar frontend.Grammar
	parser  backend.Parser

	// TODO: We should not store the LALR(1) builder. Instead we should copy over what we need and be done with it.
	lalr1Builder LALR1Builder

	// predecessorStateIdxsByStateIdx provides the state indexes for the states which are predecessors when accessed by
	// a state index. This is definition 3.15 of IELR(1) and simply named "predecessors" there.
	// TODO: This table might be valuable to calculate and use during LALR(1) construction already.
	predecessorStateIdxsByStateIdx [][]int

	// followKernelItems reports if the goto does depend on the kernel item's lookahead set. It is indexed by goto index
	// and then kernel item index. This is definition 3.16 of IELR(1) and named "follow_kernel_items" there.
	followKernelItems [][]bool
}

func NewIELR1(augmentedGrammar frontend.Grammar) IELR1 {
	result := IELR1{
		grammar: augmentedGrammar,
		parser: backend.Parser{
			Grammar: augmentedGrammar,
		},
	}
	return result
}

func (i *IELR1) BuildParser() backend.Parser {
	defer trace.StartRegion(context.TODO(), "IELR(1): BuildParser").End()

	i.phase0ComputeLALR1ParserTables()
	i.phase1ComputeAuxiliaryTables()
	i.phase2ComputeAnnotations()
	i.phase3SplitStates()
	i.phase4ComputeReductionLookaheads()
	i.phase5ResolveRemainingConflicts()
	return i.parser
}

func (i *IELR1) Parser() backend.Parser {
	return i.parser
}

func (i *IELR1) phase0ComputeLALR1ParserTables() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 0: LALR(1)").End()
	i.lalr1Builder = NewLALR1Builder(i.grammar)
	i.lalr1Builder.Build()
	i.parser = i.lalr1Builder.Parser()
}

func (i *IELR1) phase1ComputeAuxiliaryTables() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 1: Compute auxiliary tables").End()

	i.initPredecessorStateIdxsByStateIdx()
	i.initFollowKernelItems()
}

// initPredecessorStateIdxsByStateIdx initializes predecessorStateIdxsByStateIdx.
func (i *IELR1) initPredecessorStateIdxsByStateIdx() {
	i.predecessorStateIdxsByStateIdx = make([][]int, len(i.parser.States))
	for stateIdx := range i.parser.States {
		state := i.parser.States[stateIdx]
		for _, transition := range state.TransitionActions.All() {
			i.predecessorStateIdxsByStateIdx[transition.StateIdx()] = append(
				i.predecessorStateIdxsByStateIdx[transition.StateIdx()],
				stateIdx,
			)
		}
	}
}

// initFollowKernelItems initializes followKernelItems.
func (i *IELR1) initFollowKernelItems() {
	i.followKernelItems = make([][]bool, len(i.lalr1Builder.gotoRecords))
	for _, edge := range i.lalr1Builder.gotoFollowsInternalRelation {
		// We need to work with the state which the internal relation is contained in. Make sure that you are using the
		// correct goto and the correct state to stay inside the desired state.
		toGoto := i.lalr1Builder.gotoRecords[edge.ToIdx]
		state := i.parser.States[toGoto.FromStateIdx]
		for kernelItemIdx, kernelItem := range state.KernelItems.All() {
			production := i.grammar.Productions[kernelItem.ProductionIdx()]
			if kernelItem.Position() == len(production.SymbolRefs) ||
				production.SymbolRefs[kernelItem.Position()].IsTerminal() ||
				production.SymbolRefs[kernelItem.Position()].Idx() != toGoto.NonterminalIdx {
				// We are looking for the kernel item which is responsible for the goto. If the kernel item is at the
				// end of the production, or the next symbol is not a nonterminal or the nonterminal is different from
				// the goto, this is not the kernel item we are looking for.
				continue
			}
			if !i.lalr1Builder.isCoreTailEmpty(backend.NewCore(kernelItem.ProductionIdx(), kernelItem.Position()+1)) {
				// The kernel item needs to be empty after the nonterminal transition. If this is not the case, we
				// do not record the dependency.
				continue
			}

			// The goto does depend on the kernel item's lookahead set. Record it for the source and the destination.
			if i.followKernelItems[edge.FromIdx] == nil {
				i.followKernelItems[edge.FromIdx] = make([]bool, state.KernelItems.Length())
			}
			i.followKernelItems[edge.FromIdx][kernelItemIdx] = true
			if i.followKernelItems[edge.ToIdx] == nil {
				i.followKernelItems[edge.ToIdx] = make([]bool, state.KernelItems.Length())
			}
			i.followKernelItems[edge.ToIdx][kernelItemIdx] = true
		}
	}
}

func (i *IELR1) phase2ComputeAnnotations() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 2: Compute annotations").End()

	annotationsBuilder := NewAnnotationsBuilder(
		i.lalr1Builder,
		i.parser,
		i.predecessorStateIdxsByStateIdx,
		i.followKernelItems,
	)
	annotationsBuilder.Execute()
}

func (i *IELR1) phase3SplitStates() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 3: Split states").End()

	// TODO: split states
}

func (i *IELR1) phase4ComputeReductionLookaheads() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 4: Compute reduction lookaheads").End()

	// TODO: compute reduction lookaheads
}

func (i *IELR1) phase5ResolveRemainingConflicts() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 5: Resolve remaining conflicts").End()

	// TODO: resolve remaining conflicts
}
