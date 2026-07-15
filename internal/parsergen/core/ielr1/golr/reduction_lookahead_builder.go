package golr

import (
	"context"
	"runtime/trace"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ReductionLookaheadBuilder computes the reduction lookahead sets of an LR automaton from its states. It applies the
// algorithm of DeRemer and Pennello in "Efficient Computation of LALR(1) Look-Ahead Sets" at
// https://doi.org/10.1145/69622.357187, which is step 2 of phase 0 in section 3.2.2 of IELR(1) ("Reduction lookaheads
// from goto follows").
//
// The input is the bare automaton: the states with their kernel items, their transition actions and the productions
// they reduce. From those the builder re-derives everything the lookahead computation needs — the goto records of
// definition 3.4, the backward transitions, the goto follows successor/internal/predecessor relations of definitions
// 3.5, 3.8 and 3.9, the always follows of definition 3.20 and the goto follows of definition 3.24 — and finally the
// reduction lookahead sets for the reduce actions.
//
// The reduction lookaheads are needed twice in IELR(1): phase 0 runs the builder on the LR(0) automaton, and phase 4
// (section 3.6, "run step 2 of phase 0 again") runs it on the split automaton phase 3 produced. Re-deriving the
// auxiliary tables from the states rather than from the construction which produced them is what lets the same
// component serve both phases.
//
// The computed lookaheads are exposed through ReduceActions and are not written back into the states, so the input
// states are left untouched.
type ReductionLookaheadBuilder struct {
	// grammar is the augmented context free grammar the automaton was built from.
	grammar frontend.Grammar

	// productionIdxsByNonterminalIdx maps a nonterminal index to a slice of production indexes. This makes it easier to
	// find all productions which have the given nonterminal on the left hand side of the production.
	productionIdxsByNonterminalIdx map[int][]int

	// nullableByNonterminalIdx provides information about a nonterminal index being nullable or not. This is needed
	// for calculating if the rest of some item can be empty or not.
	nullableByNonterminalIdx map[int]bool

	// states is the LR automaton the reduction lookahead sets are computed for. Each state must carry its kernel items,
	// its transition actions and its reduce actions. The reduce actions only need to name the productions which reduce
	// in the state; their lookahead sets are what this builder computes.
	states []backend.State

	// reduceActions is a list with all reduce actions of the automaton, derived from the reduce actions of the states.
	// Their lookahead sets are filled in during Execute.
	reduceActions []ReduceActionRecord

	// gotoRecords provides details about each nonterminal transition. This is derived from definition 3.4 of IELR(1).
	gotoRecords []GotoRecord

	// gotoFollows holds the goto follow set for each goto, indexed by goto index. This is "goto_follows" from IELR(1)
	// definition 3.4.
	gotoFollows []backend.LookaheadSet

	// alwaysFollows holds the follow set from definition 3.20 of IELR(1), indexed by goto index.
	alwaysFollows []backend.LookaheadSet

	// gotoIdxsByStateIdx provides a list of goto indexes when indexed by state index. This is helpful when calculating
	// internal dependencies, as we need access to all gotos within the same state.
	gotoIdxsByStateIdx map[int][]int

	// backwardTransitionsByStateIdx provides information about which transitions lead into the state index.
	backwardTransitionsByStateIdx map[int]BackwardTransitionInfo

	// gotoFollowsSuccessorRelation is the digraph describing the successor dependencies as GFs(g, g') from IELR(1)
	// definition 3.5.
	gotoFollowsSuccessorRelation []Edge

	// gotoFollowsInternalRelation is the digraph describing the internal dependencies as GFi(g, g') from IELR(1)
	// definition 3.8. This relation is needed in later stages of IELR(1), therefore save it here separately.
	gotoFollowsInternalRelation []Edge

	// gotoFollowsPredecessorRelation is the digraph describing the predecessor dependencies as GFp(g, g') from IELR(1)
	// definition 3.9.
	gotoFollowsPredecessorRelation []Edge

	// successorDependencyCandidates is a list which holds the goto indexes of gotos which happen on nullable
	// nonterminals. Those gotos are the destinations for the goto follows successor relations. The relations can then
	// be built from this list.
	successorDependencyCandidates []int

	// internalDependencyCandidates provides a list of goto indexes which are part of an internal dependency. This list
	// is derived from the states and used afterward to build the goto follows internal relations.
	internalDependencyCandidates []InternalDependencyCandidate

	// predecessorDependencyCandidates provides a list of goto indexes which are part of a predecessor dependency. This
	// list is derived from the states and used afterward to build the goto follows predecessor relations.
	predecessorDependencyCandidates []PredecessorDependencyCandidate
}

// NewReductionLookaheadBuilder returns a new builder which computes the reduction lookahead sets of the given
// automaton. The grammar provided MUST be the augmented grammar the states were built from.
func NewReductionLookaheadBuilder(grammar frontend.Grammar, states []backend.State) ReductionLookaheadBuilder {
	return ReductionLookaheadBuilder{
		grammar:                        grammar,
		states:                         states,
		productionIdxsByNonterminalIdx: make(map[int][]int, 128),
		nullableByNonterminalIdx:       make(map[int]bool, 128),

		gotoIdxsByStateIdx: make(map[int][]int, 256),

		backwardTransitionsByStateIdx: make(map[int]BackwardTransitionInfo),
	}
}

// Build computes the reduction lookahead sets. You can retrieve them with a call to ReduceActions afterward. The results
// are only valid after Build has run, and it must run exactly once.
func (b *ReductionLookaheadBuilder) Build() {
	defer trace.StartRegion(context.TODO(), "Add reduction lookahead sets").End()

	b.initProductionIdxsByNonterminalIdx()
	b.initNullableByNonterminalIdx()
	b.deriveAutomatonTables()

	b.buildGotoFollowsSuccessorRelations()
	b.buildGotoFollowsInternalRelations()
	b.buildGotoFollowsPredecessorRelations()

	// We follow implementation 2 from IELR(1) section 3.3.5, which computes goto follows from always follows
	// (definition 3.24) and never computes successor follows (definition 3.6). Only the successor relation itself is
	// needed, as an input to the always follows.
	//
	// TODO: Check if we can improve performance by not calculating all always and goto follows up front, but instead
	// lazily calculate those which we need for reduce actions. This could result in a significant amount of follow
	// sets not being calculated as they are not involved in any reduce action.
	b.calculateAlwaysFollows()
	b.calculateGotoFollows()

	b.calculateReduceActionLookaheads()
}

// ReduceActions returns the reduce actions of the automaton together with their computed reduction lookahead sets. The
// result is only valid after Build has run.
func (b *ReductionLookaheadBuilder) ReduceActions() []ReduceActionRecord {
	return b.reduceActions
}

// GotoRecords returns the details about every nonterminal transition of the automaton, indexed by goto index. This is
// derived from definition 3.4 of IELR(1). The result is only valid after Build has run.
func (b *ReductionLookaheadBuilder) GotoRecords() []GotoRecord {
	return b.gotoRecords
}

// GotoIdxsByStateIdx returns the goto indexes of the gotos which come from a state, keyed by state index. The result is
// only valid after Build has run.
func (b *ReductionLookaheadBuilder) GotoIdxsByStateIdx() map[int][]int {
	return b.gotoIdxsByStateIdx
}

// GotoFollows returns the goto follow set of every goto, indexed by goto index. This is "goto_follows" from definition
// 3.4 of IELR(1). The result is only valid after Build has run.
func (b *ReductionLookaheadBuilder) GotoFollows() []backend.LookaheadSet {
	return b.gotoFollows
}

// AlwaysFollows returns the terminals which follow a goto no matter what the lookahead sets of the kernel items of the
// state the goto comes from are, indexed by goto index. This is definition 3.20 of IELR(1) and named "always_follows"
// there. The result is only valid after Build has run.
func (b *ReductionLookaheadBuilder) AlwaysFollows() []backend.LookaheadSet {
	return b.alwaysFollows
}

// GotoFollowsInternalRelation returns the digraph describing the internal dependencies as GFi(g, g') from IELR(1)
// definition 3.8. Later phases of IELR(1) need this relation, for example to propagate the follow kernel items of
// definition 3.16. The result is only valid after Build has run.
func (b *ReductionLookaheadBuilder) GotoFollowsInternalRelation() []Edge {
	return b.gotoFollowsInternalRelation
}

// IsCoreTailEmpty reports if the position within the production is at the end of the production or the symbols following
// the current position are all nullable. This is only valid after Build has run, because it depends on the nullable
// nonterminals computed there.
func (b *ReductionLookaheadBuilder) IsCoreTailEmpty(core backend.Core) bool {
	return b.isCoreTailEmpty(core)
}

// initProductionIdxsByNonterminalIdx initializes the helper variable productionIdxsByNonterminalIdx.
func (b *ReductionLookaheadBuilder) initProductionIdxsByNonterminalIdx() {
	for idx, production := range b.grammar.Productions {
		b.productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			b.productionIdxsByNonterminalIdx[production.NonterminalIdx],
			idx,
		)
	}
}

