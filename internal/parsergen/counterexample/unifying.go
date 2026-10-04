package counterexample

import (
	"fmt"
	"slices"
	"time"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// The costs of the actions of the product parser. The search postpones production steps, forward and backward, as they
// branch and can repeat without end (section 5.4), and postpones a production step even more when its parser already
// holds the item it leads to, like A -> • A ... taken again. The paper gives no numbers, these are GoLR's own.
const (
	transitionCost             = 1
	reductionCost              = 1
	emptyDerivationCost        = 1
	productionStepCost         = 4
	repeatedProductionStepCost = 16
)

// unifyingBuilder searches the unifying counterexample of a pair of conflict items, see section 5. It simulates two
// copies of the parser, the first one reducing the reduce item and the second one shifting with the other item or
// reducing it, outward from the conflict state in a priority search over configurations (section 5.2).
//
// As in section 6, "Constructing unifying counterexamples", reverse transitions only enter states of the shortest
// lookahead-sensitive path, which keeps the search small but misses unifying counterexamples which need other states.
type unifyingBuilder struct {
	tables           *LookupTables
	conflictItemIdxs [2]int
	terminalIdx      int
	deadline         time.Time

	// configurationLimit limits the configurations removed from the queue, 0 for no limit.
	configurationLimit int

	// statistics counts the work of the search.
	statistics Statistics

	// nonunifying completes the derivations of the innermost nonterminal.
	nonunifying *nonunifyingBuilder

	// stage1StateIdxs holds the states of the path from where it enters the production of the reduce item, which are
	// the only ones reverse transitions enter during stage 1. pathStateIdxs holds all states of the path.
	stage1StateIdxs utils.Bitset
	pathStateIdxs   utils.Bitset

	queue ProductConfigurationQueue

	// visitedConfigHashes holds the hash of every ProductConfiguration the search queued, see ProductConfiguration.Hash.
	// Two configurations with the same hash count as the same, which could make the search miss a unifying counterexample,
	// but never return a wrong one.
	visitedConfigHashes map[uint64]struct{}

	// Scratch space for FIRST sets and the sequences of a ProductConfiguration.
	firstBuffer      [2]backend.LookaheadSet
	itemBuffer       []int
	derivationBuffer []Derivation

	// The items of the parsers of the ProductConfiguration whose successors are queued, the hashes of their prefixes
	// and the hashes of the parsers, see prepareHashes.
	parentItems        [2][]int
	parentPrefixHashes [2][]utils.Hash
	parentParserHashes [2]uint64

	// innermost is the nonunifying counterexample of the innermost nonterminal, once a ProductConfiguration showed it.
	innermost      Counterexample
	innermostFound bool
}

// newUnifyingBuilder returns a builder for the unifying counterexample of the pair of conflict items of the nonunifying
// builder, which gives up at the deadline or the configuration limit of the configuration.
func newUnifyingBuilder(nonunifying *nonunifyingBuilder, config Config, deadline time.Time) unifyingBuilder {
	tables := nonunifying.tables
	reducePath := nonunifying.reducePath
	result := unifyingBuilder{
		tables:              tables,
		conflictItemIdxs:    [2]int{nonunifying.reduceItemIdx, nonunifying.otherItemIdx},
		terminalIdx:         nonunifying.terminalIdx,
		deadline:            deadline,
		configurationLimit:  config.ConfigurationLimit,
		nonunifying:         nonunifying,
		queue:               NewProductConfigurationQueue(),
		visitedConfigHashes: make(map[uint64]struct{}),
	}
	stage1From := 0
	for pathIdx, item := range reducePath {
		if item.productionStep {
			stage1From = pathIdx
		}
	}
	for pathIdx, item := range reducePath {
		stateIdx := tables.StateIdx(item.itemIdx)
		result.pathStateIdxs.Add(stateIdx)
		if pathIdx >= stage1From {
			result.stage1StateIdxs.Add(stateIdx)
		}
	}
	return result
}

// Build returns the unifying counterexample. When the search runs out of configurations, time or its configuration
// limit without one, it returns the nonunifying counterexample of the innermost nonterminal instead, if the search got
// far enough to find it, and reports false when it did not.
func (b *unifyingBuilder) Build() (Counterexample, bool) {
	b.addInitial()
	for processed := 0; !b.queue.IsEmpty() && !b.isLimitReached(processed); processed++ {
		pending := b.queue.Remove()
		productConfiguration := b.build(&pending)
		b.statistics.Processed++
		if result, found := b.unifyingCounterexample(productConfiguration); found {
			return result, true
		}
		if !b.innermostFound {
			b.findInnermost(productConfiguration)
		}
		b.addSuccessors(productConfiguration)
	}
	return b.innermost, b.innermostFound
}

// edit returns the change the action of the PendingConfiguration makes to the parser of its parent.
func (b *unifyingBuilder) edit(p *PendingConfiguration, parserIdx int) parserEdit {
	parser := &p.parent.Parsers[parserIdx]
	result := parserEdit{
		changed:   true,
		appended:  noItemIdx,
		prepended: noItemIdx,
		depth:     parser.depth,
	}
	onParser := p.parserIdx == parserIdx
	switch p.action {
	case initialAction:
		result.changed = false
	case transitionAction:
		result.appended = p.itemIdxs[parserIdx]
	case productionStepAction:
		result.changed = onParser
		result.appended = p.itemIdxs[parserIdx]
		if !parser.StageCompleted() {
			result.depth++
		}
	case emptyDerivationAction:
		result.changed = onParser
		result.appended = p.itemIdxs[parserIdx]
	case reductionAction:
		result.changed = onParser
		result.dropped = b.tables.Core(parser.Tail()).Position() + 1
		result.appended = p.itemIdxs[parserIdx]
		if parser.depth == 0 {
			result.depth = completedDepth
		} else if !parser.StageCompleted() {
			result.depth--
		}
	case reverseTransitionAction:
		result.prepended = p.itemIdxs[parserIdx]
	case reverseProductionStepAction:
		result.changed = onParser
		result.prepended = p.itemIdxs[parserIdx]
	}
	return result
}

// build returns the ProductConfiguration of the PendingConfiguration. It applies the action to the parent like the
// unifyingBuilder would have when it queued the configuration, which checked that the action is possible.
func (b *unifyingBuilder) build(p *PendingConfiguration) *ProductConfiguration {
	if p.action == initialAction {
		return p.parent
	}
	result := p.parent.Successor(p.Cost - p.parent.Cost)
	result.TerminalShifted = p.parent.TerminalShifted || p.action == transitionAction
	for parserIdx := range result.Parsers {
		edit := b.edit(p, parserIdx)
		if !edit.changed {
			continue
		}
		parser := &result.Parsers[parserIdx]
		parser.derivations = b.derivationsAfter(p, parserIdx)
		if edit.dropped > 0 {
			parser.items = parser.items.DropBack(edit.dropped)
		}
		if edit.appended != noItemIdx {
			parser.items = parser.items.PushBack(edit.appended)
		}
		if edit.prepended != noItemIdx {
			parser.items = parser.items.PushFront(edit.prepended)
		}
		parser.depth = edit.depth
	}
	return &result
}

// derivationsAfter returns the derivations of the parser after the action of the PendingConfiguration: with a leaf of
// the symbol of a transition, the derivation of the empty string, or the derivation of the nonterminal a reduction
// reduces to.
func (b *unifyingBuilder) derivationsAfter(p *PendingConfiguration, parserIdx int) utils.Deque[Derivation] {
	parser := &p.parent.Parsers[parserIdx]
	switch p.action {
	case initialAction, productionStepAction, reverseProductionStepAction:
		return parser.derivations
	case transitionAction:
		symbolRef, _ := b.tables.NextSymbol(p.parent.Parsers[0].Tail())
		return parser.derivations.PushBack(NewLeafDerivation(symbolRef))
	case emptyDerivationAction:
		return parser.derivations.PushBack(b.emptyDerivation(p.parent, parserIdx))
	case reductionAction:
		core := b.tables.Core(parser.Tail())
		derivations, children := parser.derivations.PopBack(
			core.Position(),
			make([]Derivation, 0, core.Position()+1),
		)
		if parser.depth == 0 {
			dotPosition := b.tables.Core(b.conflictItemIdxs[parserIdx]).Position()
			children = slices.Insert(children, dotPosition, NewDotDerivation())
		}
		return derivations.PushBack(NewExpandedDerivation(b.tables.grammar, core.ProductionIdx(), children))
	case reverseTransitionAction:
		core := b.tables.Core(p.parent.Parsers[0].Head())
		symbolRef := b.tables.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position()-1]
		return parser.derivations.PushFront(NewLeafDerivation(symbolRef))
	}
	return parser.derivations
}

