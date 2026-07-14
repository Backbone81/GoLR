package golr

import (
	"errors"
	"maps"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/utils"
)

// AnnotationsBuilder implements the algorithms for computing the state annotations necessary to decide which states
// need to be split. This is "Phase 2: Compute annotations" in section 3.4 of IELR(1).
type AnnotationsBuilder struct {
	// parser holds the LALR(1) parser tables of phase 0, together with the augmented grammar they were built from.
	parser backend.Parser

	// gotoRecords provides details about each nonterminal transition. This is derived from definition 3.4 of IELR(1).
	gotoRecords []GotoRecord

	// gotoIdxsByStateIdx provides a list of goto indexes when indexed by state index.
	gotoIdxsByStateIdx map[int][]int

	// gotoFollows holds the goto follow set for each goto, indexed by goto index. This is "goto_follows" from IELR(1)
	// definition 3.4.
	gotoFollows []backend.LookaheadSet

	// alwaysFollows holds the follow set from definition 3.20 of IELR(1), indexed by goto index.
	alwaysFollows []backend.LookaheadSet

	// predecessorStateIdxsByStateIdx provides the state indexes for the states which are predecessors when accessed by
	// a state index. This is definition 3.15 of IELR(1) and simply named "predecessors" there.
	predecessorStateIdxsByStateIdx [][]int

	// followKernelItemsByGotoIdx reports if the goto does depend on the kernel item's lookahead set. It holds the kernel
	// item indexes of the state the goto is coming from. This is definition 3.16 of IELR(1) and named
	// "follow_kernel_items" there.
	followKernelItemsByGotoIdx []utils.Bitset

	// itemLookaheadSetsByStateIdx caches the lookahead set of every kernel item of every state, indexed by state index
	// and then by kernel item index. This is definition 3.26 of IELR(1) and named "item_lookahead_sets" there.
	//
	// The lookahead sets are computed lazily, because only the kernel items which fall into the propagation path of a
	// conflicted terminal are ever needed: a nil lookahead set has not been computed yet, and the row of a state stays
	// nil until the first lookahead set of that state is requested. An empty lookahead set is a valid result, so it
	// cannot be what marks a lookahead set as not computed yet.
	itemLookaheadSetsByStateIdx [][]*backend.LookaheadSet

	// inadequaciesByStateIdx holds the inadequacies of the LALR(1) parser tables, keyed by the state index of the
	// conflicted state. Only conflicted states have an entry. This is definition 3.27 of IELR(1) and named
	// "inadequacy_lists" there.
	inadequaciesByStateIdx map[int][]*Inadequacy

	// annotationListsByStateIdx holds the annotations of a state, keyed by state index. Only states which are on a lane
	// of a conflicted state have an entry. This is definition 3.29 of IELR(1) and named "annotation_lists" there.
	annotationListsByStateIdx map[int][]Annotation
}

// annotatedState pairs a state with one of the annotations on it. This is the unit which the reverse iteration along
// the lanes of the conflicted states works on, and it needs both halves: the annotation is what gets carried to the
// predecessors of the state, and the state index is what says whose kernel items the contribution matrix of the
// annotation is indexed by. The annotation cannot tell us that itself, because the state index of its inadequacy is the
// conflicted state, which the iteration leaves behind with its first step.
type annotatedState struct {
	StateIdx   int
	Annotation Annotation
}

// NewAnnotationsBuilder creates a new builder for the state annotations of phase 2. It takes the LALR(1) parser tables
// of phase 0 together with the goto tables computed alongside them, and the auxiliary tables of phase 1.
func NewAnnotationsBuilder(
	parser backend.Parser,
	gotoRecords []GotoRecord,
	gotoIdxsByStateIdx map[int][]int,
	gotoFollows []backend.LookaheadSet,
	alwaysFollows []backend.LookaheadSet,
	predecessorStateIdxsByStateIdx [][]int,
	followKernelItemsByGotoIdx []utils.Bitset,
) *AnnotationsBuilder {
	return &AnnotationsBuilder{
		parser:                         parser,
		gotoRecords:                    gotoRecords,
		gotoIdxsByStateIdx:             gotoIdxsByStateIdx,
		gotoFollows:                    gotoFollows,
		alwaysFollows:                  alwaysFollows,
		predecessorStateIdxsByStateIdx: predecessorStateIdxsByStateIdx,
		followKernelItemsByGotoIdx:     followKernelItemsByGotoIdx,
	}
}

