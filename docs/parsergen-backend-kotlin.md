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
BenchmarkFromParser/GNU_Bison_3.8.2-32               544           3038761 ns/op         1621227 B/op       8301 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  100          16657734 ns/op         5998335 B/op      27932 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         68          18216015 ns/op         8737876 B/op      38676 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  42          31539675 ns/op        18435571 B/op      75043 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                100          16776669 ns/op         8708151 B/op      40034 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      100          12887210 ns/op         5644054 B/op      27235 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      63          24274882 ns/op        13806844 B/op      55931 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         482136835 ns/op        163887882 B/op    462146 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    57          27039550 ns/op        15606098 B/op      58995 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/kotlin    14.325s
```