// isLimitReached reports if the search has to give up after processing the number of configurations.
func (b *unifyingBuilder) isLimitReached(processed int) bool {
	return (b.configurationLimit > 0 && processed >= b.configurationLimit) || !time.Now().Before(b.deadline)
}

// unifyingCounterexample returns the unifying counterexample when the ProductConfiguration completes the search, see
// section 5.4: both parsers completed their stage, and both item sequences are of the form
// [? -> ... • A ..., ? -> ... A • ...] for the same nonterminal A. The single derivation of A of either parser are the
// two derivations.
func (b *unifyingBuilder) unifyingCounterexample(c *ProductConfiguration) (Counterexample, bool) {
	for _, parser := range c.Parsers {
		if !parser.StageCompleted() || parser.items.Len() != 2 || parser.derivations.Len() != 1 {
			return Counterexample{}, false
		}
	}
	derivations := [2]Derivation{
		c.Parsers[0].derivations.Last(),
		c.Parsers[1].derivations.Last(),
	}
	if !derivations[0].IsExpanded() || derivations[0].Symbol != derivations[1].Symbol {
		return Counterexample{}, false
	}
	return Counterexample{
		Unifying:       true,
		NonterminalIdx: derivations[0].Symbol.Idx(),
		Derivations:    derivations,
	}, true
}