// Execute runs phase 2: it identifies the inadequacies of the LALR(1) parser tables and annotates the states along the
// lanes of the conflicted states. The results are only valid after Execute has run, and it must run exactly once.
func (b *AnnotationsBuilder) Execute() {
	b.initItemLookaheadSets()
	b.initInadequacies()
	b.initAnnotationLists()
}

// Parser returns the LALR(1) parser tables the annotations were computed for.
func (b *AnnotationsBuilder) Parser() backend.Parser {
	return b.parser
}

// Inadequacies returns the inadequacies of the LALR(1) parser tables, keyed by the state index of the conflicted
// state. This is definition 3.27 of IELR(1).
func (b *AnnotationsBuilder) Inadequacies() map[int][]*Inadequacy {
	return b.inadequaciesByStateIdx
}

// AnnotationLists returns the annotations of the states, keyed by state index. This is definition 3.29 of IELR(1).
func (b *AnnotationsBuilder) AnnotationLists() map[int][]Annotation {
	return b.annotationListsByStateIdx
}

// ItemLookaheadSet returns the lookahead set of a kernel item of a state. This is definition 3.26 of IELR(1) and named
// "item_lookahead_sets" there.
//
// The lookahead sets are computed lazily, so asking for one which phase 2 did not need itself computes it on the spot.
func (b *AnnotationsBuilder) ItemLookaheadSet(stateIdx int, itemIdx int) backend.LookaheadSet {
	return b.getItemLookaheadSet(stateIdx, itemIdx)
}

// initItemLookaheadSets initializes itemLookaheadSetsByStateIdx.
func (b *AnnotationsBuilder) initItemLookaheadSets() {
	b.itemLookaheadSetsByStateIdx = make([][]*backend.LookaheadSet, len(b.parser.States))
}

// getItemLookaheadSet returns the lookahead set for the kernel item of the state. The lookahead set is computed
// lazily on first use and cached afterward. This is definition 3.26 of IELR(1).
//
// Note that the position of a core is the number of symbols already seen, while the paper counts the dot position
// starting at one. The three cases of the definition are therefore at position 0, position 1 and position greater
// than 1 here.
func (b *AnnotationsBuilder) getItemLookaheadSet(stateIdx int, itemIdx int) backend.LookaheadSet {
	if b.itemLookaheadSetsByStateIdx[stateIdx] == nil {
		b.itemLookaheadSetsByStateIdx[stateIdx] = make(
			[]*backend.LookaheadSet,
			b.parser.States[stateIdx].KernelItems.Length(),
		)
	}
	if lookaheadSet := b.itemLookaheadSetsByStateIdx[stateIdx][itemIdx]; lookaheadSet != nil {
		return *lookaheadSet
	}

	var lookaheadSet backend.LookaheadSet
	core := b.parser.States[stateIdx].KernelItems.GetByIndex(itemIdx)
	switch core.Position() {
	case 0:
		// The only kernel item which has not seen any symbol yet is the one of the start production in the start
		// state. The end of input marker follows the start symbol, so there is nothing which could ever follow it and
		// the lookahead set stays empty. This is point 3 of the definition, which phase 2 never actually reaches.
	case 1:
		lookaheadSet = b.getItemLookaheadSetFromGotoFollows(stateIdx, core)
	default:
		lookaheadSet = b.getItemLookaheadSetFromPredecessors(stateIdx, core)
	}

	// The recursion above may have computed the lookahead sets of other states, but never the one of this kernel item,
	// because the position of the core strictly decreases with every recursion step and therefore never comes back to
	// it.
	b.itemLookaheadSetsByStateIdx[stateIdx][itemIdx] = &lookaheadSet
	return lookaheadSet
}

