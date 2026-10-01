// Package counterexample finds counterexamples for the conflicts of a parser, following "Finding Counterexamples from
// Parsing Conflicts" by Chinawat Isradisaikul and Andrew C. Myers (PLDI '15). A counterexample shows the grammar author
// where a conflict comes from: a unifying one is a single string of symbols with two derivations, which proves the
// grammar ambiguous, and a nonunifying one is two derivations which share a prefix up to the conflict.
//
// The search runs on the parser tables a core returned, with the conflicts resolved or left unresolved. It only walks
// items and transitions, and takes lookaheads from the items it walks, never from the reduce actions of the tables,
// which the conflict resolution changed.
package counterexample
