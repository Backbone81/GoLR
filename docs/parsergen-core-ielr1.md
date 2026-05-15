# Parser Generator Core: IELR(1)

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation delegates the parser generation to GNU Bison for now. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison binary needs to be available
on your system for the IELR(1) core to work. While this is a simple way to make IELR(1) quickly available for GoLR, the
long term goal is to provide an IELR(1) implementation which is written natively in Go.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/core/ielr1
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                           26          46293690 ns/op         1980136 B/op      46341 allocs/op
BenchmarkGrammarToParser/GNU_GCC_2.95.3_C-32                          12         111736112 ns/op        12861540 B/op     312450 allocs/op
BenchmarkGrammarToParser/GNU_GCC_2.95.3_Objective_C-32                 8         139732381 ns/op        18262669 B/op     443000 allocs/op
BenchmarkGrammarToParser/GNU_GCC_3.3.6_C++-32                          4         267770437 ns/op        74992694 B/op    1811244 allocs/op
BenchmarkGrammarToParser/GNU_GCC_4.2.4_Java-32                         8         147587847 ns/op        21915791 B/op     528448 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                                  10         109010010 ns/op        14626240 B/op     352016 allocs/op
PASS
ok      golr/internal/parsergen/core/ielr1      7.022s
```