// getItemLookaheadSetFromPredecessors returns the lookahead set for a kernel item which has seen more than one symbol.
// Such a kernel item was created by a predecessor state moving over the symbol in front of it, so its lookahead set is
// the union of the lookahead sets of that very kernel item one position to the left in the predecessor states. This is
// point 1 of definition 3.26 of IELR(1).
func (b *AnnotationsBuilder) getItemLookaheadSetFromPredecessors(stateIdx int, core backend.Core) backend.LookaheadSet {
	predecessorCore := backend.NewCore(core.ProductionIdx(), core.Position()-1)

	var result backend.LookaheadSet
	for _, predecessorStateIdx := range b.predecessorStateIdxsByStateIdx[stateIdx] {
		for predecessorItemIdx, predecessorItem := range b.parser.States[predecessorStateIdx].KernelItems.All() {
			if predecessorItem != predecessorCore {
				// This is some other kernel item and not the one we are looking for. Continue with the next one.
				continue
			}
			predecessorLookaheadSet := b.getItemLookaheadSet(predecessorStateIdx, predecessorItemIdx)
			result.Merge(&predecessorLookaheadSet)
			// The kernel item can only be there once, so we can exit the inner loop early after we found the item.
			break
		}
	}
	return result
}

// getItemLookaheadSetFromGotoFollows returns the lookahead set for a kernel item which has seen exactly one symbol.
// Such a kernel item was added to the closure of the predecessor states by an item which is in front of the
// nonterminal on the left hand side of our production, so its lookahead set is what follows that nonterminal in the
// predecessor states. That is the goto follow set of the goto on that nonterminal. This is point 2 of definition 3.26
// of IELR(1).
func (b *AnnotationsBuilder) getItemLookaheadSetFromGotoFollows(stateIdx int, core backend.Core) backend.LookaheadSet {
	nonterminalIdx := b.parser.Grammar.Productions[core.ProductionIdx()].NonterminalIdx

	var result backend.LookaheadSet
	for _, predecessorStateIdx := range b.predecessorStateIdxsByStateIdx[stateIdx] {
		gotoIdx, ok := b.getGotoIdx(predecessorStateIdx, nonterminalIdx)
		if !ok {
			continue
		}
		result.Merge(&b.gotoFollows[gotoIdx])
	}
	return result
}

// getGotoIdx returns the goto index of the goto which happens on the nonterminal in the given state. A state cannot
// have more than one goto on the same nonterminal, so the goto is unique.
func (b *AnnotationsBuilder) getGotoIdx(stateIdx int, nonterminalIdx int) (int, bool) {
	for _, gotoIdx := range b.gotoIdxsByStateIdx[stateIdx] {
		if b.gotoRecords[gotoIdx].NonterminalIdx == nonterminalIdx {
			return gotoIdx, true
		}
	}
	return 0, false
}

// initInadequacies initializes inadequaciesByStateIdx as specified in definition 3.27 of IELR(1).
func (b *AnnotationsBuilder) initInadequacies() {
	b.inadequaciesByStateIdx = make(map[int][]*Inadequacy)
	for stateIdx := range b.parser.States {
		contributionsByTerminalIdx := conflict.ContributionsByTerminalIdx(b.parser.States[stateIdx])
		for _, terminalIdx := range slices.Sorted(maps.Keys(contributionsByTerminalIdx)) {
			contributions := contributionsByTerminalIdx[terminalIdx]
			if contributions.Length() <= 1 {
				// The terminal has a single action only, so there is no conflict and no inadequacy.
				continue
			}
			b.inadequaciesByStateIdx[stateIdx] = append(b.inadequaciesByStateIdx[stateIdx], &Inadequacy{
				StateIdx:      stateIdx,
				TerminalIdx:   terminalIdx,
				Contributions: contributions,
			})
		}
	}
}

