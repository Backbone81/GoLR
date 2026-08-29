# Scanner Generator Backends

A backend turns the DFA the core constructed into its output. These backends are currently supported:

- [C](scannergen-backend-c.md)
- [C#](scannergen-backend-csharp.md)
- [C++](scannergen-backend-cpp.md)
- [DOT](scannergen-backend-dot.md)
- [Go](scannergen-backend-golang.md)
- [Java](scannergen-backend-java.md)
- [JavaScript](scannergen-backend-javascript.md)
- [JSON](scannergen-backend-json.md)
- [Kotlin](scannergen-backend-kotlin.md)
- [Null](scannergen-backend-null.md)
- [Python](scannergen-backend-python.md)
- [Rust](scannergen-backend-rust.md)
- [TypeScript](scannergen-backend-typescript.md)
- [YAML](scannergen-backend-yaml.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the scanner as JSON and implement
your own backend by loading the JSON and output it in whatever format you need. You do not even need to do that in Go.
Any programming language which is able to load JSON can be used for such a custom backend. And with outputting JSON to
stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.

The rest of this page describes what the backends which generate a scanner have in common. The page of a language
covers what is specific to it: the names, the configuration, and how that language expresses ownership and its native
string type. The parser side is documented in [parser generator backends](parsergen-backend.md).

## Table driven

Every generated scanner holds the automaton in static lookup tables, read by a driver which is the same handful of
functions in every language. A new language is therefore a new driver rather than another encoding of the automaton,
and one set of rules tokenizes the same way everywhere.

## The tables

The transitions are compressed in two lossless steps, so a lookup returns exactly what walking the DFA would.

First the possible input bytes are partitioned into equivalence classes. Two bytes belong to the same class exactly
when every state transitions on both of them to the same target, which makes them indistinguishable to the scanner. A
scanner reading UTF-8 distinguishes far fewer byte values than it has available - whole ranges such as the continuation
bytes `0x80` to `0xBF` are treated alike everywhere - so indexing a row by the class instead of by the byte shortens
every row considerably.

The resulting rows are then packed into a single array with the row displacement method of "Storing a Sparse Table" by
Tarjan and Yao, so that the entries of one row fall into the holes of the others, with a parallel check array telling
whether the cell a lookup lands on belongs to the row which asked for it.

Where the language offers a choice, every table is emitted with the narrowest unsigned integer type which can hold its
values.

## How a token is matched

The scanner is fully Unicode capable and processes UTF-8 encoded input.

Matching is longest match: the scanner keeps running while the automaton can consume bytes and remembers the last
accepting state it passed through, so a rule matching more bytes wins over one matching fewer. Between rules which
match the same bytes, the one specified earlier in the grammar wins.

## The tokens

- The **end token** is delivered once the source is exhausted, which is how a parser learns the input is over.
- The **invalid token** covers bytes which form no token at all. Scanning does not fail on them: the invalid token
  covers at least the byte the automaton could not consume, and the scan continues after it. Reporting such a token is
  left to whoever reads the tokens.
- The **error token** exists as a constant because the grammar names it as its error recovery point, but no input ever
  produces it. See [error recovery](parsergen-backend.md#error-recovery).
- **Skipped tokens** are those the grammar marked for skipping, usually whitespace and comments. The scanner itself
  delivers them like any other token; a token skipper which wraps a scanner and drops them is generated alongside, and
  that is what a parser is usually given.

## Positions and lexemes

Every token comes with the byte offset it starts at, its line and its column, both counted from one. The column counts
bytes rather than characters, so a multi byte character advances it by more than one. The lexeme is a view into the
source rather than a copy of it, which means the source has to outlive the tokens taken from it.

## Reuse and concurrency

The tables are constant data which every scanner of the same rules shares, so creating one is cheap.

A scanner is reused by resetting it onto a source and a byte offset within it, which is what re-tokenizing the part of
a source which changed needs. The positions it reports stay absolute: the counters are brought up to date over the
bytes before the offset, so a token found after a reset carries the line and column it has in the whole source and not
one relative to where the scan resumed. Tokens read before a reset keep pointing into the source they came from, which
therefore has to stay alive for as long as they are used.

A scanner is a scan in progress and not safe to use from several threads at once. Give every thread its own; since the
tables are read only, nothing has to be shared or locked between them.

## Using a hand-written parser

A generated scanner is usable on its own. It is an iterator over tokens with no knowledge of any parser: advance it,
read the current token with its position and lexeme, and stop when the end token arrives. Nothing about it requires the
consumer to be a generated parser.

The reverse direction, a generated parser fed by a hand-written scanner, is described in
[parser generator backends](parsergen-backend.md#using-a-hand-written-scanner).
