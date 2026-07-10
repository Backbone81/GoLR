package oracle

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// DiffLALR1ParserKernelItems reports which states one parser table has and the other does not, matching states up by
// their kernel items. It looks only at the set of states, ignoring the transitions and reduce actions the states hold.
//
// An empty result means both parser tables have the same set of states.
func DiffLALR1ParserKernelItems(want backend.Parser, got backend.Parser) []string {
	differ := NewLALR1ParserDiffer(want, got)
	differ.diffKernelItems()
	return differ.differences
}

// DiffLALR1ParserStates reports how the two parser tables differ, comparing the full states: their transitions and
// reduce actions on top of the set of states DiffLALR1ParserKernelItems compares. States are matched up by their kernel
// items, so a different numbering of the states of the two parser tables does not show up as a difference. Both parser
// tables are expected to have been generated from the same grammar.
//
// An empty result means both parser tables describe the same parser.
func DiffLALR1ParserStates(want backend.Parser, got backend.Parser) []string {
	differ := NewLALR1ParserDiffer(want, got)
	differ.diffKernelItems()
	differ.diffStates()
	return differ.differences
}

// LALR1ParserDiffer compares two parser tables and collects the differences it finds.
//
// A state is identified by the raw bytes of its kernel items rather than by its state index, because two parser table
// constructions which agree on the automaton are free to number their states differently. An LALR(1) state is uniquely
// determined by its kernel items, which makes them a usable identity. The bytes are exact and collision-free, but not
// readable, so a reported difference renders the readable kernel items of a state on demand, see wantLabel and gotLabel.
//
// The states of both parser tables are indexed once up front. Every wantXxx field has a gotXxx counterpart holding the
// same derived data for the other parser table.
type LALR1ParserDiffer struct {
	// wantParser is the parser table which is taken to be correct, and gotParser is the parser table under test.
	wantParser backend.Parser
	gotParser  backend.Parser

	// wantStateIdxByKey and gotStateIdxByKey map the kernel item bytes of a state to its state index.
	wantStateIdxByKey map[string]int
	gotStateIdxByKey  map[string]int

	// wantKeyByStateIdx and gotKeyByStateIdx map a state index to the kernel item bytes of that state. This resolves
	// the destination of a transition, which the parser table names by state index, into the identity of the
	// destination state.
	wantKeyByStateIdx []string
	gotKeyByStateIdx  []string

	// differences holds every difference found so far.
	differences []string
}

// NewLALR1ParserDiffer returns a LALR1ParserDiffer for the two parser tables, with the states of both indexed by their kernel
// items.
func NewLALR1ParserDiffer(want backend.Parser, got backend.Parser) LALR1ParserDiffer {
	result := LALR1ParserDiffer{
		wantParser: want,
		gotParser:  got,
	}
	result.wantStateIdxByKey, result.wantKeyByStateIdx = result.indexStates("want", want)
	result.gotStateIdxByKey, result.gotKeyByStateIdx = result.indexStates("got", got)
	return result
}

// wantLabel and gotLabel render the identity of a state of the respective parser table back into its readable kernel
// items, so a reported difference can name the state it is talking about.
func (d *LALR1ParserDiffer) wantLabel(key string) string {
	return d.wantParser.States[d.wantStateIdxByKey[key]].KernelItems.String()
}

func (d *LALR1ParserDiffer) gotLabel(key string) string {
	return d.gotParser.States[d.gotStateIdxByKey[key]].KernelItems.String()
}

// report records a difference between the two parser tables.
func (d *LALR1ParserDiffer) report(format string, args ...any) {
	d.differences = append(d.differences, fmt.Sprintf(format, args...))
}

