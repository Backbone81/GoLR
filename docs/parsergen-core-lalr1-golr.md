# Parser Generator Core: LALR(1) GoLR

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lalr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                  486     2,685,617 ns/op       917,300 B/op       10,225 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                     106    10,385,121 ns/op     3,745,846 B/op      44,152 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32           100    17,216,893 ns/op     5,321,774 B/op      64,695 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                     27    49,792,182 ns/op    23,189,260 B/op     200,699 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                    72    18,173,304 ns/op     6,306,210 B/op      64,683 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          82    13,586,665 ns/op     4,084,055 B/op      44,775 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                         31    39,173,372 ns/op    17,369,922 B/op     156,625 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    3   369,551,878 ns/op   227,163,970 B/op   2,287,820 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/golr   10.261s
```
