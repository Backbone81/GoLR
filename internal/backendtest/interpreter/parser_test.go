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
				"SHIFT ID 0 1",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS 2 3",
				"SHIFT ID 4 5",
				"REDUCE term 1",
				"REDUCE expression 3",
				"SHIFT $end 5 5",
				"ACCEPT",
			)
		})

		It("shifts the end of input before it accepts", func() {
			// The augmented grammar is `$accept -> Start $end`, so the end of input is a symbol the automaton
			// shifts like any other and the accept happens in the state after it. It carries the offset one past
			// the last byte and an empty extent.
			expectParserTrace(emptySpec, "",
				"REDUCE statement_list 0",
				"SHIFT $end 0 0",
				"ACCEPT",
			)
		})

		It("reduces a production which derives the empty string without taking anything off the stack", func() {
			expectParserTrace(emptySpec, "a b",
				"REDUCE statement_list 0",
				"SHIFT ID 0 1",
				"REDUCE statement_list 2",
				"SHIFT ID 2 3",
				"REDUCE statement_list 2",
				"SHIFT $end 3 3",
				"ACCEPT",
			)
		})

		It("leaves the tokens the scanner skips out of the trace but keeps their offsets", func() {
			expectParserTrace(expressionSpec, "  a  +  b  ",
				"SHIFT ID 2 3",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS 5 6",
				"SHIFT ID 8 9",
				"REDUCE term 1",
				"REDUCE expression 3",
				"SHIFT $end 11 11",
				"ACCEPT",
			)
		})
	})

	Context("a parse which fails", func() {
		It("reports the error at the offset of the token it fails on", func() {
			expectParserTrace(expressionSpec, "a + + b",
				"SHIFT ID 0 1",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS 2 3",
				"ERROR 4",
				"POP 2",
			)
		})

		It("reports an error at the end of the input at the offset one past the last byte", func() {
			expectParserTrace(expressionSpec, "a +",
				"SHIFT ID 0 1",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS 2 3",
				"ERROR 3",
				"POP 2",
			)
		})

		It("takes the default reduction of a state for a token it has no action for", func() {
			// The question mark is no token of this scanner, so the parser sees a token which is no terminal of
			// the grammar and which therefore has no entry in any state. Every state answers such a token with
			// its default action, which is what makes the two reductions happen before the error is detected.
			// Where an error surfaces is thus a property of the default reductions, and every backend has to
			// delay it by exactly as much.
			expectParserTrace(expressionSpec, "a ? b",
				"SHIFT ID 0 1",
				"REDUCE term 1",
				"REDUCE expression 1",
				"ERROR 2",
				"POP 1",
			)
		})

		It("unwinds the whole stack when the grammar marks no place to resume at", func() {
			// Nothing in this grammar can shift the error symbol, so the recovery drops every state it has and
			// the parse ends without accepting. The states it dropped are still reported, because that is what
			// the generated parsers do before they give up.
			expectParserTrace(expressionSpec, "(a + b",
				"SHIFT LPAREN 0 1",
				"SHIFT ID 1 2",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS 3 4",
				"SHIFT ID 5 6",
				"REDUCE term 1",
				"REDUCE expression 3",
				"ERROR 6",
				"POP 2",
			)
		})
	})

	Context("error recovery", func() {
		It("resumes at the place the grammar marked and carries on", func() {
			// The error symbol is shifted where the grammar marks the place to resume at, which the state on top
			// of the stack already is, so nothing has to be popped for it. The offending token is kept for the
			// resumed state to look at and only thrown away when the parse fails on it a second time.
			expectParserTrace(statementSpec, "a; @ ; b;",
				"SHIFT ID 0 1",
				"SHIFT SEMI 1 2",
				"REDUCE statement 2",
				"REDUCE statement_list 1",
				"ERROR 3",
				"POP 0",
				"RESYNC",
				"DISCARD AT 3 4",
				"POP 1",
				"RESYNC",
				"SHIFT SEMI 5 6",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT ID 7 8",
				"SHIFT SEMI 8 9",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT $end 9 9",
				"ACCEPT",
			)
		})

		It("throws away one token per round until the parse can continue", func() {
			// Every round of recovery either gets the parse going again or consumes one token, which is what
			// keeps a parse from getting stuck between popping and discarding. Only the first of the errors is
			// reported, because no token of the input was shifted in between.
			expectParserTrace(statementSpec, "a; b b b b; c;",
				"SHIFT ID 0 1",
				"SHIFT SEMI 1 2",
				"REDUCE statement 2",
				"REDUCE statement_list 1",
				"SHIFT ID 3 4",
				"ERROR 5",
				"POP 1",
				"RESYNC",
				"DISCARD ID 5 6",
				"POP 1",
				"RESYNC",
				"DISCARD ID 7 8",
				"POP 1",
				"RESYNC",
				"DISCARD ID 9 10",
				"POP 1",
				"RESYNC",
				"SHIFT SEMI 10 11",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT ID 12 13",
				"SHIFT SEMI 13 14",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT $end 14 14",
				"ACCEPT",
			)
		})

		It("reports an error again once three tokens have been shifted", func() {
			// Three identifiers are shifted after the recovery, so the parser trusts its position again by the
			// time the end of the input turns out not to be a semicolon, and reports that second error.
			expectParserTrace(idListSpec, "; a a a",
				"ERROR 0",
				"POP 0",
				"RESYNC",
				"DISCARD SEMI 0 1",
				"POP 1",
				"RESYNC",
				"SHIFT ID 2 3",
				"REDUCE id_list 1",
				"SHIFT ID 4 5",
				"REDUCE id_list 2",
				"SHIFT ID 6 7",
				"REDUCE id_list 2",
				"ERROR 7",
				"POP 2",
				"RESYNC",
			)
		})

		It("suppresses the error while fewer than three tokens have been shifted", func() {
			// The same grammar and the same second error one identifier earlier. Two shifts are not enough for
			// the parser to trust its position again, so the second error is a likely consequence of the first
			// one and goes unreported. That the trace of this input is the one above without its second ERROR
			// line is the whole of the three shift rule.
			expectParserTrace(idListSpec, "; a a",
				"ERROR 0",
				"POP 0",
				"RESYNC",
				"DISCARD SEMI 0 1",
				"POP 1",
				"RESYNC",
				"SHIFT ID 2 3",
				"REDUCE id_list 1",
				"SHIFT ID 4 5",
				"REDUCE id_list 2",
				"POP 2",
				"RESYNC",
			)
		})

		It("gives up at the end of the input, which is the one token it can not discard", func() {
			expectParserTrace(statementSpec, "a; b",
				"SHIFT ID 0 1",
				"SHIFT SEMI 1 2",
				"REDUCE statement 2",
				"REDUCE statement_list 1",
				"SHIFT ID 3 4",
				"ERROR 4",
				"POP 1",
				"RESYNC",
			)
		})
	})
})
