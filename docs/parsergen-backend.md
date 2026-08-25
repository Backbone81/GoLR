# Parser Generator Backends

A backend turns the parse table the core computed into its output. These backends are currently supported:

- [C](parsergen-backend-c.md)
- [C#](parsergen-backend-csharp.md)
- [C++](parsergen-backend-cpp.md)
- [DOT](parsergen-backend-dot.md)
- [Go](parsergen-backend-golang.md)
- [Java](parsergen-backend-java.md)
- [JavaScript](parsergen-backend-javascript.md)
- [JSON](parsergen-backend-json.md)
- [Null](parsergen-backend-null.md)
- [Python](parsergen-backend-python.md)
- [Rust](parsergen-backend-rust.md)
- [TypeScript](parsergen-backend-typescript.md)
- [YAML](parsergen-backend-yaml.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the parser as JSON and implement
your own backend by loading the JSON and output it in whatever format you need. You do not even need to do that in Go.
Any programming language which is able to load JSON can be used for such a custom backend. And with outputting JSON to
stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.

The rest of this page describes what the backends which generate a parser have in common. The page of a language covers
what is specific to it: the names, the configuration, and how that language expresses errors and ownership. The scanner
side is documented in [scanner generator backends](scannergen-backend.md).

## Table driven

Every generated parser holds its parsing decisions in static lookup tables, read by a driver which is the same handful
of functions in every language. A new language is therefore a new driver rather than another encoding of the automaton,
and one grammar parses the same way everywhere.

## The tables

The action table and the goto table are both stored with the row displacement method of "Storing a Sparse Table" by
Tarjan and Yao, using the first-fit-decreasing placement of its appendix. The rows of all states are packed into a
single array in which the entries of one row fall into the holes of the others, and a parallel check array tells
whether the cell a lookup lands on really belongs to the row which asked for it. Identical rows share one displacement.

What makes the rows sparse enough for this to pay off is that whatever most entries agree on is taken out of them and
kept once as a default. The two tables default in different directions:

- The **action table defaults per state**. The reduction a state performs on most of its lookaheads is kept once as the
  default action of that state instead of once per lookahead. A terminal which the grammar rejects on purpose keeps an
  entry of its own, because an entry beats the default action while an absent one falls through to it.
- The **goto table defaults per nonterminal**. The state which a goto on a nonterminal leads to in most of the states
  which have such a goto is kept once as the default goto of that nonterminal, and only the states which deviate keep
  an entry. This matters because a goto row holds a handful of entries spread over all nonterminals, which makes the
  goto table of a large grammar the biggest of the tables. The price is one more table lookup per reduction.

Where the language offers a choice, every table is emitted with the narrowest unsigned integer type which can hold its
values, so the tables of a small grammar cost a single byte per entry.

The error symbol needs no table of its own. It is a terminal like any other, so its shifts are the entries of the
action table in its column, which is where the error recovery reads the state to resume in. Nothing else reads that
column, because no scanner ever delivers the error symbol - the parser shifts it itself while recovering.

## The parse tree

A parse returns a tree with a node per grammar symbol. A node carries the symbol it stands for, and either the bytes of
the terminal or the child nodes of the production which was reduced to it. Terminal lexemes are a view into the source
rather than a copy, so the source has to outlive the tree. Nodes for the error symbol carry neither, since no input
produced them. Walking such a tree is what the [calculator example](../examples/calculator/) shows.

## Reuse and concurrency

The tables are constant data which every parser of the same grammar shares, so creating one is cheap.

One parser serves one source after another. Parsing again discards the stacks and the errors of the parse before it and
reuses the memory they took, so a parser settles on the size its largest parse needed and allocates little after that.
A tree an earlier parse returned is owned by its result and is not affected by later parses, unless the page of the
language says otherwise.

A parser is a parse in progress and not safe to use from several threads at once. Give every thread its own; since the
tables are read only, nothing has to be shared or locked between them.

## Error recovery

A grammar which marks places to resume at with the error symbol makes the parser continue after a syntax error instead
of stopping at the first one. It then reports every syntax error it found, and returns the tree of the input as far as
it could make sense of it, with a node for the error symbol at every place it resumed at. What the recovery dropped is
not in the tree. A tree is only produced when the parse reached its end; if the recovery gave up in the middle, there
are only errors.

The recovery works the way section 9 "Error Recovery" of "LR Parsing" by Aho and Johnson and section 7 "Error Handling"
of the yacc report describe it, which is also what GNU Bison does:

1. The error is reported.
2. The parser pops its stack until it reaches a state which can shift the error symbol, and shifts it there. Everything
   the popped states had parsed is dropped with them. Where the parse resumes is therefore decided by the position the
   error symbol takes in the production: a production `"{" error "}"` resumes at the closing brace of the block.
3. The token which caused the error is kept, because the resumed state is usually waiting for exactly it - the `"}"` in
   the example above. Only when the parse fails on that same token again is the token discarded and the search for a
   state to resume in repeated. That is what makes recovery terminate: every round either gets the parse going again or
   consumes one token, and the end of input cannot be consumed.
4. Errors are not reported again until three tokens have been shifted, so that a single mistake does not produce an
   avalanche of messages which are all consequences of the first one.

A parser for a grammar which marks no place to resume at cannot recover. It reports the first syntax error it hits and
returns no tree, and it does not carry the countdown of tokens to shift in its shift actions at all.

## Using a hand-written scanner

A generated parser does not require a generated scanner. It reads its tokens through a token source abstraction - an
interface, a trait, a protocol, depending on the language - and anything which implements it will do. Generate the
scanner for your grammar once and read what that abstraction looks like in your language; a hand-written scanner has to
offer the same.

Two things are worth knowing before writing one:

- **The values of the token constants do not matter, only their names.** The parser translates a token into the column
  of the action table which holds the decisions for it, through a lookup indexed by the token value itself, so whatever
  numbering a scanner chose, the entries end up in the right place. A token which is no terminal of this grammar takes
  the default action of the state, which for a state expecting something else is the syntax error it should be. What a
  hand-written scanner does have to deliver is the end token, since that is how the parser learns the input is over.
- **The error token is never produced by a scanner.** It exists as a constant because the grammar names it, but no
  input yields it and the parser shifts it itself while recovering.

The reverse direction, a generated scanner feeding a hand-written parser, is described in
[scanner generator backends](scannergen-backend.md).
