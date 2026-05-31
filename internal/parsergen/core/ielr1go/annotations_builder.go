package ielr1go

import (
	"maps"
	"slices"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// AnnotationsBuilder implements the algorithms for computing the state annotations necessary to decide which states
// need to be split. This is "Phase 2: Compute annotations" in section 3.4 of IELR(1).
type AnnotationsBuilder struct {
	grammar      frontend.Grammar
	lalr1Builder LALR1Builder
	parser       backend.Parser

	// predecessorStateIdxsByStateIdx provides the state indexes for the states which are predecessors when accessed by
	// a state index. This is definition 3.15 of IELR(1) and simply named "predecessors" there.
	// TODO: This table might be valuable to calculate and use during LALR(1) construction already.
	predecessorStateIdxsByStateIdx [][]int

	// followKernelItems reports if the goto does depend on the kernel item's lookahead set. It is indexed by goto index
	// and then kernel item index. This is definition 3.16 of IELR(1) and named "follow_kernel_items" there.
	followKernelItems [][]bool

	// itemLookaheadSets provides the lookahead sets for the kernel items of a state. It is indexed by the state index.
	// This is definition 3.26 of IELR(1) and named "item_lookahead_sets" there.
	itemLookaheadSets []ItemLookaheadSets

	// inadequacyLists is a list of inadequacy manifestation descriptions indexed by state index. This is definition
	// 3.27 of IELR(1) .
	inadequacyLists []InadequacyManifestationDescriptions

	// inadequacyAnnotations provides a list of all inadequacy annotations over all states. It is indexed by inadequacy
	// annotation index.
	inadequacyAnnotations []InadequacyAnnotation

	// annotationLists is a list of inadequacy annotation indexes for each state indexed by state index. This is
	// definition 3.29 of IELR(1).
	annotationLists []InadequacyAnnotationIdxs
}

// ItemLookaheadSets provides the lookahead sets per kernel item of the state. It is indexed by the kernel item index.
type ItemLookaheadSets []backend.LookaheadSet

// InadequacyManifestationDescriptions provides the inadequacy manifestation descriptions of a state.
type InadequacyManifestationDescriptions []InadequacyManifestationDescription

// InadequacyManifestationDescription is a single inadequacy manifestation description.
type InadequacyManifestationDescription struct {
	StateIdx              int
	TerminalIdx           int
	ConflictContributions ConflictContributionSet
}

// InadequacyAnnotation describes how an entry of the inadequacy list impacts the conflict contributions.
type InadequacyAnnotation struct {
	InadequacyListIdx            int
	InadequacyContributionMatrix InadequacyContributionMatrix
}

// InadequacyContributionMatrix describes how any core split from the state contributes to the conflict contributions.
// It is indexed by the conflict contribution index of the state.
type InadequacyContributionMatrix []KernelItemConflictContributions

// KernelItemConflictContributions describes which kernel item of the state contributes to the conflict contribution.
// It is indexed by the kernel item index of the state.
type KernelItemConflictContributions []bool

type InadequacyAnnotationIdxs = utils.OrderedSet[int]

func NewAnnotationsBuilder(
	lalr1Builder LALR1Builder,
	parser backend.Parser,
	predecessorStateIdxsByStateIdx [][]int,
	followKernelItems [][]bool,
) *AnnotationsBuilder {
	return &AnnotationsBuilder{
		grammar:                        parser.Grammar,
		lalr1Builder:                   lalr1Builder,
		parser:                         parser,
		predecessorStateIdxsByStateIdx: predecessorStateIdxsByStateIdx,
		followKernelItems:              followKernelItems,
	}
}

func (b *AnnotationsBuilder) Execute() {
	b.initItemLookaheadSets()
	b.initInadequacyLists()

	// TODO: compute annotations
}

// initItemLookaheadSets initializes itemLookaheadSets.
func (b *AnnotationsBuilder) initItemLookaheadSets() {
	b.itemLookaheadSets = make([]ItemLookaheadSets, len(b.parser.States))
}

// getItemLookaheadSet returns the item lookahead set for the state index and the item index. The item lookahead set
// is lazily evaluated as needed. This is definition 3.26 of IELR(1).
func (b *AnnotationsBuilder) getItemLookaheadSet(stateIdx int, itemIdx int) backend.LookaheadSet {
	if b.itemLookaheadSets[stateIdx] != nil && !b.itemLookaheadSets[stateIdx][itemIdx].IsEmpty() {
		// This item lookahead set has already been constructed. We return the lookahead set which we calculated
		// earlier.
		return b.itemLookaheadSets[stateIdx][itemIdx]
	}
	currCore := b.parser.States[stateIdx].KernelItems.GetByIndex(itemIdx)
	if currCore.Position() > 1 {
		return b.getItemLookaheadSetFromPredecessors(stateIdx, itemIdx)
	} else {
		return b.getItemLookaheadSetFromGotoFollows(stateIdx, itemIdx)
	}
}

// getItemLookaheadSetFromPredecessors returns the item lookahead set for the state index and the item index derived
// from the predecessor items.
func (b *AnnotationsBuilder) getItemLookaheadSetFromPredecessors(stateIdx int, itemIdx int) backend.LookaheadSet {
	currCore := b.parser.States[stateIdx].KernelItems.GetByIndex(itemIdx)
	if b.itemLookaheadSets[stateIdx] == nil {
		// We need to allocate the lookup table per kernel item.
		b.itemLookaheadSets[stateIdx] = make([]backend.LookaheadSet, b.parser.States[stateIdx].KernelItems.Length())
	}

	// This is the core we need to find in predecessor states.
	currCorePredecessor := backend.NewCore(currCore.ProductionIdx(), currCore.Position()-1)
	for _, predecessorStateIdx := range b.predecessorStateIdxsByStateIdx[stateIdx] {
		for predecessorItemIdx, predecessorItem := range b.parser.States[predecessorStateIdx].KernelItems.All() {
			if predecessorItem != currCorePredecessor {
				// This is some other kernel item and not the one we are looking for. Continue with the next one.
				continue
			}
			predecessorItemLookaheadSet := b.getItemLookaheadSet(predecessorStateIdx, predecessorItemIdx)
			b.itemLookaheadSets[stateIdx][itemIdx].Merge(&predecessorItemLookaheadSet)
			// The kernel item can only be there once, so we can exit the inner loop early after we found the
			// item.
			break
		}
	}
	return b.itemLookaheadSets[stateIdx][itemIdx]
}

// getItemLookaheadSetFromGotoFollows returns the item lookahead set for the state index and the item index derived
// from the goto follows.
func (b *AnnotationsBuilder) getItemLookaheadSetFromGotoFollows(stateIdx int, itemIdx int) backend.LookaheadSet {
	currCore := b.parser.States[stateIdx].KernelItems.GetByIndex(itemIdx)
	if b.itemLookaheadSets[stateIdx] == nil {
		// We need to allocate the lookup table per kernel item.
		b.itemLookaheadSets[stateIdx] = make([]backend.LookaheadSet, b.parser.States[stateIdx].KernelItems.Length())
	}

	for _, predecessorStateIdx := range b.predecessorStateIdxsByStateIdx[stateIdx] {
		for _, gotoIdx := range b.lalr1Builder.gotoIdxsByStateIdx[predecessorStateIdx] {
			production := b.grammar.Productions[currCore.ProductionIdx()]
			if b.lalr1Builder.gotoRecords[gotoIdx].NonterminalIdx != production.NonterminalIdx {
				// We are only interested in gotos which are happening in the nonterminal on the left hand side of
				// the production for our core.
				continue
			}
			b.itemLookaheadSets[stateIdx][itemIdx].Merge(&b.lalr1Builder.gotoRecords[gotoIdx].GotoFollows)
		}
	}
	return b.itemLookaheadSets[stateIdx][itemIdx]
}

// initInadequacyLists initializes inadequacyLists.
func (b *AnnotationsBuilder) initInadequacyLists() {
	b.inadequacyLists = make([]InadequacyManifestationDescriptions, len(b.parser.States))
	for stateIdx := range b.parser.States {
		allConflictContributions := b.getConflictContributions(stateIdx)
		terminalIdxs := slices.Collect(maps.Keys(allConflictContributions))
		slices.Sort(terminalIdxs)
		for _, terminalIdx := range terminalIdxs {
			conflictContributions := allConflictContributions[terminalIdx]
			if conflictContributions.Length() <= 1 {
				// We are not interested in states which do not have a conflict.
				continue
			}
			b.inadequacyLists[stateIdx] = append(b.inadequacyLists[stateIdx], InadequacyManifestationDescription{
				StateIdx:              stateIdx,
				TerminalIdx:           terminalIdx,
				ConflictContributions: conflictContributions,
			})
		}
	}
}

func (b *AnnotationsBuilder) getConflictContributions(stateIdx int) map[int]ConflictContributionSet {
	result := make(map[int]ConflictContributionSet)
	state := b.parser.States[stateIdx]
	for _, transition := range state.TransitionActions.All() {
		if transition.SymbolRef().IsNonterminal() {
			continue
		}
		contributions := result[transition.SymbolRef().Idx()]
		contributions.Add(NewShiftConflictContribution())
		result[transition.SymbolRef().Idx()] = contributions
	}
	for _, reduction := range state.ReduceActions.All() {
		for terminalIdx := range reduction.LookaheadSet.All() {
			contributions := result[terminalIdx]
			contributions.Add(NewReduceConflictContribution(reduction.ProductionIdx))
			result[terminalIdx] = contributions
		}
	}
	return result
}

// annotateManifestation is definition 3.30 of IELR(1).
func (b *AnnotationsBuilder) annotateManifestation(stateIdx int, inadequacyListIdx int) int {
	state := b.parser.States[stateIdx]
	inadequacyManifestationDescription := b.inadequacyLists[stateIdx][inadequacyListIdx]
	conflictContributions := inadequacyManifestationDescription.ConflictContributions
	newInadequacyAnnotation := InadequacyAnnotation{
		InadequacyListIdx:            inadequacyListIdx,
		InadequacyContributionMatrix: make(InadequacyContributionMatrix, conflictContributions.Length()),
	}
	for idx, contribution := range conflictContributions.All() {
		if contribution.IsShiftAction() {
			// The InadequacyContributionMatrix[idx] stays undefined
			continue
		}
		production := b.grammar.Productions[contribution.ProductionIdx()]
		if len(production.SymbolRefs) != 0 {
			kernelItem := backend.NewCore(contribution.ProductionIdx(), len(production.SymbolRefs))
			newInadequacyAnnotation.InadequacyContributionMatrix[idx] = make(KernelItemConflictContributions, state.KernelItems.Length())
			for j := range newInadequacyAnnotation.InadequacyContributionMatrix[idx] {
				newInadequacyAnnotation.InadequacyContributionMatrix[idx][j] = state.KernelItems.GetByIndex(j) == kernelItem
			}
		} else {
			newInadequacyAnnotation.InadequacyContributionMatrix[idx] = b.computeLhsContributions(
				stateIdx,
				production.NonterminalIdx,
				inadequacyManifestationDescription.TerminalIdx,
			)
		}
	}
	b.inadequacyAnnotations = append(b.inadequacyAnnotations, newInadequacyAnnotation)
	return len(b.inadequacyAnnotations) - 1
}

// computeLhsContributions is definition 3.31 of IELR(1).
func (b *AnnotationsBuilder) computeLhsContributions(stateIdx int, nonterminalIdx int, terminalIdx int) KernelItemConflictContributions {
	var gotoIdx int
	for _, idx := range b.lalr1Builder.gotoIdxsByStateIdx[stateIdx] {
		if b.lalr1Builder.gotoRecords[idx].NonterminalIdx != nonterminalIdx {
			continue
		}
		gotoIdx = idx
		break
	}
	if b.lalr1Builder.gotoRecords[gotoIdx].AlwaysFollows.Contains(terminalIdx) {
		return nil
	}

	state := b.parser.States[stateIdx]
	result := make(KernelItemConflictContributions, state.KernelItems.Length())
	for itemIdx := range result {
		itemLookaheadSet := b.getItemLookaheadSet(stateIdx, itemIdx)
		result[itemIdx] = b.followKernelItems[gotoIdx][itemIdx] && itemLookaheadSet.Contains(terminalIdx)
	}
	return result
}

// annotatePredecessor is definition 3.32 of IELR(1).
func (b *AnnotationsBuilder) annotatePredecessor(predecessorStateIdx int, successorStateIdx int, successorAnnotationIdx int) int {
	// TODO: continue implementation here

	return -1
}
