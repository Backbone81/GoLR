# Parser Generator Backend: Kotlin

This backend outputs a parser as Kotlin source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: Kotlin](scannergen-backend-kotlin.md) for the scanner side.

It targets Kotlin 2.0 on JVM 17.

`--backend-kotlin-package-name` sets the package the file declares, which defaults to `parser` and has to be the one the
scanner was generated into.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished, which the Kotlin type system states
rather than the documentation. A `ParseSymbol` is a sealed interface over `TerminalSymbol` and `NonterminalSymbol`, so a
`when` over a node symbol needs no else branch. Lexemes are a `ByteBuffer` over the source rather than a copy of it.

Every nonterminal node's `production` property names the alternative it was reduced by, as one of the generated
`Production` entries (`PRODUCTION_EXPRESSION1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Tracing

`Parser` has a `trace` property of type `TraceFunc?` (`((String) -> Unit)?`). Set it after `Parser()` for a trace of
every parser action, `null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/kotlin/](../examples/calculator/kotlin/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend kotlin \
  --backend-file-path parser/Parser.kt
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/kotlin
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32          544     3,038,761 ns/op     1,621,227 B/op     8,301 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32             100    16,657,734 ns/op     5,998,335 B/op    27,932 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32    68    18,216,015 ns/op     8,737,876 B/op    38,676 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             42    31,539,675 ns/op    18,435,571 B/op    75,043 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32           100    16,776,669 ns/op     8,708,151 B/op    40,034 allocs/op
BenchmarkFromParser/Go_1.5.4-32                 100    12,887,210 ns/op     5,644,054 B/op    27,235 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 63    24,274,882 ns/op    13,806,844 B/op    55,931 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            3   482,136,835 ns/op   163,887,882 B/op   462,146 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               57    27,039,550 ns/op    15,606,098 B/op    58,995 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/kotlin    14.325s
```
