# Parser Generator Backend: Java

This backend outputs a parser as Java source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Java](scannergen-backend-java.md) for the scanner side.

It targets Java 17.

`--backend-java-package-name` sets the package the file declares, which defaults to `parser` and has to be the one the
scanner was generated into.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` record holding the tree and the errors, and can be
called again with another scanner. The tree is null when the parse could not be finished. Nodes, errors and symbols are
records too, and lexemes are a `ByteBuffer` over the source rather than a copy of it.

Every nonterminal node's `production` component names the alternative it was reduced by, as one of the generated
`Production` values (`PRODUCTION_EXPRESSION1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Tracing

`Parser` has a public `trace` field of type `Consumer<String>`. Set it after `new Parser()` for a trace of every parser
action, `null` for none. See [parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/java/](../examples/calculator/java/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend java \
  --backend-file-path parser/Parser.java
```

### Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/java
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               488           2222957 ns/op          566755 B/op       6537 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  100          11335877 ns/op         2775184 B/op      23247 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32         86          16025230 ns/op         4297058 B/op      32341 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  42          27259813 ns/op        11027636 B/op      64650 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                100          16752842 ns/op         4282084 B/op      33725 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      196           6865891 ns/op         2519857 B/op      22704 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      91          19707347 ns/op         8531550 B/op      48324 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         481864244 ns/op        137370389 B/op    424017 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    55          21820584 ns/op         9238438 B/op      50006 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/java      13.161s
```