// initAnnotationLists initializes annotationListsByStateIdx as specified in definition 3.29 of IELR(1).
//
// The definition is recursive: every conflicted state is annotated from its own inadequacies, and every predecessor of
// an annotated state is annotated from the annotations of that state. We drive that recursion with a work list, which
// keeps the reverse iteration along the lanes of the conflicted states flat and gives us a single place where the
// iteration terminates.
func (b *AnnotationsBuilder) initAnnotationLists() {
	b.annotationListsByStateIdx = make(map[int][]Annotation)

	workList := utils.NewDynamicRingBuffer[annotatedState]()
	for _, stateIdx := range slices.Sorted(maps.Keys(b.inadequaciesByStateIdx)) {
		for _, inadequacy := range b.inadequaciesByStateIdx[stateIdx] {
			// This is point 1 of definition 3.29.
			b.addAnnotation(&workList, stateIdx, b.annotateManifestation(inadequacy))
		}
	}
	for !workList.IsEmpty() {
		curr := workList.Remove()
		for _, predecessorStateIdx := range b.predecessorStateIdxsByStateIdx[curr.StateIdx] {
			// This is point 2 of definition 3.29.
			b.addAnnotation(&workList, predecessorStateIdx, b.annotatePredecessor(
				predecessorStateIdx,
				curr.StateIdx,
				curr.Annotation,
			))
		}
	}
}

// addAnnotation adds the annotation to the annotation list of the state and schedules the state for annotating its
// predecessors. It drops the annotation and terminates the iteration along the lane in the two cases described by
// definition 3.29 and observation 3.34 of IELR(1):
//
//  1. The annotation is useless, because splitting the state cannot change which contribution dominates the conflict.
//  2. The state already carries an identical annotation, so iterating further would only replicate annotations which
//     have been computed already. This is what guarantees termination, as the number of possible annotations is finite
//     while the lanes of a conflicted state can be cyclic.
//
// A state which has no predecessors terminates the iteration along the lane implicitly.
func (b *AnnotationsBuilder) addAnnotation(
	workList *utils.DynamicRingBuffer[annotatedState],
	stateIdx int,
	annotation Annotation,
) {
	if annotation.IsSplitStable() {
		return
	}
	for _, existingAnnotation := range b.annotationListsByStateIdx[stateIdx] {
		if existingAnnotation.Equal(&annotation) {
			return
		}
	}
	b.annotationListsByStateIdx[stateIdx] = append(b.annotationListsByStateIdx[stateIdx], annotation)
	workList.Add(annotatedState{
		StateIdx:   stateIdx,
		Annotation: annotation,
	})
}

// annotateManifestation computes the annotation for the conflicted state of the inadequacy. This is definition 3.30 of
// IELR(1).
func (b *AnnotationsBuilder) annotateManifestation(inadequacy *Inadequacy) Annotation {
	state := b.parser.States[inadequacy.StateIdx]
	contributionMatrix := make(ContributionMatrix, inadequacy.Contributions.Length())
	for contributionIdx, contribution := range inadequacy.Contributions.All() {
		if contribution.IsShiftAction() {
			// Every isocore split from the conflicted state keeps the transition on the conflicted terminal, so the
			// shift is an always contribution and the contribution row stays undefined. This is point 1 of the
			// definition.
			continue
		}
		production := b.parser.Grammar.Productions[contribution.ProductionIdx()]
		if len(production.SymbolRefs) == 0 {
			// The production is empty, so there is no kernel item in the conflicted state which carries the lookahead
			// set of the reduction. The reduction happens when the conflicted terminal can follow the nonterminal on
			// the left hand side of the production in this state. This is point 2b of the definition.
			contributionMatrix[contributionIdx] = b.computeLhsContributions(
				inadequacy.StateIdx,
				production.NonterminalIdx,
				inadequacy.TerminalIdx,
			)
			continue
		}

		// The reduction happens on the kernel item which has seen the full production. The isocore makes the
		// contribution when the conflicted terminal is in the lookahead set of that kernel item. This is point 2a of
		// the definition.
		reduceCore := backend.NewCore(contribution.ProductionIdx(), len(production.SymbolRefs))
		contributionRow := ContributionRow{Defined: true}
		for kernelItemIdx, kernelItem := range state.KernelItems.All() {
			if kernelItem == reduceCore {
				contributionRow.KernelItems.Add(kernelItemIdx)
			}
		}
		utils.DebugAssert(func() error {
			if contributionRow.KernelItems.Length() != 1 {
				// A state which reduces a non-empty production has the kernel item which has seen the full
				// production, and cores are unique within a state, so exactly one kernel item must match. Without a
				// match the contribution row would silently describe a never contribution instead.
				return errors.New("conflicted state does not have exactly one kernel item for the reduction")
			}
			return nil
		})
		contributionMatrix[contributionIdx] = contributionRow
	}
	return Annotation{
		Inadequacy:         inadequacy,
		ContributionMatrix: contributionMatrix,
	}
}

