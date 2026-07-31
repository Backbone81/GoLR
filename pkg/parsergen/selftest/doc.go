// Package selftest provides the public API re-exports for the IELR(1) self test. It checks the IELR(1) parser core
// against a canonical LR(1) oracle by generating random grammars, building both parser tables and driving them through
// the same generated sentences, and reports whether the two ever disagree.
//
// This is testing infrastructure for GoLR itself, not a step of generating a parser: a run builds a canonical LR(1)
// table per grammar, which is expensive on purpose, and it never terminates on its own unless it is told to.
package selftest
