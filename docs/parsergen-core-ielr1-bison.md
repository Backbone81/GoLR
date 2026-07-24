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
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   24      44,518,491 ns/op     1,983,660 B/op        46,348 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      10     104,413,784 ns/op    12,799,565 B/op       311,503 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             8     135,036,064 ns/op    18,151,969 B/op       441,627 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      4     291,144,836 ns/op    74,354,762 B/op     1,808,309 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     8     149,219,068 ns/op    21,745,794 B/op       527,624 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          10     107,250,655 ns/op    14,537,556 B/op       351,286 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          5     244,404,737 ns/op    53,768,100 B/op     1,286,445 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    1   2,507,860,223 ns/op   818,595,760 B/op   19,746,162 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/bison  10.396s
```
