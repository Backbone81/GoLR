# Parser Generator Backend: Rust

This backend outputs a parser as Rust source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Rust](scannergen-backend-rust.md) for the scanner side.

The output is a module which needs nothing beyond the standard library, and puts no constraint on the edition of the
crate it is dropped into.

`--backend-rust-scanner-module` sets the module path the token type and the token source trait are taken from, which
defaults to `super::scanner`.

`Parser::parse` takes anything implementing `TokenSource` and returns a `ParseResult` holding the tree and the errors,
and can be called again with another scanner. The tree is `None` when the parse could not be finished. The result is
`#[must_use]`, since dropping it unlooked at drops every error the parse reported. Every node carries its span as
`byte_offset` and `byte_length`, which the scanner's `text` turns into its bytes. A `ParseNode` does not borrow the
source and has no lifetime parameter; the `ParseResult` still has one, because its errors borrow their lexemes.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` variants (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It is
`None` on a terminal node.

## Tracing

`Parser` has a `trace` field of type `Option<TraceFunc>` (`Option<Box<dyn FnMut(&str)>>`). Set it after `Parser::new`
for a trace of every parser action, `None` for none. See [parser generator backends](parsergen-backend.md) for the line
format.

## Example

[examples/calculator/rust/](../examples/calculator/rust/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend rust \
  --backend-file-path src/parser/parser.rs
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/rust
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32          586     2,062,307 ns/op       579,541 B/op    11,067 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32              91    12,383,959 ns/op     2,900,584 B/op    37,462 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32   100    16,510,180 ns/op     4,758,329 B/op    51,931 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             39    29,595,849 ns/op    11,909,840 B/op   101,335 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32            93    16,654,309 ns/op     4,824,823 B/op    56,982 allocs/op
BenchmarkFromParser/Go_1.5.4-32                 100    10,524,181 ns/op     2,807,951 B/op    36,065 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 79    21,461,977 ns/op     9,269,092 B/op    75,141 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            3   485,172,606 ns/op   139,005,538 B/op   568,489 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               50    24,349,272 ns/op    10,063,868 B/op    77,754 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/rust      13.097s
```
