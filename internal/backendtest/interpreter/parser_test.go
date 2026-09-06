package interpreter_test

import (
	. "github.com/onsi/ginkgo/v2"
)

// expressionSpec is a grammar which needs a stack to parse and marks no place to resume at, so that a failing parse
// gives up instead of recovering.
const expressionSpec = `
@scanner {
    WHITESPACE: /[ \t\n]+/ @skip;
    ID:         /[a-z]+/;
    PLUS:       "+";
    LPAREN:     "(";
    RPAREN:     ")";
}

@parser {
    expression
        : expression PLUS term
        | term
        ;

    term
        : ID
        | LPAREN expression RPAREN
        ;
}
`

// emptySpec is the smallest grammar with a production which derives the empty string, which is what a reduction taking
// nothing off the stack comes from.
const emptySpec = `
@scanner {
    WHITESPACE: /[ \t\n]+/ @skip;
    ID:         /[a-z]+/;
}

@parser {
    statement_list
        : @empty
        | statement_list ID
        ;
}
`

// statementSpec marks a place to resume at after a syntax error and is written so that the resumed state is waiting for
// exactly one token, the semicolon, which is the shape the error recovery is designed around.
//
// AT is a rule of the scanner which no production uses. It gives the recovery a token to throw away which the scanner
// nevertheless has a name for, next to the input no rule matches at all.
const statementSpec = `
@scanner {
    WHITESPACE: /[ \t\n]+/ @skip;
    ID:         /[a-z]+/;
    AT:         "@";
    SEMI:       ";";
}

@parser {
    statement_list
        : statement
        | statement_list statement
        ;

    statement
        : ID SEMI
        | @error SEMI
        ;
}
`

// idListSpec resumes into a production which shifts an unbounded number of tokens before it ends, so that a second
// error can be placed at any distance from the first one. That distance is what decides whether the second error is
// reported or suppressed.
const idListSpec = `
@scanner {
    WHITESPACE: /[ \t\n]+/ @skip;
    ID:         /[a-z]+/;
    SEMI:       ";";
}

@parser {
    statement_list
        : statement
        | statement_list statement
        ;

    statement
        : ID SEMI
        | @error id_list SEMI
        ;

    id_list
        : ID
        | id_list ID
        ;
}
`