// initNullableByNonterminalIdx initializes the helper variable nullableByNonterminalIdx. It is doing a fixed-point
// computation to find all the nullable nonterminals by inspecting the productions and checking for directly empty
// right hand sides of the productions or by indirectly empty right hand sides.
func (b *ReductionLookaheadBuilder) initNullableByNonterminalIdx() {
	changed := true
	for changed {
		changed = false
		for nonterminalIdx, productionIdxs := range b.productionIdxsByNonterminalIdx {
			if b.nullableByNonterminalIdx[nonterminalIdx] {
				// We already know that this nonterminal is nullable, so we do not need to check all productions
				// for that nonterminal again.
				continue
			}
			if slices.ContainsFunc(productionIdxs, b.isProductionNullable) {
				// As the right hand side of the production can be empty, we know that the nonterminal on the
				// left hand side of the production is nullable.
				b.nullableByNonterminalIdx[nonterminalIdx] = true
				changed = true
			}
		}
	}
}

// isProductionNullable reports if the right hand side of the production is empty or the right hand side consists
// only of nonterminals which are nullable themselves.
func (b *ReductionLookaheadBuilder) isProductionNullable(productionIdx int) bool {
	return b.isCoreTailEmpty(backend.NewCore(productionIdx, 0))
}

// isCoreTailEmpty reports if the position within the production is at the end of the production or the symbols for the
// following the current position are all nullable.
func (b *ReductionLookaheadBuilder) isCoreTailEmpty(core backend.Core) bool {
	production := b.grammar.Productions[core.ProductionIdx()]

	if core.Position() == len(production.SymbolRefs) {
		// The item is already at the end of the production. The tail is therefore empty.
		return true
	}

	for _, symbolRef := range production.SymbolRefs[core.Position():] {
		if symbolRef.IsTerminal() {
			// The symbol is a terminal which means the tail can not be empty.
			return false
		}
		if !b.nullableByNonterminalIdx[symbolRef.Idx()] {
			// The symbol is a nonterminal which is not nullable which means the tail can not be empty.
			return false
		}
	}
	// All remaining symbols were nonterminals and each nonterminal was nullable. Therefore, the core tail is empty.
	return true
}