// findInnermost keeps the nonunifying counterexample of the innermost nonterminal, when the ProductConfiguration shows
// it. That is stage 3 (p. 5): both parsers completed their stage, and their first items belong to productions of the
// same nonterminal with the same symbols in front of the dot. Completing both item sequences up to these productions
// gives two derivations of the nonterminal, which share the symbols up to the dot.
//
// While the conflict terminal is not shifted yet, which only happens for two reduce items, the completion has to place
// it directly behind the dot within these productions, or the ProductConfiguration does not show the innermost
// nonterminal.
func (b *unifyingBuilder) findInnermost(c *ProductConfiguration) {
	if !c.Parsers[0].StageCompleted() || !c.Parsers[1].StageCompleted() {
		return
	}
	cores := [2]backend.Core{
		b.tables.Core(c.Parsers[0].Head()),
		b.tables.Core(c.Parsers[1].Head()),
	}
	productions := [2]frontend.Production{
		b.tables.grammar.Productions[cores[0].ProductionIdx()],
		b.tables.grammar.Productions[cores[1].ProductionIdx()],
	}
	if productions[0].NonterminalIdx != productions[1].NonterminalIdx ||
		cores[0].Position() != cores[1].Position() ||
		!slices.Equal(productions[0].SymbolRefs[:cores[0].Position()], productions[1].SymbolRefs[:cores[1].Position()]) {
		return
	}

	var derivations [2]Derivation
	for parserIdx := range derivations {
		derivation, completed := b.completeDerivation(c, parserIdx)
		if !completed {
			return
		}
		derivations[parserIdx] = derivation
	}
	b.innermost = Counterexample{
		NonterminalIdx: productions[0].NonterminalIdx,
		Derivations:    derivations,
	}
	b.innermostFound = true
}

