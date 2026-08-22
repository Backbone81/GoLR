# Parser Generator Backend: Golang

This backend outputs a parser as Go source code. The generated Go code is a table driven parser which holds the parsing
decisions in lookup tables.

## Error Recovery

`Parse` returns a parse tree and an error. A grammar which marks places to resume at with the error symbol makes the
parser continue after a syntax error instead of stopping at the first one, so both return values can be set at the same
time:

- The error joins every syntax error the parse found with `errors.Join`, in the order they were found. A parse without
  any error returns `nil`, and a parse with a single error joins that one error, so the shape of the returned error
  never depends on how many errors there were. Every one of them wraps `ErrSyntax` and is an `*Error` carrying the
  position it was found at, so `errors.Is(err, ErrSyntax)` tells a syntax error apart from an `ErrInternal`, and
  `errors.As` gets at the position of the first one - both look into a join. Printing the joined error lists all of
  them, one per line. To reach every single error instead of the first, unwrap the join with
  `err.(interface{ Unwrap() []error })`.
- The tree is the tree of the input as far as the parser could make sense of it, with a node for the error symbol at
  every place the parse resumed at. Those nodes carry no lexeme; what the recovery dropped is not in the tree. A tree is
  only returned when the parse reached its end. If the recovery gave up in the middle there is no tree, only the errors.

The recovery works the way section 9 "Error Recovery" of "LR Parsing" by Aho and Johnson and section 7 "Error Handling"
of the yacc report describe it, which is also what GNU Bison does:

1. The error is reported.
2. The parser pops its stack until it reaches a state which can shift the error symbol, and shifts the symbol there.
   Everything the popped states had parsed is dropped with them. Where the parse resumes is therefore decided by the
   position the error symbol takes in the production: a production `"{" error "}"` resumes at the closing brace of the
   block.
3. The token which caused the error is kept, because the resumed state is usually waiting for exactly it - the `"}"` in
   the example above. Only when the parse fails on that same token again is the token discarded and the search for a
   state to resume in repeated. That is what makes recovery terminate: every round either gets the parse going again or
   consumes one token, and the end of input cannot be consumed.
4. Errors are not reported again until three tokens have been shifted, so that a single mistake in the input does not
   produce an avalanche of messages which are all consequences of the first one.

A parser for a grammar which marks no place to resume at cannot recover. It reports the first syntax error it hits and
returns no tree, and it does not carry the countdown of tokens to shift in its shift actions at all.

## The tables

The action table and the goto table are both stored with the row displacement method described in "Storing a Sparse
Table" by Tarjan and Yao, using the first-fit-decreasing placement of its appendix. The rows of all states are packed
into a single array in which the entries of one row fall into the holes of the others, and a parallel check array tells
whether the cell a lookup lands on really belongs to the row which asked for it. Rows which are identical share one
displacement.

What makes the rows sparse enough for this to pay off is that what most entries of a table agree on is taken out of the
rows and kept once as a default. The two tables default in different directions.

The action table defaults per state: the reduction a state performs on most of its lookaheads is kept once as the
default action of the state instead of once per lookahead. A terminal which the grammar rejects on purpose keeps an
entry of its own, because an entry beats the default action while an absent one falls through to it.

The goto table defaults per nonterminal: the state which a goto on a nonterminal leads to in most of the states which
have such a goto is kept once as the default goto of that nonterminal, and only the states which deviate from it keep an
entry. This matters because a goto row holds a handful of entries spread over all nonterminals, which makes the goto
table of a large grammar the biggest of the tables. On the Go grammar of the `examples/golang` example it removes 82 %
of the goto entries, which is 75 % of the bytes the goto tables take. The price is one more table lookup per reduction,
which costs about 2 % of the parse time.

The shift of the error symbol needs no table of its own. The error symbol is a terminal like any other, so its shifts
are the entries of the action table in its column, which is where the error recovery reads the state to resume in.
Nothing else ever reads that column, because no scanner delivers the error symbol - the parser shifts it itself while
recovering.

Every table is emitted with the narrowest unsigned integer type which can hold its values, so the tables of a small
grammar cost a single byte per entry.

## Relationship to the scanner

The generated parser translates the token its scanner delivers into the column of the action table which holds the
decisions for it. It does that with a lookup table whose entries are indexed by the token constants themselves, so the
two generators need not agree on any numbering: whatever values the scanner gave its tokens, the entries end up in the
right place. A token which is not a terminal of the grammar is not in that table at all and takes the default action of
the state.

A hand written scanner therefore only has to provide the token constants under the names the grammar uses. It does not
have to give them particular values.
