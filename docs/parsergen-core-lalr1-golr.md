# Parser Generator Core: LALR(1) GoLR

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lalr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                 2475            786434 ns/op          245103 B/op       3101 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                     132           8279986 ns/op         2227493 B/op      24989 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32            74          14597906 ns/op         3080387 B/op      35304 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                     32          36213963 ns/op        16461877 B/op     129476 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                    96          13811251 ns/op         4170393 B/op      39698 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                         124           8095774 ns/op         2914147 B/op      28866 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                         42          32901993 ns/op        10921923 B/op      94163 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    7         157442370 ns/op        68707142 B/op    1137367 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                       40          28870955 ns/op        12552800 B/op     111299 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/golr   11.301s
```
