# Parser Generator Backend: C#

This backend outputs a parser as C# source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C#](scannergen-backend-csharp.md) for the scanner side.

`--backend-csharp-namespace` sets the namespace, which defaults to `Parser` and has to be the one the scanner was
generated into.

`Parser.Parse` takes an `ITokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished. Lexemes are a
`ReadOnlyMemory<byte>` over the source rather than a copy of it.

Every nonterminal node's `Production` property names the alternative it was reduced by, as one of the generated
`Production` members (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Tracing

`Parser` has a `Trace` property of type `Action<string>?`. Set it after `new Parser()` for a trace of every parser
action, `null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/csharp/](../examples/calculator/csharp/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend csharp \
  --backend-csharp-namespace Calculator.Parser \
  --backend-file-path Parser/Parser.cs
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/csharp
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               660           1533952 ns/op          560666 B/op       8453 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  104          11090831 ns/op         2820517 B/op      29693 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         80          12668989 ns/op         4395699 B/op      41290 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  45          27852856 ns/op        11202068 B/op      81515 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                100          13368913 ns/op         4745099 B/op      44311 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      160           7932771 ns/op         2753541 B/op      28761 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      54          23480112 ns/op         8598926 B/op      60500 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         491001208 ns/op        139326752 B/op    490372 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    62          20368280 ns/op         9409679 B/op      62724 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/csharp    11.939s
```
