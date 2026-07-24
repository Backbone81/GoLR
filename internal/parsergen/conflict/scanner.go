package conflict

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// TerminalContributions is one conflicted terminal of a state together with the actions which compete for it. This is
// one entry of the contributions function of definition 2.17 of IELR(1), restricted to the conflicted terminals of
// definition 2.18.
type TerminalContributions struct {
	// TerminalIdx is the terminal index of the conflicted terminal.
	TerminalIdx int

	// Contributions are the actions which compete for the conflicted terminal. It always holds more than one
	// contribution, because that is what makes the terminal conflicted.
	Contributions ContributionSet
}

// Scanner reports the conflicted terminals of a state, which is what everyone who works with the contributions
// function is really after.
//
// The scanner only materializes a contribution set for the terminals it finds a second action for. Both of those are
// bitsets indexed by the terminal index, so the conflicted terminals come out in ascending order, which is the order
// the conflicts are reported in.
//
// The zero value is ready to use. A scanner is not safe for concurrent use, and it is worth keeping one alive for a
// whole pass over the parser tables: the bitsets are cleared per state but keep the storage they have grown to, so they
// stop growing once the scanner has seen the highest terminal index of the grammar.
type Scanner struct {
	// seenTerminalIdxs holds the terminals the state has at least one action for.
	seenTerminalIdxs utils.Bitset

	// conflictedTerminalIdxs holds the terminals the state has more than one action for.
	conflictedTerminalIdxs utils.Bitset
}

// Conflicts appends every conflicted terminal of the state to dst, in ascending terminal order, and returns the
// extended slice.
func (s *Scanner) Conflicts(state *backend.State) []TerminalContributions {
	s.findConflictedTerminals(state)

	var result []TerminalContributions
	for terminalIdx := range s.conflictedTerminalIdxs.All() {
		result = append(result, TerminalContributions{
			TerminalIdx:   terminalIdx,
			Contributions: s.contributionsOfTerminal(state, terminalIdx),
		})
	}
	return result
}

// findConflictedTerminals fills conflictedTerminalIdxs with the terminals the state has more than one action for.
//
// A terminal is conflicted as soon as an action is recorded for it which is not the first one, so every terminal an
// action covers which seenTerminalIdxs already holds is conflicted. Recording the terminal as conflicted more than
// once, which happens when a terminal has three actions or more, does not change the outcome.
//
// The terminals a reduce action covers are its lookahead set, which is a bitset already, so a whole reduce action is
// recorded by combining that bitset with the two the scanner keeps. That is what makes the scan cost a handful of word
// operations per reduce action instead of one bitset lookup per terminal in its lookahead set.
func (s *Scanner) findConflictedTerminals(state *backend.State) {
	s.seenTerminalIdxs.Clear()
	s.conflictedTerminalIdxs.Clear()

	for _, transition := range state.TransitionActions.All() {
		if transition.SymbolRef().IsNonterminal() {
			// A goto is not an action on a terminal, so it can never take part in a conflict.
			continue
		}
		if !s.seenTerminalIdxs.Add(transition.SymbolRef().Idx()) {
			// We have already seen this terminal, so it must be a conflict.
			s.conflictedTerminalIdxs.Add(transition.SymbolRef().Idx())
		}
	}
	for _, reduction := range state.ReduceActions.All() {
		// Every terminal the reduction covers which was already seen has more than one action on it now.
		s.conflictedTerminalIdxs.MergeIntersection(&s.seenTerminalIdxs, &reduction.LookaheadSet)
		s.seenTerminalIdxs.Merge(&reduction.LookaheadSet)
	}
}

// contributionsOfTerminal builds the contribution set of a single terminal. This is only called for the terminals which
// the counting pass found to be conflicted, so the cost of walking the actions of the state again is paid for the few
// conflicts only.
func (s *Scanner) contributionsOfTerminal(state *backend.State, terminalIdx int) ContributionSet {
	var result ContributionSet
	for _, transition := range state.TransitionActions.All() {
		if transition.SymbolRef().IsNonterminal() {
			continue
		}
		if transition.SymbolRef().Idx() == terminalIdx {
			result.Add(NewShiftContribution())
		}
	}
	for _, reduction := range state.ReduceActions.All() {
		if reduction.LookaheadSet.Contains(terminalIdx) {
			result.Add(NewReduceContribution(reduction.ProductionIdx))
		}
	}
	return result
}
