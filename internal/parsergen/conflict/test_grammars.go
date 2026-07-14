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