// completeDerivation completes every production of the item sequence of the parser, from the innermost one outward, up
// to the production of its first item. The symbols in front of the dot of the first item, and those behind the dot or
// the expanded nonterminal of every production, stay leaves, see nonunifyingBuilder.appendRest. It reports false when
// the conflict terminal is still required behind the dot, or cannot directly follow it.
func (b *unifyingBuilder) completeDerivation(
	productConfiguration *ProductConfiguration,
	parserIdx int,
) (Derivation, bool) {
	parser := &productConfiguration.Parsers[parserIdx]
	b.itemBuffer = parser.items.AppendAll(b.itemBuffer[:0])
	itemIdxs := b.itemBuffer
	b.derivationBuffer = parser.derivations.AppendAll(b.derivationBuffer[:0])
	derivations := b.derivationBuffer
	required := !productConfiguration.TerminalShifted

	var result Derivation
	productionTo := len(itemIdxs)
	derivationTo := len(derivations)
	for productionFrom, itemIdx := range slices.Backward(itemIdxs) {
		// A production step enters an item with the dot at the start, a transition one with the dot further on.
		firstCore := b.tables.Core(itemIdx)
		if productionFrom > 0 && firstCore.Position() != 0 {
			continue
		}
		restItemIdx := itemIdxs[productionTo-1]
		symbolRefs := b.tables.grammar.Productions[firstCore.ProductionIdx()].SymbolRefs
		transitionCount := productionTo - 1 - productionFrom

		children := make([]Derivation, 0, len(symbolRefs)+1)
		for _, symbolRef := range symbolRefs[:firstCore.Position()] {
			children = append(children, NewLeafDerivation(symbolRef))
		}
		children = append(children, derivations[derivationTo-transitionCount:derivationTo]...)
		derivationTo -= transitionCount

		if productionTo < len(itemIdxs) {
			children = append(children, result)
			restItemIdx, _ = b.tables.Transition(restItemIdx)
		}
		var possible bool
		children, required, possible = b.nonunifying.appendRest(children, restItemIdx, required)
		if !possible {
			return Derivation{}, false
		}
		result = NewExpandedDerivation(b.tables.grammar, firstCore.ProductionIdx(), children)
		productionTo = productionFrom
	}
	return result, !required
}

// addSuccessors adds the successor configurations of the ProductConfiguration to the queue, see section 5.3. A parser
// whose last item is a reduce item reduces, or is prepended until it has the items to reduce. Only when neither parser
// reduces, the parsers move forward by a transition together or by a production step of either one.
func (b *unifyingBuilder) addSuccessors(c *ProductConfiguration) {
	b.prepareHashes(c)
	reducing := false
	for parserIdx := range c.Parsers {
		if _, found := b.tables.NextSymbol(c.Parsers[parserIdx].Tail()); found {
			continue
		}
		reducing = true
		if b.canReduce(c, parserIdx) {
			b.addReduction(c, parserIdx)
		} else {
			b.addPrepended(c, parserIdx)
		}
	}
	if !reducing {
		b.addTransition(c)
		for parserIdx := range c.Parsers {
			b.addProductionSteps(c, parserIdx)
		}
	}
	for parserIdx := range c.Parsers {
		b.addEmptyDerivation(c, parserIdx)
	}
}

// addInitial adds the initial configuration of figure 8(b) to the queue.
func (b *unifyingBuilder) addInitial() {
	initial := &ProductConfiguration{
		Parsers: [2]SimulatedParser{
			NewSimulatedParser(b.conflictItemIdxs[0]),
			NewSimulatedParser(b.conflictItemIdxs[1]),
		},
	}
	hash := initial.Hash()
	b.addWithHash(PendingConfiguration{parent: initial, action: initialAction}, hash)
}

// add adds the PendingConfiguration to the queue. Its parent must be the configuration prepareHashes was called with
// last.
func (b *unifyingBuilder) add(p PendingConfiguration) {
	hash := b.pendingHash(&p)
	utils.DebugAssert(func() error {
		c := b.build(&p)
		if builtHash := c.Hash(); builtHash != hash {
			return fmt.Errorf("action %d hashed to %x without building and to %x when built", p.action, hash, builtHash)
		}
		// The invariant of section 5.4.
		heads := [2]int{c.Parsers[0].Head(), c.Parsers[1].Head()}
		if b.tables.StateIdx(heads[0]) != b.tables.StateIdx(heads[1]) {
			return fmt.Errorf("the first items %d and %d are in different states", heads[0], heads[1])
		}
		return nil
	})
	b.addWithHash(p, hash)
}

// addWithHash adds the PendingConfiguration with the hash of its ProductConfiguration to the queue.
//
// A configuration queued before is dropped, even when it comes at a lower cost, so the queue never holds the same
// configuration twice. The queue hands out configurations by cost, so a later duplicate is cheaper by at most the cost
// of one repeated production step, while the costs of the searches on large grammars reach thousands. The costs only
// lead the search to cheaper counterexamples first, they do not guarantee the cheapest one.
func (b *unifyingBuilder) addWithHash(p PendingConfiguration, hash uint64) {
	b.statistics.Generated++
	if _, found := b.visitedConfigHashes[hash]; found {
		return
	}
	b.visitedConfigHashes[hash] = struct{}{}
	b.queue.Add(p)
	b.statistics.Queued++
	b.statistics.PeakQueueLength = max(b.statistics.PeakQueueLength, b.queue.Len())
}

