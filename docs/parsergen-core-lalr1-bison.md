# Parser Generator Core: LALR(1) Bison

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the LALR(1) Bison core to work.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lalr1/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   26          45171778 ns/op         2021140 B/op      46347 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      10         103851683 ns/op        12930591 B/op     311504 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             9         138734468 ns/op        18282704 B/op     441626 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      4         272628536 ns/op        74146460 B/op    1795395 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     8         140266836 ns/op        21877251 B/op     527625 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          10         113088494 ns/op        14636024 B/op     351287 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          5         244593421 ns/op        53931297 B/op    1286690 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    1        2070072536 ns/op        819238632 B/op  19733198 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                        4         252278171 ns/op        58845542 B/op    1426598 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/bison  11.146s
```