// deriveAutomatonTables reconstructs the reduce action records, the goto records, the backward transitions and the goto
// follows dependency candidates from the states. These tables come for free while an LR(0) automaton is constructed,
// but this builder also runs on the split automaton of phase 3, where that construction is not available, so it
// re-derives them from the reduce actions, the transition actions and the kernel items of the states.
func (b *ReductionLookaheadBuilder) deriveAutomatonTables() {
	for stateIdx := range b.states {
		state := &b.states[stateIdx]

		// A reduce action of a state only names the production which reduces there. Its core sits at the end of the
		// production, which is where an item reduces, and it is that core we later trace back to the gotos which
		// generated it. For an empty production the length is zero, so the core is at position zero, matching a start
		// item as it must.
		for _, reduceAction := range state.ReduceActions.All() {
			production := b.grammar.Productions[reduceAction.ProductionIdx]
			b.reduceActions = append(b.reduceActions, ReduceActionRecord{
				StateIdx: stateIdx,
				Core:     backend.NewCore(reduceAction.ProductionIdx, len(production.SymbolRefs)),
			})
		}

		// The transition actions of a state are the forward edges of the automaton. We invert them into backward
		// transitions and, for the nonterminal transitions, record the goto records and the goto follows dependency
		// candidates.
		for _, transitionAction := range state.TransitionActions.All() {
			symbolRef := transitionAction.SymbolRef()
			toStateIdx := transitionAction.StateIdx()
			if symbolRef.IsNonterminal() {
				b.recordNonterminalTransition(stateIdx, symbolRef.Idx(), toStateIdx)
			} else {
				b.recordTerminalTransition(stateIdx, symbolRef.Idx(), toStateIdx)
			}
		}
	}
}