// prepareHashes collects the items of both parsers of the ProductConfiguration and the hashes of their prefixes, from
// which pendingHash computes the hashes of its successors.
func (b *unifyingBuilder) prepareHashes(c *ProductConfiguration) {
	for parserIdx := range c.Parsers {
		parser := &c.Parsers[parserIdx]
		b.parentItems[parserIdx] = parser.items.AppendAll(b.parentItems[parserIdx][:0])
		itemsHash := utils.NewHash()
		b.parentPrefixHashes[parserIdx] = append(b.parentPrefixHashes[parserIdx][:0], itemsHash)
		for _, itemIdx := range b.parentItems[parserIdx] {
			utils.WriteHash(&itemsHash, itemIdx)
			b.parentPrefixHashes[parserIdx] = append(b.parentPrefixHashes[parserIdx], itemsHash)
		}
		b.parentParserHashes[parserIdx] = hashParser(itemsHash, parser.depth)
	}
}

// pendingHash returns the hash of the ProductConfiguration of the PendingConfiguration without building it, see
// ProductConfiguration.Hash. The parent must be the configuration prepareHashes was called with last. Appending an item
// continues the hash of the items in front of it, only prepending an item hashes all items again.
func (b *unifyingBuilder) pendingHash(p *PendingConfiguration) uint64 {
	parserHashes := b.parentParserHashes
	for parserIdx := range parserHashes {
		edit := b.edit(p, parserIdx)
		if !edit.changed {
			continue
		}
		items := b.parentItems[parserIdx]
		var itemsHash utils.Hash
		if edit.prepended != noItemIdx {
			itemsHash = utils.NewHash()
			utils.WriteHash(&itemsHash, edit.prepended)
			utils.WriteHashSlice(&itemsHash, items)
		} else {
			itemsHash = b.parentPrefixHashes[parserIdx][len(items)-edit.dropped]
		}
		if edit.appended != noItemIdx {
			utils.WriteHash(&itemsHash, edit.appended)
		}
		parserHashes[parserIdx] = hashParser(itemsHash, edit.depth)
	}
	return hashConfiguration(parserHashes, p.parent.TerminalShifted || p.action == transitionAction)
}

// canReduce reports if the parser holds the items of the production of its last item, which is a reduce item of the
// form A -> X1 ... Xl •, and the item in front of them with the dot before A, see figure 10(f).
func (b *unifyingBuilder) canReduce(c *ProductConfiguration, parserIdx int) bool {
	parser := &c.Parsers[parserIdx]
	return parser.items.Len() > b.tables.Core(parser.Tail()).Position()+1
}

// addTransition adds the transition of both parsers on the symbol after the dot of their last items, when it is the
// same symbol, see figure 10(a). It becomes a leaf of both derivations. Until the parsers shifted the conflict
// terminal, the symbol has to be that terminal: a nonterminal beginning with it is entered by production steps instead,
// so the derivations show the terminal directly behind the dot.
func (b *unifyingBuilder) addTransition(c *ProductConfiguration) {
	symbolRef, _ := b.tables.NextSymbol(c.Parsers[0].Tail())
	if otherSymbolRef, _ := b.tables.NextSymbol(c.Parsers[1].Tail()); otherSymbolRef != symbolRef {
		return
	}
	if !c.TerminalShifted && symbolRef != frontend.NewTerminalRef(b.terminalIdx) {
		return
	}

	successor := PendingConfiguration{
		Cost:   c.Cost + transitionCost,
		parent: c,
		action: transitionAction,
	}
	for parserIdx := range c.Parsers {
		tailItemIdx := c.Parsers[parserIdx].Tail()
		targetItemIdx, found := b.tables.Transition(tailItemIdx)
		if !found || !b.isShiftAllowed(b.tables.StateIdx(tailItemIdx), symbolRef) {
			return
		}
		successor.itemIdxs[parserIdx] = targetItemIdx
	}
	b.add(successor)
}

