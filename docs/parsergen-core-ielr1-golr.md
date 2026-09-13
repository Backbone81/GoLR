# Parser Generator Core: IELR(1) GoLR

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/ielr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                 1132           1135589 ns/op          402098 B/op       5274 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      37          30916133 ns/op         8418460 B/op     234910 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32            27          38979613 ns/op        10217952 B/op     259720 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                     10         101529643 ns/op        48111604 B/op     936241 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                    14          87890290 ns/op        25037587 B/op     907317 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          37          44238401 ns/op        13197800 B/op     390335 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          7         157727338 ns/op        82275042 B/op    3123022 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    4         299335323 ns/op        181124936 B/op   4056245 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                       12          99888308 ns/op        49797668 B/op    1462854 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/golr   10.924s
```
