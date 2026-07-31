// Package selftest checks the IELR(1) parser core against canonical LR(1) by driving the two parser tables through the
// same generated sentences and comparing the LR actions they take. It builds on the generators and the parser
// interpreter of the oracle package and adds the comparison itself, so that the behavioral differential test of the
// IELR(1) core and any long running soak test on top of it exercise the very same code.
//
// This package is testing infrastructure. It builds a canonical LR(1) table per grammar, which is expensive on purpose,
// and must never be part of generating a parser for a user.
package selftest