// addProductionSteps adds the production steps of the parser, see figure 10(b). A production with an empty right hand
// side is not entered, addEmptyDerivation passes over its nonterminal instead. The first symbol of the production has
// to be able to begin like the symbol after the dot of the other parser, and like the conflict terminal until it is
// shifted.
func (b *unifyingBuilder) addProductionSteps(c *ProductConfiguration, parserIdx int) {
	parser := &c.Parsers[parserIdx]
	from, to := b.tables.ProductionSteps(parser.Tail())
	otherSymbolRef, _ := b.tables.NextSymbol(c.Parsers[1-parserIdx].Tail())
	for itemIdx := from; itemIdx < to; itemIdx++ {
		symbolRef, found := b.tables.NextSymbol(itemIdx)
		if !found || !b.canBeginAlike(symbolRef, otherSymbolRef) {
			continue
		}
		if !c.TerminalShifted && !b.canBeginAlike(symbolRef, frontend.NewTerminalRef(b.terminalIdx)) {
			continue
		}

		cost := productionStepCost
		if utils.DequeContains(parser.items, itemIdx) {
			cost += repeatedProductionStepCost
		}
		b.add(c.pendingSuccessor(cost, productionStepAction, parserIdx, itemIdx))
	}
}

// addEmptyDerivation passes the parser over the nonterminal after the dot of its last item, when it can vanish, with a
// shortest derivation of the empty string. This takes the place of production steps into empty productions, which
// would have to be reduced right away. The paper leaves empty productions open.
func (b *unifyingBuilder) addEmptyDerivation(c *ProductConfiguration, parserIdx int) {
	parser := &c.Parsers[parserIdx]
	symbolRef, found := b.tables.NextSymbol(parser.Tail())
	if !found || symbolRef.IsTerminal() || !b.tables.firstSets.IsNullable(symbolRef.Idx()) {
		return
	}
	targetItemIdx, found := b.tables.Transition(parser.Tail())
	if !found {
		return
	}
	if !b.isEmptyDerivationAllowed(c, parserIdx) {
		return
	}
	b.add(c.pendingSuccessor(emptyDerivationCost, emptyDerivationAction, parserIdx, targetItemIdx))
}

// isEmptyDerivationAllowed reports if the nonterminal after the dot of the last item of the parser has a derivation of
// the empty string, whose reductions the declarations allow on the terminal which follows, see emptyDerivation.
func (b *unifyingBuilder) isEmptyDerivationAllowed(c *ProductConfiguration, parserIdx int) bool {
	reach, known := b.followingTerminalReach(c, parserIdx)
	return !known || reach.CanVanish(c.Parsers[parserIdx].Tail())
}

// emptyDerivation returns a shortest derivation of the empty string from the nonterminal after the dot of the last item
// of the parser, whose reductions the declarations allow on the terminal which follows: the conflict terminal until it
// is shifted, and the symbol after the dot of the other parser then, when it is a terminal. There has to be one, see
// isEmptyDerivationAllowed. When the terminal which follows is not known yet, the derivation is the shortest one of the
// grammar, like isReductionAllowed allows a reduction then.
func (b *unifyingBuilder) emptyDerivation(c *ProductConfiguration, parserIdx int) Derivation {
	itemIdx := c.Parsers[parserIdx].Tail()
	reach, known := b.followingTerminalReach(c, parserIdx)
	if !known {
		nonterminalRef, _ := b.tables.NextSymbol(itemIdx)
		return b.tables.shortestEmptyDerivation(nonterminalRef.Idx())
	}
	return b.tables.emptyDerivation(reach, itemIdx)
}

// followingTerminalReach returns the reach of the terminal which follows a derivation of the empty string of the
// parser, see emptyDerivation, and reports false when that terminal is not known yet.
func (b *unifyingBuilder) followingTerminalReach(c *ProductConfiguration, parserIdx int) (*terminalReach, bool) {
	terminalIdx := b.terminalIdx
	if c.TerminalShifted {
		symbolRef, found := b.tables.NextSymbol(c.Parsers[1-parserIdx].Tail())
		if !found || !symbolRef.IsTerminal() {
			return nil, false
		}
		terminalIdx = symbolRef.Idx()
	}
	return b.tables.terminalReach(terminalIdx), true
}

