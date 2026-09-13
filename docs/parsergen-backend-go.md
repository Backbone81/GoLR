# Parser Generator Backend: Go

This backend outputs a parser as Go source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Go](scannergen-backend-go.md) for the scanner side.

`--backend-go-package-name` sets the package the file declares, which defaults to `parser`.

`Parse` takes a `TokenSource` and returns a parse tree and an error, and can be called again with another scanner. Both
can be set at the same time when the parse recovered from an error:

- The error joins every syntax error the parse found with `errors.Join`, in the order they were found. A parse without
  any error returns `nil`, and a parse with a single error joins that one error, so the shape of the returned error
  never depends on how many there were. Every one of them wraps `ErrSyntax` and is an `*Error` carrying the position it
  was found at, so `errors.Is(err, ErrSyntax)` tells a syntax error apart from an `ErrInternal`, and `errors.As` gets
  at the position of the first one - both look into a join. Printing the joined error lists all of them, one per line.
  To reach every single error instead of the first, unwrap the join with `err.(interface{ Unwrap() []error })`.
- The tree is the zero `Node` when the parse could not be finished. Its lexemes are a `[]byte` viewing into the source
  rather than a copy of it.

The nodes of the tree are handed out from an arena which the parser reuses, which is what keeps a parse from allocating
once per node. A tree therefore stays valid only until the next call to `Parse` on the same parser, which hands the
same memory out again. Where the trees of two parses have to be alive at the same time, parse them with a parser each.

Every nonterminal node's `Production` field names the alternative it was reduced by, as one of the generated
`Production` constants (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It
is `NoProduction` on a terminal node, which no production reduces to.

## Tracing

`Parser` has a `Trace` field of type `TraceFunc` (`func(line string)`). Set it after `NewParser` for a trace of every
parser action, `nil` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/golang/](../examples/calculator/golang/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend go \
  --backend-file-path parser/parser.go
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/golang
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32          140     8,792,898 ns/op     1,740,183 B/op      25,582 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32              38    30,271,635 ns/op     6,980,064 B/op      86,356 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32    30    33,932,609 ns/op    10,406,859 B/op     121,466 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             18    61,846,396 ns/op    25,532,936 B/op     256,425 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32            62    32,371,588 ns/op    12,057,729 B/op     132,929 allocs/op
BenchmarkFromParser/Go_1.5.4-32                  58    24,625,770 ns/op     6,648,402 B/op      80,470 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 25    49,095,633 ns/op    20,530,715 B/op     209,148 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            2   668,769,280 ns/op   259,272,740 B/op   1,812,616 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               19    57,069,744 ns/op    23,384,700 B/op     235,858 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/golang    12.503s
```