var _ = Describe("Parser", func() {
	Context("a parse which succeeds", func() {
		It("shifts, reduces and accepts", func() {
			expectParserTrace(expressionSpec, "a + b",
				`1:1     SHIFT   ID "a"`,
				"1:3     REDUCE  term => ID",
				"1:3     REDUCE  expression => term",
				`1:3     SHIFT   PLUS "+"`,
				`1:5     SHIFT   ID "b"`,
				"1:6     REDUCE  term => ID",
				"1:6     REDUCE  expression => expression PLUS term",
				`1:6     SHIFT   $end ""`,
				"1:6     ACCEPT",
			)
		})

		It("shifts the end of input like any other symbol before it accepts", func() {
			// The augmented grammar is `$accept -> Start $end`, so the end of input is a symbol the automaton shifts
			// like any other and the accept happens in the state after it. The tree derived trace left it out
			// because a generated parser returns only the start symbol node, but the action trace shows it.
			expectParserTrace(emptySpec, "",
				"1:1     REDUCE  statement_list => ε",
				`1:1     SHIFT   $end ""`,
				"1:1     ACCEPT",
			)
		})

		It("reduces a production which derives the empty string without taking anything off the stack", func() {
			expectParserTrace(emptySpec, "a b",
				"1:1     REDUCE  statement_list => ε",
				`1:1     SHIFT   ID "a"`,
				"1:3     REDUCE  statement_list => statement_list ID",
				`1:3     SHIFT   ID "b"`,
				"1:4     REDUCE  statement_list => statement_list ID",
				`1:4     SHIFT   $end ""`,
				"1:4     ACCEPT",
			)
		})

		It("leaves the tokens the scanner skips out of the trace", func() {
			expectParserTrace(expressionSpec, "  a  +  b  ",
				`1:3     SHIFT   ID "a"`,
				"1:6     REDUCE  term => ID",
				"1:6     REDUCE  expression => term",
				`1:6     SHIFT   PLUS "+"`,
				`1:9     SHIFT   ID "b"`,
				"1:12    REDUCE  term => ID",
				"1:12    REDUCE  expression => expression PLUS term",
				`1:12    SHIFT   $end ""`,
				"1:12    ACCEPT",
			)
		})
	})

	Context("a parse which fails", func() {
		It("reports the error at the position of the token it fails on", func() {
			expectParserTrace(expressionSpec, "a + + b",
				`1:1     SHIFT   ID "a"`,
				"1:3     REDUCE  term => ID",
				"1:3     REDUCE  expression => term",
				`1:3     SHIFT   PLUS "+"`,
				"1:5     ERROR   unexpected token PLUS",
				"1:5     POP     PLUS",
				"1:5     POP     expression",
				"1:5     FAIL",
			)
		})

		It("reports an error at the end of the input one column past the last byte", func() {
			expectParserTrace(expressionSpec, "a +",
				`1:1     SHIFT   ID "a"`,
				"1:3     REDUCE  term => ID",
				"1:3     REDUCE  expression => term",
				`1:3     SHIFT   PLUS "+"`,
				"1:4     ERROR   unexpected token $end",
				"1:4     POP     PLUS",
				"1:4     POP     expression",
				"1:4     FAIL",
			)
		})

		It("takes the default reduction of a state for a token it has no action for", func() {
			// The question mark is no token of this scanner, so the parser sees a token which is no terminal of the
			// grammar and which therefore has no entry in any state. Every state answers such a token with its
			// default action, so the two reductions of `a` into `expression` happen before the error is detected at
			// all, on exactly the question mark, which every backend has to agree on.
			expectParserTrace(expressionSpec, "a ? b",
				`1:1     SHIFT   ID "a"`,
				"1:3     REDUCE  term => ID",
				"1:3     REDUCE  expression => term",
				"1:3     ERROR   unexpected token $invalid",
				"1:3     POP     expression",
				"1:3     FAIL",
			)
		})

		It("unwinds the whole stack when the grammar marks no place to resume at", func() {
			// Nothing in this grammar can shift the error symbol, so the recovery pops every state it has and the
			// parse fails without accepting. The POP lines show how far it unwound, which the tree derived trace
			// could not.
			expectParserTrace(expressionSpec, "(a + b",
				`1:1     SHIFT   LPAREN "("`,
				`1:2     SHIFT   ID "a"`,
				"1:4     REDUCE  term => ID",
				"1:4     REDUCE  expression => term",
				`1:4     SHIFT   PLUS "+"`,
				`1:6     SHIFT   ID "b"`,
				"1:7     REDUCE  term => ID",
				"1:7     REDUCE  expression => expression PLUS term",
				"1:7     ERROR   unexpected token $end",
				"1:7     POP     expression",
				"1:7     POP     LPAREN",
				"1:7     FAIL",
			)
		})
	})

	Context("error recovery", func() {
		It("resumes at the place the grammar marked and carries on", func() {
			// The recovery is visible from the outside: the state after `statement_list` can already shift the error
			// symbol, so the first RESYNC needs no pop, but the resumed state wants a semicolon and the input still
			// holds the `@`. So the parser fails again, this time discarding the `@`, popping the error symbol and
			// resyncing a second time, where the semicolon finally matches. Only the first ERROR is reported; the
			// second is marked suppressed.
			expectParserTrace(statementSpec, "a; @ ; b;",
				`1:1     SHIFT   ID "a"`,
				`1:2     SHIFT   SEMI ";"`,
				"1:4     REDUCE  statement => ID SEMI",
				"1:4     REDUCE  statement_list => statement",
				"1:4     ERROR   unexpected token AT",
				"1:4     RESYNC",
				"1:4     ERROR   (suppressed) unexpected token AT",
				`1:4     DISCARD AT "@"`,
				"1:6     POP     $error",
				"1:6     RESYNC",
				`1:6     SHIFT   SEMI ";"`,
				"1:8     REDUCE  statement => $error SEMI",
				"1:8     REDUCE  statement_list => statement_list statement",
				`1:8     SHIFT   ID "b"`,
				`1:9     SHIFT   SEMI ";"`,
				"1:10    REDUCE  statement => ID SEMI",
				"1:10    REDUCE  statement_list => statement_list statement",
				`1:10    SHIFT   $end ""`,
				"1:10    ACCEPT",
			)
		})

		It("carries on at the first token the resumed state can shift", func() {
			// Every round of recovery either gets the parse going again or consumes one token, which is what keeps a
			// parse from getting stuck between popping and discarding. The identifier `b` shifted before the error is
			// popped, then three more `b` tokens are discarded one per round until the semicolon is reached and the
			// statement reduces.
			expectParserTrace(statementSpec, "a; b b b b; c;",
				`1:1     SHIFT   ID "a"`,
				`1:2     SHIFT   SEMI ";"`,
				"1:4     REDUCE  statement => ID SEMI",
				"1:4     REDUCE  statement_list => statement",
				`1:4     SHIFT   ID "b"`,
				"1:6     ERROR   unexpected token ID",
				"1:6     POP     ID",
				"1:6     RESYNC",
				"1:6     ERROR   (suppressed) unexpected token ID",
				`1:6     DISCARD ID "b"`,
				"1:8     POP     $error",
				"1:8     RESYNC",
				"1:8     ERROR   (suppressed) unexpected token ID",
				`1:8     DISCARD ID "b"`,
				"1:10    POP     $error",
				"1:10    RESYNC",
				"1:10    ERROR   (suppressed) unexpected token ID",
				`1:10    DISCARD ID "b"`,
				"1:11    POP     $error",
				"1:11    RESYNC",
				`1:11    SHIFT   SEMI ";"`,
				"1:13    REDUCE  statement => $error SEMI",
				"1:13    REDUCE  statement_list => statement_list statement",
				`1:13    SHIFT   ID "c"`,
				`1:14    SHIFT   SEMI ";"`,
				"1:15    REDUCE  statement => ID SEMI",
				"1:15    REDUCE  statement_list => statement_list statement",
				`1:15    SHIFT   $end ""`,
				"1:15    ACCEPT",
			)
		})

		It("reports an error again once three tokens have been shifted", func() {
			// Three identifiers are shifted after the recovery, so the parser trusts its position again by the time
			// the end of the input turns out not to end the id_list, and the ERROR at 1:8 is reported rather than
			// suppressed. The statement never ends, so the recovery unwinds the whole stack and the parse fails.
			expectParserTrace(idListSpec, "; a a a",
				"1:1     ERROR   unexpected token SEMI",
				"1:1     RESYNC",
				"1:1     ERROR   (suppressed) unexpected token SEMI",
				`1:1     DISCARD SEMI ";"`,
				"1:3     POP     $error",
				"1:3     RESYNC",
				`1:3     SHIFT   ID "a"`,
				"1:5     REDUCE  id_list => ID",
				`1:5     SHIFT   ID "a"`,
				"1:7     REDUCE  id_list => id_list ID",
				`1:7     SHIFT   ID "a"`,
				"1:8     REDUCE  id_list => id_list ID",
				"1:8     ERROR   unexpected token $end",
				"1:8     POP     id_list",
				"1:8     POP     $error",
				"1:8     RESYNC",
				"1:8     ERROR   (suppressed) unexpected token $end",
				"1:8     FAIL",
			)
		})

		It("suppresses the error while fewer than three tokens have been shifted", func() {
			// The same grammar with one identifier fewer. Two shifts are not enough for the parser to trust its
			// position again, so the ERROR at 1:6 where the id_list fails to end is marked suppressed rather than
			// reported.
			expectParserTrace(idListSpec, "; a a",
				"1:1     ERROR   unexpected token SEMI",
				"1:1     RESYNC",
				"1:1     ERROR   (suppressed) unexpected token SEMI",
				`1:1     DISCARD SEMI ";"`,
				"1:3     POP     $error",
				"1:3     RESYNC",
				`1:3     SHIFT   ID "a"`,
				"1:5     REDUCE  id_list => ID",
				`1:5     SHIFT   ID "a"`,
				"1:6     REDUCE  id_list => id_list ID",
				"1:6     ERROR   (suppressed) unexpected token $end",
				"1:6     POP     id_list",
				"1:6     POP     $error",
				"1:6     RESYNC",
				"1:6     ERROR   (suppressed) unexpected token $end",
				"1:6     FAIL",
			)
		})

		It("gives up at the end of the input, which is the one token it can not discard", func() {
			expectParserTrace(statementSpec, "a; b",
				`1:1     SHIFT   ID "a"`,
				`1:2     SHIFT   SEMI ";"`,
				"1:4     REDUCE  statement => ID SEMI",
				"1:4     REDUCE  statement_list => statement",
				`1:4     SHIFT   ID "b"`,
				"1:5     ERROR   unexpected token $end",
				"1:5     POP     ID",
				"1:5     RESYNC",
				"1:5     ERROR   (suppressed) unexpected token $end",
				"1:5     FAIL",
			)
		})
	})
})
