package golr

import (
	"context"
	"errors"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// SplitStatesBuilder implements the algorithms for splitting the LALR(1) states into the isocores of the minimal LR(1)
// parser tables. This is "Phase 3: Split states" in section 3.5 of IELR(1).
//
// Phase 3 starts from the LALR(1) automaton of phase 0 and, guided by the annotations of phase 2, splits every LALR(1)
// state into as many isocores as are needed to keep the parses of canonical LR(1) which merging into a single state
// would have lost. It behaves like canonical LR(1), Pager's algorithm and phase 0 step 1: it walks the automaton from
// the start state, propagates the recomputed kernel item lookahead sets along the transitions, and reuses a compatible
// existing isocore instead of creating a new one whenever the state compatibility test of definition 3.43 allows it.
//
// The states it produces still carry the LALR(1) reduction lookahead sets on their reduce actions. Phase 3 discards
// those conceptually (they are recomputed for the split automaton in phase 4); this builder leaves them in place, as
// phase 4 overwrites them anyway and only the productions the reduce actions name are read before then.
type SplitStatesBuilder struct {
	// grammar is the augmented context free grammar the automaton was built from.
	grammar frontend.Grammar

	// maxStates is the number of states after which the splitting gives up with backend.ErrStateLimitExceeded.
	maxStates int

	// policy is the conflict resolution by which the dominant contribution of definition 3.42 is decided. It is the
	// paper's Delta. A policy which leaves conflicts unresolved makes phase 3 preserve every lookahead distinction
	// canonical LR(1) makes, which is what keeps the builder conflict-preserving.
	policy conflict.Policy

	// states is the set of states computed so far. This is the paper's Sigma from section 3.5. It starts as the LALR(1)
	// states of phase 0 and grows as phase 3 splits states into isocores. The transitions of the states are redirected
	// in place as isocores are created and reused.
	states []backend.State

	// annotationListsByStateIdx holds the annotations of a state, keyed by state index. Only the original LALR(1) states
	// have an entry, which is why it is always accessed through the LALR(1) isocore of a state. This is definition 3.29
	// of IELR(1).
	annotationListsByStateIdx map[int][]Annotation

	// gotoRecords provides details about each nonterminal transition of the LALR(1) automaton. This is derived from
	// definition 3.4 of IELR(1).
	gotoRecords []GotoRecord

	// gotoIdxsByStateIdx provides a list of goto indexes when indexed by the state index of the LALR(1) state the goto
	// comes from.
	gotoIdxsByStateIdx map[int][]int

	// alwaysFollows holds the follow set from definition 3.20 of IELR(1), indexed by goto index.
	alwaysFollows []backend.LookaheadSet

	// followKernelItemsByGotoIdx reports which kernel items of the state a goto comes from the goto follow set depends
	// on. This is definition 3.16 of IELR(1) and named "follow_kernel_items" there.
	followKernelItemsByGotoIdx []utils.Bitset

	// lalr1IsocoreByStateIdx maps a state index to the index of its LALR(1) isocore. An original LALR(1) state is its
	// own LALR(1) isocore. This is definition 3.36 of IELR(1) and named "lalr1_isocores" there.
	lalr1IsocoreByStateIdx []int

	// isocoreNextByStateIdx is a circularly linked list, keyed by state index, whose members form the set of states
	// which are isocores of each other. This is definition 3.37 of IELR(1) and named "isocore_nexts" there.
	isocoreNextByStateIdx []int

	// lookaheadsRecomputedByStateIdx reports if phase 3 has already computed the lookaheads which at least one
	// predecessor propagates to the item lookahead sets of a state. This is definition 3.41 of IELR(1) and named
	// "lookaheads_recomputed" there.
	lookaheadsRecomputedByStateIdx []bool

	// itemLookaheadSetsByStateIdx holds the recomputed lookahead set of every kernel item of every state, indexed by
	// state index and then by kernel item index. This is definition 3.26 of IELR(1) recomputed by phase 3 and named
	// "item_lookahead_sets" there. A nil row is a state phase 3 has not propagated any lookaheads to yet, which for an
	// original LALR(1) state means it is still a placeholder for the lookaheads of its first predecessor.
	itemLookaheadSetsByStateIdx [][]backend.LookaheadSet

	// lookaheadSetFiltersByLalr1Isocore caches the lookahead set filters of definition 3.38, keyed by LALR(1) isocore
	// state index. The filters depend only on the annotations of the LALR(1) isocore, so every isocore of the same
	// LALR(1) state shares them.
	lookaheadSetFiltersByLalr1Isocore map[int][]backend.LookaheadSet
}

// NewSplitStatesBuilder returns a new builder which splits the LALR(1) states into the isocores of the minimal LR(1)
// parser tables. It takes the LALR(1) states of phase 0, the annotations of phase 2, the goto tables of phase 0 and the
// auxiliary tables of phase 1, together with the conflict resolution policy which decides the dominant contribution of
// definition 3.42.
//
// The builder takes ownership of the states slice and mutates it in place: it redirects transitions and appends the
// isocores it splits off. This is sound because phase 3 never reads the original LALR(1) transitions again - it walks the
// automaton through the transitions it is redirecting, and the goto follow sets it recomputes come from the auxiliary
// tables keyed by the LALR(1) isocore state index, not from the transitions. The goto tables and annotations refer to
// the LALR(1) states only by state index and kernel item, both of which stay stable, so mutating the transition and
// reduce actions does not disturb them. The caller must not keep using the states it passes in.
func NewSplitStatesBuilder(
	grammar frontend.Grammar,
	states []backend.State,
	policy conflict.Policy,
	annotationListsByStateIdx map[int][]Annotation,
	gotoRecords []GotoRecord,
	gotoIdxsByStateIdx map[int][]int,
	alwaysFollows []backend.LookaheadSet,
	followKernelItemsByGotoIdx []utils.Bitset,
) SplitStatesBuilder {
	return SplitStatesBuilder{
		grammar:                           grammar,
		maxStates:                         backend.MaxAddressableStates(grammar),
		policy:                            policy,
		states:                            states,
		annotationListsByStateIdx:         annotationListsByStateIdx,
		gotoRecords:                       gotoRecords,
		gotoIdxsByStateIdx:                gotoIdxsByStateIdx,
		alwaysFollows:                     alwaysFollows,
		followKernelItemsByGotoIdx:        followKernelItemsByGotoIdx,
		lookaheadSetFiltersByLalr1Isocore: make(map[int][]backend.LookaheadSet),
	}
}

// Build runs phase 3: it splits the LALR(1) states into the isocores of the minimal LR(1) parser tables. The result
// is only valid after Build has run, and it must run exactly once. This is the split_states routine of definition
// 3.45 of IELR(1).
//
// It gives up with backend.ErrStateLimitExceeded once the splitting grows the automaton beyond what a parser table can
// address. Phase 3 is bounded above by canonical LR(1) and normally stays close to LALR(1), so this needs a grammar
// which is both large and heavily non-LALR, but nothing in the algorithm rules that out and the split states are handed
// to backend.NewTransitionAction like any others.
func (b *SplitStatesBuilder) Build() error {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 3: split states").End()

	// This is the initialization of lines 1-4 of definition 3.45 for the original LALR(1) states. The isocores which
	// are appended later initialize their own entries in computeState.
	b.lalr1IsocoreByStateIdx = make([]int, len(b.states))
	b.isocoreNextByStateIdx = make([]int, len(b.states))
	b.lookaheadsRecomputedByStateIdx = make([]bool, len(b.states))
	b.itemLookaheadSetsByStateIdx = make([][]backend.LookaheadSet, len(b.states))
	for stateIdx := range b.states {
		b.lalr1IsocoreByStateIdx[stateIdx] = stateIdx
		b.isocoreNextByStateIdx[stateIdx] = stateIdx
	}

	// This is the main loop of lines 5-9 of definition 3.45. The length of the state slice grows as isocores are
	// appended, and the loop processes those newly appended states as well. Every non-start state has a predecessor with
	// a smaller index in the breadth-first order the LR(0) construction assigns, so a state's lookaheads are always
	// recomputed by a predecessor before the loop reaches the state itself.
	for stateIdx := 0; stateIdx < len(b.states); stateIdx++ {
		for _, symbolRef := range b.transitionSymbolRefs(stateIdx) {
			b.computeState(stateIdx, symbolRef)
		}

		// Every transition of the state can split off at most one isocore, so a single state can push us over the limit
		// by as many states as it has transitions. That overshoot is what MaxAddressableStates leaves room for, which
		// lets us check once per state instead of on every isocore.
		if err := backend.CheckStateLimit("IELR(1)", len(b.states), b.maxStates); err != nil {
			return err
		}
	}
	return nil
}

// States returns the states of the split automaton. The result is only valid after Build has run.
func (b *SplitStatesBuilder) States() []backend.State {
	return b.states
}

// computeState propagates the recomputed lookaheads from a state to the successor a transition leads to, and either
// reuses a compatible isocore of the successor or splits off a new one. This is definition 3.47 of IELR(1).
//
// The transition is identified by the symbol it happens on rather than by its index, because the transition actions of
// a state are an ordered set. The symbol is stable while the target of the transition is redirected, so it names the
// transition uniquely throughout.
func (b *SplitStatesBuilder) computeState(fromStateIdx int, symbolRef frontend.SymbolRef) {
	successorStateIdx := b.transitionTarget(fromStateIdx, symbolRef)
	lookaheads := b.propagateLookaheads(fromStateIdx, successorStateIdx)

	// Walk the circular list of the isocores of the successor, looking for one which is compatible with the propagated
	// lookaheads. This is lines 2-10 of definition 3.47.
	found := false
	isocoreStateIdx := successorStateIdx
	for {
		if b.isCompatible(isocoreStateIdx, lookaheads) {
			found = true
			break
		}
		if b.isocoreNextByStateIdx[isocoreStateIdx] == successorStateIdx {
			// We are back at the start of the circular list, so no compatible isocore exists.
			break
		}
		isocoreStateIdx = b.isocoreNextByStateIdx[isocoreStateIdx]
	}

	switch {
	case !found:
		// No compatible isocore exists, so split off a new one and redirect the transition to it. This is lines 14-20 of
		// definition 3.47.
		newStateIdx := b.appendIsocore(successorStateIdx, lookaheads)
		b.redirectTransition(fromStateIdx, symbolRef, newStateIdx)
	case !b.lookaheadsRecomputedByStateIdx[isocoreStateIdx]:
		// The compatible isocore is an original LALR(1) state no predecessor has propagated to yet, so this is its first
		// predecessor. It adopts the propagated lookaheads. The transition already points at it, so it is not redirected.
		// This is lines 21-23 of definition 3.47.
		b.itemLookaheadSetsByStateIdx[isocoreStateIdx] = lookaheads
		b.lookaheadsRecomputedByStateIdx[isocoreStateIdx] = true
	default:
		// The compatible isocore already has recomputed lookaheads, so redirect the transition to it and merge the
		// propagated lookaheads into it. This is lines 24-26 of definition 3.47.
		b.redirectTransition(fromStateIdx, symbolRef, isocoreStateIdx)
		b.mergeLookaheads(isocoreStateIdx, lookaheads)
	}
}

// appendIsocore splits off a new isocore of the LALR(1) isocore of the successor, carrying the propagated lookaheads,
// and returns its state index. This is lines 14-19 of definition 3.47 of IELR(1).
//
// This deviates from the paper in two details: line 14 clones the last isocore the compatibility walk visited and lines
// 16-17 insert the new state after that isocore, which is just in front of the successor in the circular list, while we
// clone the successor itself and insert right after it. Both choices are behaviorally equivalent: isocores of the same
// LALR(1) state differ only in the current targets of their transitions, and the main loop of split_states reprocesses
// every transition of the new state, walking the full circular list no matter which member the transition points at.
// The only visible difference is the order in which later compatibility walks visit the isocores, which can make them
// reuse a different, but equally compatible, isocore.
func (b *SplitStatesBuilder) appendIsocore(successorStateIdx int, lookaheads []backend.LookaheadSet) int {
	newState := b.states[successorStateIdx]
	newState.TransitionActions = b.states[successorStateIdx].TransitionActions.Clone()
	newState.ReduceActions = b.states[successorStateIdx].ReduceActions.Clone()
	b.states = append(b.states, newState)
	newStateIdx := len(b.states) - 1

	b.lalr1IsocoreByStateIdx = append(b.lalr1IsocoreByStateIdx, b.lalr1IsocoreByStateIdx[successorStateIdx])

	// Insert the new isocore right after the successor in its circular list.
	b.isocoreNextByStateIdx = append(b.isocoreNextByStateIdx, b.isocoreNextByStateIdx[successorStateIdx])
	b.isocoreNextByStateIdx[successorStateIdx] = newStateIdx

	b.lookaheadsRecomputedByStateIdx = append(b.lookaheadsRecomputedByStateIdx, true)
	b.itemLookaheadSetsByStateIdx = append(b.itemLookaheadSetsByStateIdx, lookaheads)
	return newStateIdx
}

// mergeLookaheads merges the propagated lookaheads into the item lookahead sets of a state and re-propagates them to the
// successors of that state if it gained new lookaheads. This is definition 3.46 of IELR(1).
//
// The re-propagation stops at the first successor phase 3 has not recomputed yet, because the main loop of split_states
// has not propagated this state's lookaheads along that transition or any transition after it, so it will propagate the
// new lookaheads there later and there is no need to recurse.
func (b *SplitStatesBuilder) mergeLookaheads(stateIdx int, lookaheads []backend.LookaheadSet) {
	newLookaheads := false
	for kernelItemIdx := range b.states[stateIdx].KernelItems.Length() {
		// Adding the propagated lookaheads reports whether any of them were new, which is exactly the difference the
		// definition keeps in the reduced lookahead set before merging it.
		if b.itemLookaheadSetsByStateIdx[stateIdx][kernelItemIdx].Merge(&lookaheads[kernelItemIdx]) {
			newLookaheads = true
		}
	}
	if !newLookaheads {
		return
	}
	for _, symbolRef := range b.transitionSymbolRefs(stateIdx) {
		successorStateIdx := b.transitionTarget(stateIdx, symbolRef)
		if !b.lookaheadsRecomputedByStateIdx[successorStateIdx] {
			break
		}
		b.computeState(stateIdx, symbolRef)
	}
}

// propagateLookaheads computes the lookaheads a state propagates to a successor, filtered down to the terminals which
// can affect the state compatibility test. This is definition 3.40 of IELR(1) together with the lookahead set filters
// of definition 3.38.
//
// A successor kernel item which has seen more than one symbol inherits the lookahead set of the kernel item it was
// advanced from in the state. A successor kernel item which has seen a single symbol was added by the closure of the
// state, so its lookahead set is the goto follow set of the nonterminal on the left hand side of its production. The dot
// position d of the paper counts from one, so it is one more than the position of a core, which counts symbols seen.
func (b *SplitStatesBuilder) propagateLookaheads(fromStateIdx int, successorStateIdx int) []backend.LookaheadSet {
	filters := b.lookaheadSetFilters(successorStateIdx)

	result := make([]backend.LookaheadSet, b.states[successorStateIdx].KernelItems.Length())
	for successorItemIdx, successorCore := range b.states[successorStateIdx].KernelItems.All() {
		var lookaheadSet backend.LookaheadSet
		switch {
		case successorCore.Position() > 1:
			// This is point 1 of definition 3.40. The successor kernel item was advanced from the kernel item one
			// position to the left in the state.
			predecessorCore := backend.NewCore(successorCore.ProductionIdx(), successorCore.Position()-1)
			kernelItemIdx, ok := b.kernelItemIdx(fromStateIdx, predecessorCore)
			utils.DebugAssert(func() error {
				if !ok {
					return errors.New("the kernel item the successor kernel item was advanced from is missing in the state")
				}
				return nil
			})
			if ok {
				itemLookaheadSet := b.itemLookaheadSet(fromStateIdx, kernelItemIdx)
				lookaheadSet = itemLookaheadSet.Clone()
			}
		default:
			// This is point 2 of definition 3.40. The successor kernel item has seen a single symbol, so it was added by
			// the closure and its lookahead set follows the nonterminal on its left hand side. Position 0 is impossible
			// for a successor, as only the start state has a kernel item which has seen no symbol.
			utils.DebugAssert(func() error {
				if successorCore.Position() != 1 {
					return errors.New("a successor kernel item cannot have seen no symbol")
				}
				return nil
			})
			nonterminalIdx := b.grammar.Productions[successorCore.ProductionIdx()].NonterminalIdx
			lookaheadSet = b.computeGotoFollowSet(fromStateIdx, nonterminalIdx)
		}
		lookaheadSet.Intersect(&filters[successorItemIdx])
		result[successorItemIdx] = lookaheadSet
	}
	return result
}

// computeGotoFollowSet computes the goto follow set of the goto on a nonterminal in a state, using the recomputed item
// lookahead sets of that state instead of the ones of its LALR(1) isocore. This is definition 3.39 of IELR(1).
//
// The goto is the one on the nonterminal from the LALR(1) isocore of the state. Its follow set is the always follows of
// the goto together with the recomputed lookahead sets of the kernel items the goto follow set depends on.
func (b *SplitStatesBuilder) computeGotoFollowSet(stateIdx int, nonterminalIdx int) backend.LookaheadSet {
	lalr1IsocoreStateIdx := b.lalr1IsocoreByStateIdx[stateIdx]

	var result backend.LookaheadSet
	for _, gotoIdx := range b.gotoIdxsByStateIdx[lalr1IsocoreStateIdx] {
		if b.gotoRecords[gotoIdx].NonterminalIdx != nonterminalIdx {
			continue
		}
		result.Merge(&b.alwaysFollows[gotoIdx])
		for kernelItemIdx := range b.followKernelItemsByGotoIdx[gotoIdx].All() {
			itemLookaheadSet := b.itemLookaheadSet(stateIdx, kernelItemIdx)
			result.Merge(&itemLookaheadSet)
		}
	}
	return result
}

// lookaheadSetFilters returns the lookahead set filters of a state: for each kernel item, the terminals whose presence
// in the kernel item's lookahead set can change which contribution dominates an inadequacy. This is definition 3.38 of
// IELR(1).
//
// The filters depend only on the annotations of the LALR(1) isocore, so they are computed once per LALR(1) isocore and
// cached. Restricting the propagated lookaheads to the filters is what keeps phase 3 from recomputing lookaheads which
// no annotation examines.
func (b *SplitStatesBuilder) lookaheadSetFilters(stateIdx int) []backend.LookaheadSet {
	lalr1IsocoreStateIdx := b.lalr1IsocoreByStateIdx[stateIdx]
	if filters, ok := b.lookaheadSetFiltersByLalr1Isocore[lalr1IsocoreStateIdx]; ok {
		return filters
	}

	filters := make([]backend.LookaheadSet, b.states[lalr1IsocoreStateIdx].KernelItems.Length())
	for _, annotation := range b.annotationListsByStateIdx[lalr1IsocoreStateIdx] {
		terminalIdx := annotation.Inadequacy.TerminalIdx
		for _, contributionRow := range annotation.ContributionMatrix {
			if !contributionRow.IsPotential() {
				// Point 2 of definition 3.38 only ranges over the defined rows, and the kernel items of an always
				// contribution are meaningless, so only a potential contribution names kernel items to filter on.
				continue
			}
			for kernelItemIdx := range contributionRow.KernelItems.All() {
				filters[kernelItemIdx].Add(terminalIdx)
			}
		}
	}
	b.lookaheadSetFiltersByLalr1Isocore[lalr1IsocoreStateIdx] = filters
	return filters
}

// isCompatible reports if a state can carry the propagated lookaheads without changing any dominant contribution it
// already makes. This is definition 3.43 of IELR(1).
//
// A state no predecessor has propagated to yet is compatible with anything, because it is still a placeholder for the
// lookaheads of its first predecessor. Otherwise the state and the candidate lookaheads must agree on the dominant
// contribution of every annotation, unless one of them makes no contribution at all.
func (b *SplitStatesBuilder) isCompatible(stateIdx int, candidateLookaheads []backend.LookaheadSet) bool {
	if !b.lookaheadsRecomputedByStateIdx[stateIdx] {
		return true
	}
	currentLookaheads := b.itemLookaheadSetsByStateIdx[stateIdx]
	for _, annotation := range b.annotationListsByStateIdx[b.lalr1IsocoreByStateIdx[stateIdx]] {
		current := b.dominantContribution(stateIdx, annotation, currentLookaheads)
		candidate := b.dominantContribution(stateIdx, annotation, candidateLookaheads)
		if current.Equal(candidate) {
			// Point 1 of the definition: the state and the candidate lookaheads agree on the dominant contribution.
			continue
		}
		if current.Kind == conflict.DecisionUndefined || candidate.Kind == conflict.DecisionUndefined {
			// Points 2 and 3 of the definition: one side makes no contribution, so it says nothing about the terminal
			// and stays compatible with whatever the other side decides.
			continue
		}
		return false
	}
	return true
}

// dominantContribution computes the dominant contribution a state with the given lookahead sets would make to the
// inadequacy an annotation refers to. This is definition 3.42 of IELR(1).
//
// It first reconstructs the contributions the state actually makes: a contribution whose contribution row is an always
// contribution is always made, and a contribution whose row is a potential contribution is made when the conflicted
// terminal is in the lookahead set of one of the kernel items the row depends on. The dominant contribution among those
// is then decided by the conflict resolution policy.
func (b *SplitStatesBuilder) dominantContribution(
	stateIdx int,
	annotation Annotation,
	lookaheads []backend.LookaheadSet,
) conflict.Decision {
	terminalIdx := annotation.Inadequacy.TerminalIdx

	var madeContributions conflict.ContributionSet
	for contributionIdx, contribution := range annotation.Inadequacy.Contributions.All() {
		contributionRow := annotation.ContributionMatrix[contributionIdx]
		if contributionRow.IsAlways() {
			madeContributions.Add(contribution)
			continue
		}
		for kernelItemIdx := range contributionRow.KernelItems.All() {
			if lookaheads[kernelItemIdx].Contains(terminalIdx) {
				madeContributions.Add(contribution)
				break
			}
		}
	}
	utils.DebugAssert(func() error {
		if len(lookaheads) != b.states[stateIdx].KernelItems.Length() {
			return errors.New("the lookahead sets are not indexed by the kernel items of the state")
		}
		return nil
	})
	return conflict.DominantContribution(b.policy, terminalIdx, madeContributions)
}

// transitionSymbolRefs returns the symbols the transitions of a state happen on, in the order of the transition action
// set. Redirecting a transition changes only its target, never the symbol it happens on, so the symbols stay a stable
// identity for the transitions while phase 3 rewrites the automaton.
func (b *SplitStatesBuilder) transitionSymbolRefs(stateIdx int) []frontend.SymbolRef {
	result := make([]frontend.SymbolRef, 0, b.states[stateIdx].TransitionActions.Length())
	for _, transitionAction := range b.states[stateIdx].TransitionActions.All() {
		result = append(result, transitionAction.SymbolRef())
	}
	return result
}

// transitionTarget returns the state index the transition on the symbol currently leads to.
func (b *SplitStatesBuilder) transitionTarget(stateIdx int, symbolRef frontend.SymbolRef) int {
	for _, transitionAction := range b.states[stateIdx].TransitionActions.All() {
		if transitionAction.SymbolRef() == symbolRef {
			return transitionAction.StateIdx()
		}
	}
	utils.DebugAssert(func() error {
		return errors.New("the state does not have a transition on the symbol")
	})
	return 0
}

// redirectTransition points the transition on the symbol at a new target state. The transition action set is keyed by
// the symbol and the target together, so the old action is removed and a new one is added rather than amended in place.
func (b *SplitStatesBuilder) redirectTransition(stateIdx int, symbolRef frontend.SymbolRef, newTargetStateIdx int) {
	oldTargetStateIdx := b.transitionTarget(stateIdx, symbolRef)
	if oldTargetStateIdx == newTargetStateIdx {
		return
	}
	b.states[stateIdx].TransitionActions.Remove(backend.NewTransitionAction(symbolRef, oldTargetStateIdx))
	b.states[stateIdx].TransitionActions.Add(backend.NewTransitionAction(symbolRef, newTargetStateIdx))
}

// kernelItemIdx returns the index of a core within the kernel items of a state. Cores are unique within a state, so the
// index is unique when the core is present.
func (b *SplitStatesBuilder) kernelItemIdx(stateIdx int, core backend.Core) (int, bool) {
	for kernelItemIdx, kernelItem := range b.states[stateIdx].KernelItems.All() {
		if kernelItem == core {
			return kernelItemIdx, true
		}
	}
	return 0, false
}

// itemLookaheadSet returns the recomputed lookahead set of a kernel item of a state. A state phase 3 has not propagated
// any lookaheads to yet has no recomputed lookahead sets, so the lookahead set is empty. This is the start state, whose
// goto follow sets never depend on any of its kernel items, and any state which is read before its first predecessor has
// propagated to it, which the breadth-first ordering prevents.
func (b *SplitStatesBuilder) itemLookaheadSet(stateIdx int, kernelItemIdx int) backend.LookaheadSet {
	if b.itemLookaheadSetsByStateIdx[stateIdx] == nil {
		return backend.LookaheadSet{}
	}
	return b.itemLookaheadSetsByStateIdx[stateIdx][kernelItemIdx]
}
