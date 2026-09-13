# Parser Generator Core: LR(1) GoLR

This core generates an LR(1) parser from a context free grammar.

The LR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          1,406       1,045,661 ns/op       298,082 B/op       5,355 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                18      65,521,593 ns/op    12,467,784 B/op     271,657 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32      13      79,949,707 ns/op    17,439,293 B/op     374,487 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                1   1,058,482,938 ns/op   220,291,072 B/op   5,309,129 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32               9     133,541,615 ns/op    39,298,722 B/op     779,206 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                     7     180,124,686 ns/op    41,080,544 B/op     827,871 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                    1   3,869,502,421 ns/op   343,467,368 B/op   7,129,944 allocs/op
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
