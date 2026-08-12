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
				"SHIFT ID",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS",
				"SHIFT ID",
				"REDUCE term 1",
				"REDUCE expression 3",
				"ACCEPT",
			)
		})

		It("does not report the end of input, which it shifts but does not return", func() {
			// The augmented grammar is `$accept -> Start $end`, so the end of input is a symbol the automaton shifts
			// like any other and the accept happens in the state after it. It is nevertheless absent from the trace:
			// a generated parser returns the first node of its stack, which is the start symbol, and the end of input
			// sits next to it as a second node which the caller never sees.
			expectParserTrace(emptySpec, "",
				"REDUCE statement_list 0",
				"ACCEPT",
			)
		})

		It("reduces a production which derives the empty string without taking anything off the stack", func() {
			expectParserTrace(emptySpec, "a b",
				"REDUCE statement_list 0",
				"SHIFT ID",
				"REDUCE statement_list 2",
				"SHIFT ID",
				"REDUCE statement_list 2",
				"ACCEPT",
			)
		})

		It("leaves the tokens the scanner skips out of the trace", func() {
			expectParserTrace(expressionSpec, "  a  +  b  ",
				"SHIFT ID",
				"REDUCE term 1",
				"REDUCE expression 1",
				"SHIFT PLUS",
				"SHIFT ID",
				"REDUCE term 1",
				"REDUCE expression 3",
				"ACCEPT",
			)
		})
	})

	Context("a parse which fails", func() {
		It("reports the error at the offset of the token it fails on", func() {
			expectParserTrace(expressionSpec, "a + + b",
				"ERROR 4",
			)
		})

		It("reports an error at the end of the input at the offset one past the last byte", func() {
			expectParserTrace(expressionSpec, "a +",
				"ERROR 3",
			)
		})

		It("takes the default reduction of a state for a token it has no action for", func() {
			// The question mark is no token of this scanner, so the parser sees a token which is no terminal of
			// the grammar and which therefore has no entry in any state. Every state answers such a token with
			// its default action, so two reductions happen before the error is detected at all. Which of them
			// happened is no longer visible, since none of them survives in a tree which is never returned, but
			// that the parser gets as far as the question mark and no further is, and every backend has to detect
			// the error on exactly this token.
			expectParserTrace(expressionSpec, "a ? b",
				"ERROR 2",
			)
		})

		It("unwinds the whole stack when the grammar marks no place to resume at", func() {
			// Nothing in this grammar can shift the error symbol, so the recovery drops every state it has and the
			// parse ends without accepting. How far it unwound is not observable - the states a parser pops are no
			// part of the tree it returns - and the input parsed cleanly right up to the missing closing bracket,
			// so the single error line is the whole of what a caller learns.
			expectParserTrace(expressionSpec, "(a + b",
				"ERROR 6",
			)
		})
	})

	Context("error recovery", func() {
		It("resumes at the place the grammar marked and carries on", func() {
			// The error symbol is shifted where the grammar marks the place to resume at, and the leaf it leaves in
			// the tree is what RESYNC reports. It stands in for the part of the input which was dropped, which is
			// why the statement it belongs to still reduces with two children.
			//
			// The trace also shows the recovery from the outside: the ERROR is reported once, while the discarded
			// AT token and the states which were popped for it leave nothing behind at all.
			expectParserTrace(statementSpec, "a; @ ; b;",
				"ERROR 3",
				"SHIFT ID",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 1",
				"RESYNC",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT ID",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"ACCEPT",
			)
		})

		It("carries on at the first token the resumed state can shift", func() {
			// Every round of recovery either gets the parse going again or consumes one token, which is what keeps a
			// parse from getting stuck between popping and discarding. Four rounds happen here, and the trace is the
			// one of the case above with a different error offset: how many tokens were thrown away is not
			// observable, because a discarded token reaches no node of the tree. What is observable is that the
			// identifier which was already shifted before the error is gone from the trace, since the recovery
			// popped it, and that the parse reaches the end of the input and accepts.
			expectParserTrace(statementSpec, "a; b b b b; c;",
				"ERROR 5",
				"SHIFT ID",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 1",
				"RESYNC",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"SHIFT ID",
				"SHIFT SEMI",
				"REDUCE statement 2",
				"REDUCE statement_list 2",
				"ACCEPT",
			)
		})

		It("reports an error again once three tokens have been shifted", func() {
			// Three identifiers are shifted after the recovery, so the parser trusts its position again by the time
			// the end of the input turns out not to be a semicolon, and reports that second error. The statement it
			// was recovering in never ends, so the parse is given up and the two errors are the whole trace.
			expectParserTrace(idListSpec, "; a a a",
				"ERROR 0",
				"ERROR 7",
			)
		})

		It("suppresses the error while fewer than three tokens have been shifted", func() {
			// The same grammar and the same second error one identifier earlier. Two shifts are not enough for the
			// parser to trust its position again, so the second error is a likely consequence of the first one and
			// goes unreported. That the trace of this input is the one above without its second ERROR line is the
			// whole of the three shift rule.
			expectParserTrace(idListSpec, "; a a",
				"ERROR 0",
			)
		})

		It("gives up at the end of the input, which is the one token it can not discard", func() {
			expectParserTrace(statementSpec, "a; b",
				"ERROR 4",
			)
		})
	})
})
