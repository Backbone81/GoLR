# Parser Generator Core: LR(1) GoLR

This core generates an LR(1) parser from a context free grammar.

The LR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                 1406           1045661 ns/op          298082 B/op       5355 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      18          65521593 ns/op        12467784 B/op     271657 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32            13          79949707 ns/op        17439293 B/op     374487 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      1        1058482938 ns/op        220291072 B/op   5309129 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     9         133541615 ns/op        39298722 B/op     779206 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                           7         180124686 ns/op        41080544 B/op     827871 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          1        3869502421 ns/op        343467368 B/op   7129944 allocs/op
--- FAIL: BenchmarkGrammarToParser/PostgreSQL_18.4
    lr1_test.go:46: canonical LR(1): the number of states exceeds the state limit of 64261 states
--- FAIL: BenchmarkGrammarToParser/Ruby_3.2.11
    lr1_test.go:46: canonical LR(1): the number of states exceeds the state limit of 65149 states
--- FAIL: BenchmarkGrammarToParser
FAIL
exit status 1
FAIL    github.com/backbone81/golr/internal/parsergen/core/lr1/golr     14.791s
FAIL
```