// recordNonterminalTransition records the goto record, the backward transition and the goto follows dependency
// candidates for a nonterminal transition.
func (b *ReductionLookaheadBuilder) recordNonterminalTransition(fromStateIdx int, nonterminalIdx int, toStateIdx int) {
	// record goto record
	b.gotoRecords = append(b.gotoRecords, GotoRecord{
		FromStateIdx:   fromStateIdx,
		ToStateIdx:     toStateIdx,
		NonterminalIdx: nonterminalIdx,
	})
	gotoIdx := len(b.gotoRecords) - 1
	b.gotoIdxsByStateIdx[fromStateIdx] = append(b.gotoIdxsByStateIdx[fromStateIdx], gotoIdx)

	// record backward transition
	transitions, exist := b.backwardTransitionsByStateIdx[toStateIdx]
	if !exist {
		transitions = NewBackwardTransitionInfo()
	}
	transitions.NonterminalTransitions[nonterminalIdx] = append(
		transitions.NonterminalTransitions[nonterminalIdx],
		fromStateIdx,
	)
	b.backwardTransitionsByStateIdx[toStateIdx] = transitions

	// Record information needed to calculate goto follows later. The kernel items reached by this goto are the kernel
	// items of the destination state: every kernel item of a state is an item advanced over the state's single entry
	// symbol, which for a goto is exactly its nonterminal.
	b.recordSuccessorDependencyCandidate(nonterminalIdx, gotoIdx)
	b.recordIncludesDependencyCandidate(&b.states[toStateIdx].KernelItems, gotoIdx)
}

// recordSuccessorDependencyCandidate checks if the goto is part of a goto follows successor relation as specified by
// definition 3.5 of IELR(1) and records it as candidate for later use.
func (b *ReductionLookaheadBuilder) recordSuccessorDependencyCandidate(nonterminalIdx int, gotoIdx int) {
	// Check if this goto is part of a successor dependency for the goto follows.
	if b.nullableByNonterminalIdx[nonterminalIdx] {
		// We need this information when constructing the goto follows successor relation.
		b.successorDependencyCandidates = append(b.successorDependencyCandidates, gotoIdx)
	}
}

// recordIncludesDependencyCandidate checks if the kernel items are part of a goto follows includes relation as
// specified by definition 3.7 of IELR(1). As goto follows includes relations can be broken down into goto follows
// internal relations as specified by definition 3.8 of IELR(1) and goto follows predecessor relations as specified by
// definition 3.9 of IELR(1), this method checks for both cases and records the goto as a candidate for later use.
func (b *ReductionLookaheadBuilder) recordIncludesDependencyCandidate(nextKernelItems *backend.CoreSet, gotoIdx int) {
	for _, kernelItem := range nextKernelItems.All() {
		if !b.isCoreTailEmpty(kernelItem) {
			// We are not interested in items which are not empty for the rest of the production.
			continue
		}

		if kernelItem.Position() == 1 {
			// We found an item which is at the start of the production and empty after the current position. This
			// is a candidate for an internal dependency which we need to record for later. Note that the item here
			// was already advanced by one symbol, so we need to check for 1 instead of 0.
			production := b.grammar.Productions[kernelItem.ProductionIdx()]
			b.internalDependencyCandidates = append(b.internalDependencyCandidates, InternalDependencyCandidate{
				GotoIdx:        gotoIdx,
				NonterminalIdx: production.NonterminalIdx,
			})
		} else {
			// We found an item which is not at the start of the production and empty after the current position.
			// This is a candidate for a predecessor dependency which we need to record for later.
			b.predecessorDependencyCandidates = append(b.predecessorDependencyCandidates, PredecessorDependencyCandidate{
				GotoIdx: gotoIdx,
				// Note that we need to move back the kernel by one position, because the kernel we have here is already
				// moved forward by one, but we need the core as it was for the state we are transitioning from.
				Core: backend.NewCore(kernelItem.ProductionIdx(), kernelItem.Position()-1),
			})
		}
	}
}

// recordTerminalTransition records the backward transition for a terminal transition.
func (b *ReductionLookaheadBuilder) recordTerminalTransition(fromStateIdx int, terminalIdx int, toStateIdx int) {
	transitions, exist := b.backwardTransitionsByStateIdx[toStateIdx]
	if !exist {
		transitions = NewBackwardTransitionInfo()
	}
	transitions.TerminalTransitions[terminalIdx] = append(transitions.TerminalTransitions[terminalIdx], fromStateIdx)
	b.backwardTransitionsByStateIdx[toStateIdx] = transitions
}