// indexStates indexes the states of the parser table by their kernel items, returning the two lookups the comparison
// works with: the state index by identity and the identity by state index. The label names the parser table in the
// differences which are reported.
//
// Two states with the same kernel items are reported as a difference, because a parser table which holds such a pair of
// states cannot be an LALR(1) parser table to begin with. The comparison continues with the first of the two states, so
// that the remaining differences still get reported.
func (d *LALR1ParserDiffer) indexStates(label string, parser backend.Parser) (map[string]int, []string) {
	stateIdxByKey := make(map[string]int, len(parser.States))
	keyByStateIdx := make([]string, len(parser.States))

	for stateIdx := range parser.States {
		kernelItems := parser.States[stateIdx].KernelItems
		// The core set is ordered, so the raw bytes of its values are a stable and collision-free identity.
		key := string(kernelItems.Bytes())
		keyByStateIdx[stateIdx] = key

		if earlierStateIdx, exists := stateIdxByKey[key]; exists {
			d.report(
				"%s: state %d repeats the kernel items %s of state %d, which no LALR(1) parser table can do",
				label, stateIdx, kernelItems.String(), earlierStateIdx,
			)
			continue
		}
		stateIdxByKey[key] = stateIdx
	}
	return stateIdxByKey, keyByStateIdx
}

// sortedKeys returns the identities of the indexed states in a stable order, so that reported differences come out in
// the same order every time. The order follows the raw identity bytes and carries no further meaning.
func sortedKeys(stateIdxByKey map[string]int) []string {
	result := slices.Collect(maps.Keys(stateIdxByKey))
	slices.Sort(result)
	return result
}

// diffKernelItems reports which states one parser table has and the other does not.
func (d *LALR1ParserDiffer) diffKernelItems() {
	for _, key := range sortedKeys(d.wantStateIdxByKey) {
		if _, exists := d.gotStateIdxByKey[key]; !exists {
			d.report("state %s is missing", d.wantLabel(key))
		}
	}
	for _, key := range sortedKeys(d.gotStateIdxByKey) {
		if _, exists := d.wantStateIdxByKey[key]; !exists {
			d.report("state %s is unexpected", d.gotLabel(key))
		}
	}
}

// diffStates reports how the states the two parser tables have in common differ.
func (d *LALR1ParserDiffer) diffStates() {
	for _, key := range sortedKeys(d.wantStateIdxByKey) {
		gotStateIdx, exists := d.gotStateIdxByKey[key]
		if !exists {
			// Already reported by diffKernelItems. Reporting every transition and reduce action of the missing state on
			// top of that would only bury the difference which matters.
			continue
		}
		wantState := &d.wantParser.States[d.wantStateIdxByKey[key]]
		gotState := &d.gotParser.States[gotStateIdx]

		d.diffTransitions(key, wantState, gotState)
		d.diffReduceActions(key, wantState, gotState)
	}
}

// diffTransitions reports how the transitions of two states with the same kernel items differ. A transition is compared
// by the kernel items of the state it leads to, never by the index of that state. The state label names the state the
// two transitions belong to.
func (d *LALR1ParserDiffer) diffTransitions(key string, wantState *backend.State, gotState *backend.State) {
	wantTransitions := transitions(wantState, d.wantKeyByStateIdx)
	gotTransitions := transitions(gotState, d.gotKeyByStateIdx)

	for _, symbolRef := range slices.Sorted(maps.Keys(wantTransitions)) {
		switch gotDestinationKey, exists := gotTransitions[symbolRef]; {
		case !exists:
			d.report(
				"state %s: missing transition on %s to %s",
				d.wantLabel(key), d.symbolString(symbolRef), d.wantLabel(wantTransitions[symbolRef]),
			)
		case gotDestinationKey != wantTransitions[symbolRef]:
			d.report(
				"state %s: transition on %s leads to %s, want %s",
				d.wantLabel(key), d.symbolString(symbolRef),
				d.gotLabel(gotDestinationKey), d.wantLabel(wantTransitions[symbolRef]),
			)
		}
	}
	for _, symbolRef := range slices.Sorted(maps.Keys(gotTransitions)) {
		if _, exists := wantTransitions[symbolRef]; !exists {
			d.report(
				"state %s: unexpected transition on %s to %s",
				d.wantLabel(key), d.symbolString(symbolRef), d.gotLabel(gotTransitions[symbolRef]),
			)
		}
	}
}

