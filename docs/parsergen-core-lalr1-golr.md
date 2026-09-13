# Parser Generator Core: LALR(1) GoLR

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lalr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          2,475       786,434 ns/op      245,103 B/op       3,101 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32               132     8,279,986 ns/op    2,227,493 B/op      24,989 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32      74    14,597,906 ns/op    3,080,387 B/op      35,304 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32               32    36,213,963 ns/op   16,461,877 B/op     129,476 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32              96    13,811,251 ns/op    4,170,393 B/op      39,698 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                   124     8,095,774 ns/op    2,914,147 B/op      28,866 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                   42    32,901,993 ns/op   10,921,923 B/op      94,163 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32              7   157,442,370 ns/op   68,707,142 B/op   1,137,367 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                 40    28,870,955 ns/op   12,552,800 B/op     111,299 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/golr   11.301s
```