// calculateReduceActionLookaheads fills the reduction lookahead set of every reduce action. We do this by tracing the
// core of the reduction back to the gotos which initially generated the core. The goto follows of those gotos are then
// responsible for the reduction lookahead set.
func (b *ReductionLookaheadBuilder) calculateReduceActionLookaheads() {
	for i := range b.reduceActions {
		for _, gotoIdx := range b.getGeneratedGotoIdxs(b.reduceActions[i].StateIdx, b.reduceActions[i].Core) {
			b.reduceActions[i].LookaheadSet.Merge(&b.gotoFollows[gotoIdx])
		}
	}
}

// buildGotoFollowsSuccessorRelations is building up the digraph for the goto follows successor relation as specified
// in IELR(1) definition 3.5. We are taking all the gotos we found to happen on nullable nonterminals during LR(0)
// state construction, and we are creating edges to those gotos from the gotos which are pointing to the same state.
func (b *ReductionLookaheadBuilder) buildGotoFollowsSuccessorRelations() {
	// Index all gotos by their target state, so we can find the gotos entering a state without scanning every goto for
	// each candidate. This turns the relation construction from quadratic into linear in the number of gotos plus the
	// number of produced edges.
	gotoIdxsByToStateIdx := make([][]int, len(b.states))
	for gotoIdx := range b.gotoRecords {
		toStateIdx := b.gotoRecords[gotoIdx].ToStateIdx
		gotoIdxsByToStateIdx[toStateIdx] = append(gotoIdxsByToStateIdx[toStateIdx], gotoIdx)
	}

	for _, nullableGotoIdx := range b.successorDependencyCandidates {
		// A goto g contributes an edge to the nullable goto g' exactly when to_state[g] = from_state[g'], so we only
		// need the gotos which end in the state the nullable goto starts from.
		fromStateIdx := b.gotoRecords[nullableGotoIdx].FromStateIdx
		for _, gotoIdx := range gotoIdxsByToStateIdx[fromStateIdx] {
			b.gotoFollowsSuccessorRelation = append(b.gotoFollowsSuccessorRelation, Edge{
				FromIdx: gotoIdx,
				ToIdx:   nullableGotoIdx,
			})
		}
	}
}

// buildGotoFollowsInternalRelations builds up the digraph for the goto follows internal relation as specified in
// definition 3.8 of IELR(1).
func (b *ReductionLookaheadBuilder) buildGotoFollowsInternalRelations() {
	for _, candidate := range b.internalDependencyCandidates {
		// We are looking for gotos within the same state which are done on the nonterminal which is on the left hand
		// side of the item of the candidate.
		stateIdx := b.gotoRecords[candidate.GotoIdx].FromStateIdx
		gotoIdxs := b.gotoIdxsByStateIdx[stateIdx]
		for _, gotoIdx := range gotoIdxs {
			if b.gotoRecords[gotoIdx].NonterminalIdx != candidate.NonterminalIdx {
				// This can not be an internal dependency, as the goto is happening on a different symbol than the
				// candidate.
				continue
			}
			// The internal relations are needed separately in later IELR(1) phases, therefore we note them down
			// separately.
			b.gotoFollowsInternalRelation = append(b.gotoFollowsInternalRelation, Edge{
				FromIdx: candidate.GotoIdx,
				ToIdx:   gotoIdx,
			})
		}
	}
}

// buildGotoFollowsPredecessorRelations builds up the digraph for the goto follows predecessor relation as specified in
// definition 3.9 of IELR(1). This is done by moving backwards through the states to find the goto which generated the
// core of our candidate goto.
func (b *ReductionLookaheadBuilder) buildGotoFollowsPredecessorRelations() {
	for _, candidate := range b.predecessorDependencyCandidates {
		gotoIdxs := b.getGeneratedGotoIdxs(b.gotoRecords[candidate.GotoIdx].FromStateIdx, candidate.Core)
		for _, gotoIdx := range gotoIdxs {
			b.gotoFollowsPredecessorRelation = append(b.gotoFollowsPredecessorRelation, Edge{
				FromIdx: candidate.GotoIdx,
				ToIdx:   gotoIdx,
			})
		}
	}
}