// computeLhsContributions computes the contribution row for a reduction on a production whose left hand side is the
// nonterminal, in a state where the reduction leaves no kernel item behind to carry the lookahead set. This is
// definition 3.31 of IELR(1).
//
// The reduction contributes when the terminal can follow the nonterminal in this state, which is what the goto follow
// set of the goto on that nonterminal describes. When the terminal is in the always follows of that goto, it follows
// the nonterminal no matter what the kernel item lookahead sets of the state are, so the contribution is an always
// contribution. Otherwise the terminal can only reach the goto follow set through the kernel items the goto follow set
// depends on, so the contribution depends on exactly those kernel items which currently see the terminal.
func (b *AnnotationsBuilder) computeLhsContributions(stateIdx int, nonterminalIdx int, terminalIdx int) ContributionRow {
	gotoIdx, ok := b.getGotoIdx(stateIdx, nonterminalIdx)
	utils.DebugAssert(func() error {
		if !ok {
			// A state which reduces a production, or which is the predecessor of a state that has seen the first
			// symbol of a production, always has a goto on the nonterminal on the left hand side of that production.
			return errors.New("state does not have a goto on the nonterminal")
		}
		return nil
	})

	if b.alwaysFollows[gotoIdx].Contains(terminalIdx) {
		return ContributionRow{}
	}

	contributionRow := ContributionRow{Defined: true}
	for kernelItemIdx := range b.parser.States[stateIdx].KernelItems.Length() {
		if !b.followKernelItemsByGotoIdx[gotoIdx].Contains(kernelItemIdx) {
			// The goto follow set does not depend on the lookahead set of this kernel item, so the terminal cannot
			// reach the goto follow set through it. We check this before asking for the item lookahead set, because
			// computing item lookahead sets is the expensive part of phase 2.
			continue
		}
		itemLookaheadSet := b.getItemLookaheadSet(stateIdx, kernelItemIdx)
		if !itemLookaheadSet.Contains(terminalIdx) {
			continue
		}
		contributionRow.KernelItems.Add(kernelItemIdx)
	}
	return contributionRow
}

// annotatePredecessor computes the annotation for a predecessor of an already annotated state. This is definition 3.32
// of IELR(1).
//
// The annotation of the successor tells us which of its kernel items must see the conflicted terminal for the isocore
// to make a contribution. This function translates those kernel items back into the kernel items of the predecessor
// which feed them, which is where the lanes of the conflicted state are traced backwards.
func (b *AnnotationsBuilder) annotatePredecessor(
	stateIdx int,
	successorStateIdx int,
	successorAnnotation Annotation,
) Annotation {
	terminalIdx := successorAnnotation.Inadequacy.TerminalIdx
	successorState := b.parser.States[successorStateIdx]

	// The same nonterminal can be on the left hand side of several of the successor kernel items we look at, so we
	// cache the contribution rows we compute for this predecessor.
	lhsContributions := make(map[int]ContributionRow)
	getLhsContributions := func(nonterminalIdx int) ContributionRow {
		if contributionRow, ok := lhsContributions[nonterminalIdx]; ok {
			return contributionRow
		}
		contributionRow := b.computeLhsContributions(stateIdx, nonterminalIdx, terminalIdx)
		lhsContributions[nonterminalIdx] = contributionRow
		return contributionRow
	}

	contributionMatrix := make(ContributionMatrix, len(successorAnnotation.ContributionMatrix))
	for contributionIdx, successorContributionRow := range successorAnnotation.ContributionMatrix {
		if !successorContributionRow.Defined {
			// An always contribution of the successor stays an always contribution in the predecessor. This is the
			// first half of point 1 of the definition.
			continue
		}
		if b.isAlwaysContribution(successorState, successorContributionRow, getLhsContributions) {
			continue
		}
		contributionMatrix[contributionIdx] = b.getPredecessorContributions(
			stateIdx,
			successorState,
			successorContributionRow,
			terminalIdx,
			getLhsContributions,
		)
	}
	return Annotation{
		Inadequacy:         successorAnnotation.Inadequacy,
		ContributionMatrix: contributionMatrix,
	}
}

