package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
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
	// and holds the kernel item indexes of the state the goto is coming from. This is definition 3.16 of IELR(1) and
	// named "follow_kernel_items" there.
	followKernelItems []utils.Bitset
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

// phase1ComputeAuxiliaryTables computes the auxiliary tables of section 3.3 of IELR(1).
//
// Section 3.3 asks for three tables: predecessors, follow_kernel_items and always_follows. Only the first two are
// computed here. The always_follows of definition 3.20 are already computed by the LALR(1) builder, because we follow
// implementation 2 of section 3.3.5, which the paper recommends for a parser generator without an existing LALR(1)
// implementation: the always follows are computed between the two steps of phase 0 and phase 0 derives its goto follows
// from them with definition 3.24. That saves a closure computation and means successor follows are never computed at
// all.
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

// initFollowKernelItems initializes followKernelItems as specified in definition 3.16 of IELR(1).
//
// A goto follow set depends on the lookahead set of a kernel item of the state the goto is coming from, when the kernel
// item has its position in front of the nonterminal of that goto and the rest of the production after that nonterminal
// can be empty. Definition 3.16 does not stop at the goto itself, but follows the goto follows internal relation: the
// follow set of a goto contains the follow sets of all the gotos it depends on internally, so it also depends on the
// kernel items those gotos depend on.
//
// We therefore seed every goto with the kernel items the goto itself depends on, and let the digraph algorithm add the
// kernel items of the gotos which are reachable through the internal relation. This gives us the reflexive transitive
// closure of the internal relation which the definition asks for. Kernel item indexes are only meaningful within a
// single state, but the internal relation only ever relates gotos which come from the same state, so the propagation
// never mixes kernel item indexes of different states.
func (i *IELR1) initFollowKernelItems() {
	i.followKernelItems = make([]utils.Bitset, len(i.lalr1Builder.gotoRecords))
	for gotoIdx, gotoRecord := range i.lalr1Builder.gotoRecords {
		state := i.parser.States[gotoRecord.FromStateIdx]
		for kernelItemIdx, kernelItem := range state.KernelItems.All() {
			production := i.grammar.Productions[kernelItem.ProductionIdx()]
			if kernelItem.Position() == len(production.SymbolRefs) {
				// The kernel item is at the end of the production, so it does not take the goto.
				continue
			}
			symbolRef := production.SymbolRefs[kernelItem.Position()]
			if symbolRef.IsTerminal() || symbolRef.Idx() != gotoRecord.NonterminalIdx {
				// The kernel item does not move over the nonterminal of the goto, so it does not take the goto.
				continue
			}
			if !i.lalr1Builder.isCoreTailEmpty(backend.NewCore(kernelItem.ProductionIdx(), kernelItem.Position()+1)) {
				// The rest of the production after the nonterminal transition can not be empty, so the lookahead set of
				// the kernel item can never follow the nonterminal of the goto.
				continue
			}
			i.followKernelItems[gotoIdx].Add(kernelItemIdx)
		}
	}
	propagation := NewDigraphAlgorithm(i.followKernelItems, i.lalr1Builder.gotoFollowsInternalRelation)
	propagation.Execute()
}

// FollowKernelItems returns the kernel items whose lookahead sets a goto follow set depends on, indexed by goto index.
// The kernel item indexes are indexes into the kernel items of the state the goto is coming from. This is definition
// 3.16 of IELR(1) and named "follow_kernel_items" there.
//
// The table is only valid after phase 1 has run.
func (i *IELR1) FollowKernelItems() []utils.Bitset {
	return i.followKernelItems
}

// GotoRecords returns the details about every nonterminal transition of the parser, indexed by goto index. This is what
// the goto indexes of FollowKernelItems refer to.
//
// The table is only valid after phase 0 has run.
func (i *IELR1) GotoRecords() []GotoRecord {
	return i.lalr1Builder.gotoRecords
}

// Predecessors returns the state indexes of the states which have a transition into the state, indexed by state index.
// This is definition 3.15 of IELR(1) and named "predecessors" there.
//
// The table is only valid after phase 1 has run.
func (i *IELR1) Predecessors() [][]int {
	return i.predecessorStateIdxsByStateIdx
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