// calculateAlwaysFollows fills the always follows as specified in definition 3.20 of IELR(1).
func (b *ReductionLookaheadBuilder) calculateAlwaysFollows() {
	b.alwaysFollows = make([]backend.LookaheadSet, len(b.gotoRecords))
	// Initialize the always follows with the terminal transitions of the target state.
	for i := range b.gotoRecords {
		toStateIdx := b.gotoRecords[i].ToStateIdx
		for _, transitionAction := range b.states[toStateIdx].TransitionActions.All() {
			symbolRef := transitionAction.SymbolRef()
			if symbolRef.IsTerminal() {
				b.alwaysFollows[i].Add(symbolRef.Idx())
			}
		}
	}

	gotoFollowsAlwaysRelation := make([]Edge, len(b.gotoFollowsSuccessorRelation)+len(b.gotoFollowsInternalRelation))
	copy(gotoFollowsAlwaysRelation[:len(b.gotoFollowsSuccessorRelation)], b.gotoFollowsSuccessorRelation)
	copy(gotoFollowsAlwaysRelation[len(b.gotoFollowsSuccessorRelation):], b.gotoFollowsInternalRelation)
	propagation := NewDigraphAlgorithm(b.alwaysFollows, gotoFollowsAlwaysRelation)
	propagation.Execute()
}

// calculateGotoFollows fills the goto follows as specified in definition 3.24 of IELR(1) by propagating the follow sets
// along the goto follows includes relations.
func (b *ReductionLookaheadBuilder) calculateGotoFollows() {
	b.gotoFollows = make([]backend.LookaheadSet, len(b.gotoRecords))
	// Initialize the goto follows with the always follows of the same goto.
	for i := range b.gotoRecords {
		b.gotoFollows[i].Merge(&b.alwaysFollows[i])
	}
	gotoFollowsIncludesRelation := make([]Edge, len(b.gotoFollowsInternalRelation)+len(b.gotoFollowsPredecessorRelation))
	copy(gotoFollowsIncludesRelation[:len(b.gotoFollowsInternalRelation)], b.gotoFollowsInternalRelation)
	copy(gotoFollowsIncludesRelation[len(b.gotoFollowsInternalRelation):], b.gotoFollowsPredecessorRelation)
	propagation := NewDigraphAlgorithm(b.gotoFollows, gotoFollowsIncludesRelation)
	propagation.Execute()
}

// getGeneratedGotoIdxs returns a list of goto indexes which generated the given core. This is done by tracing the core
// backwards through the states until we are at the start of the production and the left hand side appears in a goto.
func (b *ReductionLookaheadBuilder) getGeneratedGotoIdxs(stateIdx int, core backend.Core) []int {
	predecessorStateIdxs := []int{
		stateIdx,
	}

	// We need to move back through the states until we are at the start of the item.
	for position := core.Position(); position > 0; position-- {
		predecessorStateIdxs = b.followCoreBackward(backend.NewCore(core.ProductionIdx(), position), predecessorStateIdxs)
	}

	// Now let's look for the goto which has the left hand side of the production as a nonterminal transition.
	production := b.grammar.Productions[core.ProductionIdx()]
	var result []int
	for _, predecessorStateIdx := range predecessorStateIdxs {
		gotoIdxs := b.gotoIdxsByStateIdx[predecessorStateIdx]
		for _, gotoIdx := range gotoIdxs {
			if b.gotoRecords[gotoIdx].NonterminalIdx != production.NonterminalIdx {
				// We are looking for gotos on nonterminals which equal the left hand side of our production. This
				// is not one.
				continue
			}
			result = append(result, gotoIdx)
		}
	}
	return result
}

// followCoreBackward is moving the core one step back through the states and returns the list of state indexes the core
// was coming from.
func (b *ReductionLookaheadBuilder) followCoreBackward(core backend.Core, stateIdxs []int) []int {
	production := b.grammar.Productions[core.ProductionIdx()]
	symbolRef := production.SymbolRefs[core.Position()-1]
	var predecessorStateIdxs []int
	for _, stateIdx := range stateIdxs {
		// TODO: It is not necessary to check for specific symbols. As we only have kernel item cores in each state,
		// every predecessor must have the core in their state and every transition must be part of the core.
		if symbolRef.IsNonterminal() {
			predecessorStateIdxs = append(
				predecessorStateIdxs,
				b.backwardTransitionsByStateIdx[stateIdx].NonterminalTransitions[symbolRef.Idx()]...,
			)
		} else {
			predecessorStateIdxs = append(
				predecessorStateIdxs,
				b.backwardTransitionsByStateIdx[stateIdx].TerminalTransitions[symbolRef.Idx()]...,
			)
		}
	}
	return predecessorStateIdxs
}
