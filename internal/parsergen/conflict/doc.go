// Package conflict provides the conflict resolution of a parser: it describes the actions which compete for a terminal
// in a state, and it decides which of them wins.
//
// The core of the package is the dominant contribution function, which the paper "The IELR(1) algorithm for generating
// minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy
// calls delta and defines in section 2.3. It takes a conflicted terminal and the set of actions which compete for that
// terminal, and it returns the action which wins, that the terminal is rejected, or that the conflict stays
// unresolved. It deliberately does not take the state the conflict occurs in, because IELR(1) evaluates it on
// hypothetical sets of actions which do not exist in any state.
//
// The rules by which the dominant contribution is picked are not hard coded. They are composed from policies, so that a
// user of the parser generator can decide which of them apply. See Policy and CompoundPolicy for the details.
//
// The package serves two purposes. Resolve applies the dominant contribution function to a whole parser table, which
// turns a parser table with conflicts into a parser table without them. This is phase 5 of IELR(1), and it is what
// makes the LR(1) and LALR(1) parser tables usable as well. A conflict which the policies do not decide leaves the
// parser table with more than one action for a terminal, which no parser can be generated from, so Resolve reports it
// as an error. And phase 3 of IELR(1) uses the dominant contribution function on its own to decide whether two states
// can be merged.
package conflict
