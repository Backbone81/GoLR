# Parser Generator Core: IELR(1) Bison

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the IELR(1) Bison core to work.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/ielr1/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   24          45478338 ns/op         2030080 B/op      46830 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      12         110362378 ns/op        12960702 B/op     313092 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             8         143874956 ns/op        18378025 B/op     443790 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      4         302936800 ns/op        74788266 B/op    1812256 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     7         150152298 ns/op        21982441 B/op     529901 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          12         115414717 ns/op        14684495 B/op     352694 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          4         262174202 ns/op        54024902 B/op    1289309 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    1        2594671886 ns/op        820041232 B/op  19761286 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                        4         263386593 ns/op        58899534 B/op    1429766 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/bison  11.949s
```
