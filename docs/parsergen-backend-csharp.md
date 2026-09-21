# Parser Generator Backend: C#

This backend outputs a parser as C# source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C#](scannergen-backend-csharp.md) for the scanner side.

`--backend-csharp-namespace` sets the namespace, which defaults to `Parser` and has to be the one the scanner was
generated into.

`Parser.Parse` takes an `ITokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished. Every node
carries its span as `ByteOffset` and `ByteLength`, which the scanner's `Text` turns into its bytes.

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
BenchmarkFromParser/GNU_Bison_3.8.2-32          660     1,533,952 ns/op       560,666 B/op     8,453 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32             104    11,090,831 ns/op     2,820,517 B/op    29,693 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32    80    12,668,989 ns/op     4,395,699 B/op    41,290 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             45    27,852,856 ns/op    11,202,068 B/op    81,515 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32           100    13,368,913 ns/op     4,745,099 B/op    44,311 allocs/op
BenchmarkFromParser/Go_1.5.4-32                 160     7,932,771 ns/op     2,753,541 B/op    28,761 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 54    23,480,112 ns/op     8,598,926 B/op    60,500 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            3   491,001,208 ns/op   139,326,752 B/op   490,372 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               62    20,368,280 ns/op     9,409,679 B/op    62,724 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/csharp    11.939s
```
