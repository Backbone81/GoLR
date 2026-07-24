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
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                  271     4,729,509 ns/op     1,802,717 B/op      20,340 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      43    30,047,329 ns/op    12,109,732 B/op     317,463 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32            33    33,399,732 ns/op    15,505,019 B/op     368,178 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                     12   100,863,338 ns/op    64,589,066 B/op   1,275,405 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                    16    75,531,426 ns/op    29,961,156 B/op     979,557 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          32    33,539,526 ns/op    16,454,090 B/op     493,073 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          7   170,827,791 ns/op   101,598,875 B/op   3,857,363 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    2   752,167,772 ns/op   555,968,408 B/op   7,379,586 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/golr   9.908s
```