// addReduction adds the reduction of the parser, see figure 10(f). It removes the items of the production and appends
// the goto of the item in front of them, and the derivations of the production become one derivation of its
// nonterminal. The reduction of the production of the conflict item puts the dot at the position of that item.
//
// The symbol after the dot of the other parser, and the conflict terminal until it is shifted, have to be in the
// lookahead set of the reduce item, a nonterminal through its FIRST set.
func (b *unifyingBuilder) addReduction(c *ProductConfiguration, parserIdx int) {
	if !b.isReductionAllowed(c, parserIdx) {
		return
	}
	items := b.parentItems[parserIdx]
	core := b.tables.Core(items[len(items)-1])
	parentItemIdx := items[len(items)-core.Position()-2]
	gotoItemIdx, found := b.tables.Transition(parentItemIdx)
	utils.DebugAssert(func() error {
		if !found {
			return fmt.Errorf("item %d has no goto", parentItemIdx)
		}
		return nil
	})
	b.add(c.pendingSuccessor(reductionCost, reductionAction, parserIdx, gotoItemIdx))
}

// isReductionAllowed reports if the symbol after the dot of the other parser, and the conflict terminal until it is
// shifted, allow the reduction of the last item of the parser. A nonterminal which can vanish allows it, as the symbols
// behind it are not known yet.
func (b *unifyingBuilder) isReductionAllowed(c *ProductConfiguration, parserIdx int) bool {
	reduceItemIdx := c.Parsers[parserIdx].Tail()
	stateIdx := b.tables.StateIdx(reduceItemIdx)
	contribution := conflict.NewReduceContribution(b.tables.Core(reduceItemIdx).ProductionIdx())
	if !c.TerminalShifted && !b.tables.IsActionAllowed(stateIdx, b.terminalIdx, contribution) {
		return false
	}

	symbolRef, found := b.tables.NextSymbol(c.Parsers[1-parserIdx].Tail())
	switch {
	case !found:
		return true
	case symbolRef.IsTerminal():
		return b.tables.IsActionAllowed(stateIdx, symbolRef.Idx(), contribution)
	case b.tables.firstSets.IsNullable(symbolRef.Idx()):
		return true
	default:
		first := b.first(0, []frontend.SymbolRef{symbolRef})
		first.Intersect(b.tables.Lookahead(reduceItemIdx))
		return !first.IsEmpty()
	}
}

// addPrepended adds the configurations which prepend items to prepare the reduction of the parser, which does not hold
// the items of its production yet, see figure 10(c) to (e). When its first item has the dot at the start, it needs a
// reverse production step. Otherwise both parsers need a reverse transition together, which the other parser first
// prepares by a reverse production step when its first item has the dot at the start.
func (b *unifyingBuilder) addPrepended(c *ProductConfiguration, parserIdx int) {
	if b.tables.Core(c.Parsers[parserIdx].Head()).Position() == 0 {
		b.addReverseProductionSteps(c, parserIdx)
		return
	}
	if b.tables.Core(c.Parsers[1-parserIdx].Head()).Position() == 0 {
		b.addReverseProductionSteps(c, 1-parserIdx)
		return
	}
	b.addReverseTransitions(c)
}

// addReverseTransitions adds the reverse transitions of both parsers into the same state, see figure 10(c). The symbol
// of the transitions becomes a leaf of both derivations. The state has to be on the shortest lookahead-sensitive path,
// see unifyingBuilder. During stage 1, the item prepended to the first parser has to hold the conflict terminal in its
// lookahead set (p. 6).
func (b *unifyingBuilder) addReverseTransitions(c *ProductConfiguration) {
	core := b.tables.Core(c.Parsers[0].Head())
	symbolRef := b.tables.grammar.Productions[core.ProductionIdx()].SymbolRefs[core.Position()-1]
	for _, predecessorItemIdx := range b.tables.ReverseTransitions(c.Parsers[0].Head()) {
		stateIdx := b.tables.StateIdx(predecessorItemIdx)
		if !b.isPrependedStateAllowed(c, stateIdx) || !b.isShiftAllowed(stateIdx, symbolRef) {
			continue
		}
		if !c.Parsers[0].StageCompleted() && !b.tables.Lookahead(predecessorItemIdx).Contains(b.terminalIdx) {
			continue
		}
		for _, otherPredecessorItemIdx := range b.tables.ReverseTransitions(c.Parsers[1].Head()) {
			if b.tables.StateIdx(otherPredecessorItemIdx) != stateIdx {
				continue
			}
			b.add(PendingConfiguration{
				Cost:     c.Cost + transitionCost,
				parent:   c,
				action:   reverseTransitionAction,
				itemIdxs: [2]int{predecessorItemIdx, otherPredecessorItemIdx},
			})
		}
	}
}

