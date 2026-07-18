package conflict

import (
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/dsl"
)

// The symbols and the productions of PrecedenceTestGrammar, which the tests of the policies refer to. The precedence
// levels are numbered in the order in which the declarations appear, so "*" binds tighter than "+".
const (
	PrecedenceTestGrammarTerminalIdxPlus     = 0 // left associative, precedence 1
	PrecedenceTestGrammarTerminalIdxTimes    = 1 // left associative, precedence 2
	PrecedenceTestGrammarTerminalIdxCompare  = 2 // nonassociative, precedence 3
	PrecedenceTestGrammarTerminalIdxPower    = 3 // right associative, precedence 4
	PrecedenceTestGrammarTerminalIdxIdentity = 4 // no precedence declared

	PrecedenceTestGrammarProductionIdxPlus     = 0 // E -> E + E, precedence of "+"
	PrecedenceTestGrammarProductionIdxTimes    = 1 // E -> E * E, precedence of "*"
	PrecedenceTestGrammarProductionIdxCompare  = 2 // E -> E < E, precedence of "<"
	PrecedenceTestGrammarProductionIdxPower    = 3 // E -> E ^ E, precedence of "^"
	PrecedenceTestGrammarProductionIdxIdentity = 4 // E -> id, no precedence
)

// PrecedenceTestGrammar is the classic ambiguous expression grammar. It is ambiguous on purpose, because a grammar
// without conflicts has nothing for a conflict resolution policy to do.
//
// The grammar declares a terminal for every case a policy has to tell apart: an operator which binds looser than the
// one it is compared against, one which binds equally tight, one which binds tighter, an operator of each
// associativity, and an identity terminal without any precedence at all. Every production inherits the precedence of
// its operator, because that operator is the rightmost terminal on the right hand side, except for the production of
// the identity terminal, which has no precedence to inherit.
//
//	E -> E + E    left associative, precedence 1
//	E -> E * E    left associative, precedence 2
//	E -> E < E    nonassociative,   precedence 3
//	E -> E ^ E    right associative, precedence 4
//	E -> id       no precedence
var PrecedenceTestGrammar = newPrecedenceTestGrammar()

func newPrecedenceTestGrammar() frontend.Grammar {
	grammar := dsl.NewGrammar()

	plus := grammar.Terminal("+")
	times := grammar.Terminal("*")
	compare := grammar.Terminal("<")
	power := grammar.Terminal("^")
	identity := grammar.Terminal("id")

	// The order of the declarations is what gives the terminals their precedence level.
	grammar.Left(plus)
	grammar.Left(times)
	grammar.Nonassoc(compare)
	grammar.Right(power)

	expression := grammar.Nonterminal("E")
	grammar.Production(expression).Rhs(expression, plus, expression)
	grammar.Production(expression).Rhs(expression, times, expression)
	grammar.Production(expression).Rhs(expression, compare, expression)
	grammar.Production(expression).Rhs(expression, power, expression)
	grammar.Production(expression).Rhs(identity)

	return grammar.Build()
}

// The symbols and the productions of MultiRejecterTestGrammar, which the tests of the policies refer to.
const (
	MultiRejecterTestGrammarTerminalIdxLess      = 0 // nonassociative, precedence 1
	MultiRejecterTestGrammarTerminalIdxLessEqual = 1 // nonassociative, precedence 1
	MultiRejecterTestGrammarTerminalIdxTilde     = 2 // no associativity declared, precedence 2
	MultiRejecterTestGrammarTerminalIdxIdentity  = 3 // no precedence declared

	MultiRejecterTestGrammarProductionIdxLess      = 0 // E -> E < E, precedence of "<"
	MultiRejecterTestGrammarProductionIdxLessEqual = 1 // E -> E <= E, precedence of "<="
	MultiRejecterTestGrammarProductionIdxTilde     = 2 // E -> E ~ E, precedence of "~"
	MultiRejecterTestGrammarProductionIdxIdentity  = 3 // E -> id, no precedence
)

// MultiRejecterTestGrammar is an ambiguous expression grammar which covers the two precedence declarations
// PrecedenceTestGrammar cannot express.
//
// The comparison operators share a single nonassociativity declaration, so their productions carry the same precedence
// level as both terminals: a conflict on a comparison operator can hold two reductions which reject the terminal,
// which is what exercises a policy on several rejecting reductions at once. PrecedenceTestGrammar can never produce
// more than one rejecter, because only a single production carries the precedence of its nonassociative terminal.
//
// The "~" operator has a precedence but no associativity, like a terminal declared with %precedence in GNU Bison. A
// conflict between its shift and its production compares equal precedence levels and then finds no associativity to
// decide with, which is the one outcome of a precedence comparison the declarations of PrecedenceTestGrammar cannot
// reach.
//
//	E -> E < E    nonassociative, precedence 1
//	E -> E <= E   nonassociative, precedence 1
//	E -> E ~ E    no associativity, precedence 2
//	E -> id       no precedence
var MultiRejecterTestGrammar = newMultiRejecterTestGrammar()

func newMultiRejecterTestGrammar() frontend.Grammar {
	grammar := dsl.NewGrammar()

	less := grammar.Terminal("<")
	lessEqual := grammar.Terminal("<=")
	tilde := grammar.Terminal("~")
	identity := grammar.Terminal("id")

	// A single declaration puts both comparison operators on the same precedence level, and the precedence declaration
	// gives "~" a level without an associativity.
	grammar.Nonassoc(less, lessEqual)
	grammar.Precedence(tilde)

	expression := grammar.Nonterminal("E")
	grammar.Production(expression).Rhs(expression, less, expression)
	grammar.Production(expression).Rhs(expression, lessEqual, expression)
	grammar.Production(expression).Rhs(expression, tilde, expression)
	grammar.Production(expression).Rhs(identity)

	return grammar.Build()
}