// isAlwaysContribution reports if a contribution of the successor becomes an always contribution in the predecessor.
// This is the second half of point 1 of definition 3.32 of IELR(1).
//
// The contribution depends on a kernel item of the successor which has seen a single symbol only. The lookahead set of
// such a kernel item is the goto follow set of the goto on the nonterminal of its production in the predecessor. When
// the conflicted terminal is an always follow of that goto, the kernel item sees the terminal in every isocore of the
// successor which the predecessor can reach, so the contribution is made no matter how the predecessor is split.
func (b *AnnotationsBuilder) isAlwaysContribution(
	successorState backend.State,
	successorContributionRow ContributionRow,
	getLhsContributions func(nonterminalIdx int) ContributionRow,
) bool {
	for successorItemIdx := range successorContributionRow.KernelItems.All() {
		successorCore := successorState.KernelItems.GetByIndex(successorItemIdx)
		if successorCore.Position() != 1 {
			continue
		}
		nonterminalIdx := b.parser.Grammar.Productions[successorCore.ProductionIdx()].NonterminalIdx
		if getLhsContributions(nonterminalIdx).IsAlways() {
			return true
		}
	}
	return false
}

// getPredecessorContributions computes the contribution row of the predecessor from the contribution row of the
// successor. This is point 2 of definition 3.32 of IELR(1).
//
// A kernel item of the predecessor carries the contribution when it feeds a kernel item of the successor the
// contribution depends on. There are two ways for a kernel item of the successor to be fed. It was moved one position
// to the right when the predecessor moved over the symbol in front of it, in which case the kernel item of the
// predecessor must see the conflicted terminal. Or it has seen a single symbol only and was created by the closure of
// the predecessor, in which case the kernel items of the predecessor which let the terminal follow the nonterminal of
// its production carry the contribution.
func (b *AnnotationsBuilder) getPredecessorContributions(
	stateIdx int,
	successorState backend.State,
	successorContributionRow ContributionRow,
	terminalIdx int,
	getLhsContributions func(nonterminalIdx int) ContributionRow,
) ContributionRow {
	state := b.parser.States[stateIdx]

	contributionRow := ContributionRow{Defined: true}
	for successorItemIdx := range successorContributionRow.KernelItems.All() {
		successorCore := successorState.KernelItems.GetByIndex(successorItemIdx)

		// This is point 2(b)ii of the definition.
		if successorCore.Position() == 1 {
			nonterminalIdx := b.parser.Grammar.Productions[successorCore.ProductionIdx()].NonterminalIdx
			lhsContributionRow := getLhsContributions(nonterminalIdx)
			utils.DebugAssert(func() error {
				if !lhsContributionRow.Defined {
					// An undefined contribution row holds an empty kernel item set, so merging it would silently turn
					// an always contribution into a never contribution. It cannot be undefined here, because
					// isAlwaysContribution already checked every kernel item of the successor contribution row and
					// made the whole contribution an always contribution if any of them had an undefined row.
					return errors.New("the contribution row of the left hand side is undefined")
				}
				return nil
			})
			contributionRow.KernelItems.Merge(&lhsContributionRow.KernelItems)
		}

		// This is point 2(b)i of the definition. The successor has a predecessor, so it is not the start state and all
		// its kernel items have seen at least one symbol.
		predecessorCore := backend.NewCore(successorCore.ProductionIdx(), successorCore.Position()-1)
		for kernelItemIdx, kernelItem := range state.KernelItems.All() {
			if kernelItem != predecessorCore {
				continue
			}
			itemLookaheadSet := b.getItemLookaheadSet(stateIdx, kernelItemIdx)
			if itemLookaheadSet.Contains(terminalIdx) {
				contributionRow.KernelItems.Add(kernelItemIdx)
			}
			// The kernel item can only be there once, so we can exit the inner loop early after we found the item.
			break
		}
	}
	return contributionRow
}