// isPrependedStateAllowed reports if a reverse transition may enter the state, see unifyingBuilder.
func (b *unifyingBuilder) isPrependedStateAllowed(c *ProductConfiguration, stateIdx int) bool {
	if !c.Parsers[0].StageCompleted() {
		return b.stage1StateIdxs.Contains(stateIdx)
	}
	return b.pathStateIdxs.Contains(stateIdx)
}

// addReverseProductionSteps adds the reverse production steps of the parser, see figure 10(d) and (e). When the parser
// holds nothing but the items of the production it reduces next, the item prepended becomes its last one after the
// reduction, so the symbols behind its nonterminal have to be able to begin like the symbol after the dot of the other
// parser, and like the conflict terminal until it is shifted, or have to vanish.
func (b *unifyingBuilder) addReverseProductionSteps(c *ProductConfiguration, parserIdx int) {
	parser := &c.Parsers[parserIdx]
	_, hasNextSymbol := b.tables.NextSymbol(parser.Tail())
	reducesNext := !hasNextSymbol && parser.items.Len() == b.tables.Core(parser.Tail()).Position()+1
	otherSymbolRef, otherFound := b.tables.NextSymbol(c.Parsers[1-parserIdx].Tail())
	for _, parentItemIdx := range b.tables.ReverseProductionSteps(parser.Head()) {
		if reducesNext {
			rest := b.tables.restBehindNextSymbol(parentItemIdx)
			if otherFound && !b.canSequenceBeginAlike(rest, otherSymbolRef) {
				continue
			}
			if !c.TerminalShifted && !b.canSequenceBeginAlike(rest, frontend.NewTerminalRef(b.terminalIdx)) {
				continue
			}
		}

		cost := productionStepCost
		if utils.DequeContains(parser.items, parentItemIdx) {
			cost += repeatedProductionStepCost
		}
		b.add(c.pendingSuccessor(cost, reverseProductionStepAction, parserIdx, parentItemIdx))
	}
}

// isShiftAllowed reports if the action filter allows the transition of the state on the symbol, see
// LookupTables.IsActionAllowed. A transition on a nonterminal is always allowed.
func (b *unifyingBuilder) isShiftAllowed(stateIdx int, symbolRef frontend.SymbolRef) bool {
	return symbolRef.IsNonterminal() ||
		b.tables.IsActionAllowed(stateIdx, symbolRef.Idx(), conflict.NewShiftContribution())
}

// canBeginAlike reports if the derivations of two symbols can begin with the same terminal. A nonterminal which can
// vanish can be passed over, so it begins alike with every symbol.
//
// The paper leaves open how production steps are restricted. Without a restriction a production step could enter every
// production of the nonterminal, though only those whose first symbol can begin like the other parser can lead to a
// transition of both parsers together.
func (b *unifyingBuilder) canBeginAlike(symbolRef frontend.SymbolRef, otherSymbolRef frontend.SymbolRef) bool {
	return b.canSequenceBeginAlike([]frontend.SymbolRef{symbolRef}, otherSymbolRef)
}

// canSequenceBeginAlike reports if the derivations of the sequence and the symbol can begin with the same terminal,
// see canBeginAlike. A sequence which can vanish begins alike with every symbol.
func (b *unifyingBuilder) canSequenceBeginAlike(symbolRefs []frontend.SymbolRef, symbolRef frontend.SymbolRef) bool {
	if symbolRef.IsNonterminal() && b.tables.firstSets.IsNullable(symbolRef.Idx()) {
		return true
	}
	if b.tables.firstSets.IsSequenceNullable(symbolRefs) {
		return true
	}
	first := b.first(0, symbolRefs)
	if symbolRef.IsTerminal() {
		return first.Contains(symbolRef.Idx())
	}
	otherFirst := b.first(1, []frontend.SymbolRef{symbolRef})
	first.Intersect(otherFirst)
	return !first.IsEmpty()
}

// first returns the FIRST set of the sequence in one of the two buffers of the builder, which stays valid until the
// buffer is used again.
func (b *unifyingBuilder) first(bufferIdx int, symbolRefs []frontend.SymbolRef) *backend.LookaheadSet {
	first := &b.firstBuffer[bufferIdx]
	first.Clear()
	b.tables.firstSets.FirstOfSequence(symbolRefs, first)
	return first
}