// transitions returns the transitions of the state, with every destination named by the identity of its kernel items.
// keyByStateIdx resolves the state index a transition names into that state's identity.
func transitions(state *backend.State, keyByStateIdx []string) map[frontend.SymbolRef]string {
	result := make(map[frontend.SymbolRef]string, state.TransitionActions.Length())
	for _, transitionAction := range state.TransitionActions.All() {
		result[transitionAction.SymbolRef()] = keyByStateIdx[transitionAction.StateIdx()]
	}
	return result
}

// diffReduceActions reports how the reduce actions of two states with the same kernel items differ. The state label
// names the state the two sets of reduce actions belong to.
func (d *LALR1ParserDiffer) diffReduceActions(key string, wantState *backend.State, gotState *backend.State) {
	wantLookaheadSets := d.reduceActions("want", key, wantState)
	gotLookaheadSets := d.reduceActions("got", key, gotState)

	for _, productionIdx := range slices.Sorted(maps.Keys(wantLookaheadSets)) {
		wantLookaheadSet := wantLookaheadSets[productionIdx]
		gotLookaheadSet, exists := gotLookaheadSets[productionIdx]
		if !exists {
			d.report(
				"state %s: missing reduce action for production %d on %s",
				d.wantLabel(key), productionIdx, d.lookaheadSetString(wantLookaheadSet),
			)
			continue
		}
		if !gotLookaheadSet.Equal(wantLookaheadSet) {
			d.report(
				"state %s: reduce action for production %d on %s, want %s",
				d.wantLabel(key), productionIdx,
				d.lookaheadSetString(gotLookaheadSet), d.lookaheadSetString(wantLookaheadSet),
			)
		}
	}
	for _, productionIdx := range slices.Sorted(maps.Keys(gotLookaheadSets)) {
		if _, exists := wantLookaheadSets[productionIdx]; !exists {
			d.report(
				"state %s: unexpected reduce action for production %d on %s",
				d.wantLabel(key), productionIdx, d.lookaheadSetString(gotLookaheadSets[productionIdx]),
			)
		}
	}
}

// reduceActions returns the lookahead set of every production the state reduces. A state holds at most one reduce
// action per production, as two reduce actions for the same production are one reduce action whose lookahead set was
// never merged. Such a pair is reported as a difference, because it makes the two reduce actions incomparable with the
// single reduce action the other parser table has.
func (d *LALR1ParserDiffer) reduceActions(sideLabel string, key string, state *backend.State) map[int]backend.LookaheadSet {
	result := make(map[int]backend.LookaheadSet, state.ReduceActions.Length())

	for _, reduceAction := range state.ReduceActions.All() {
		if _, exists := result[reduceAction.ProductionIdx]; exists {
			d.report(
				"%s: state %s: production %d has more than one reduce action",
				sideLabel, d.wantLabel(key), reduceAction.ProductionIdx,
			)
			continue
		}
		result[reduceAction.ProductionIdx] = reduceAction.LookaheadSet
	}
	return result
}

// symbolString returns the name of the symbol the symbol reference points at, falling back to the symbol reference
// itself when the grammar does not know the symbol. Both parser tables were generated from the same grammar, so the
// grammar of the parser table which is taken to be correct names the symbols of both.
func (d *LALR1ParserDiffer) symbolString(symbolRef frontend.SymbolRef) string {
	symbols := d.wantParser.Grammar.Terminals
	if symbolRef.IsNonterminal() {
		symbols = d.wantParser.Grammar.Nonterminals
	}
	if symbolRef.Idx() >= len(symbols) {
		return symbolRef.String()
	}
	return symbols[symbolRef.Idx()].String()
}

// lookaheadSetString returns the names of the terminals of the lookahead set.
func (d *LALR1ParserDiffer) lookaheadSetString(lookaheadSet backend.LookaheadSet) string {
	var builder strings.Builder
	builder.WriteString("{")
	for terminalIdx := range lookaheadSet.All() {
		if builder.Len() > 1 {
			builder.WriteString(", ")
		}
		builder.WriteString(d.symbolString(frontend.NewTerminalRef(terminalIdx)))
	}
	builder.WriteString("}")
	return builder.String()
}
