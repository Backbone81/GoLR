# Parser Generator Core: LR(1) GoLR

This core generates an LR(1) parser from a context free grammar.

The LR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                  363       3,017,001 ns/op     1,081,457 B/op      14,490 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      19      61,886,183 ns/op    16,775,543 B/op     336,840 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32            14      75,155,007 ns/op    23,758,374 B/op     479,082 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      1   1,058,058,914 ns/op   308,685,800 B/op   6,369,290 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     9     122,180,162 ns/op    51,124,312 B/op     949,037 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                           7     157,781,861 ns/op    49,511,249 B/op     972,951 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          1   3,846,715,848 ns/op   453,096,936 B/op   8,187,791 allocs/op
--- FAIL: BenchmarkGrammarToParser/PostgreSQL_18.4
    lr1_test.go:45: the number of states exceeds the state limit of 64261 states
--- FAIL: BenchmarkGrammarToParser
FAIL
exit status 1
FAIL    github.com/backbone81/golr/internal/parsergen/core/lr1/golr     10.671s
FAIL
```
