# Parser Generator Core: LR(1) Bison

This core generates an LR(1) parser from a context free grammar.

The LR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the LR(1) Bison core to work.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lr1/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          15       73,295,579 ns/op       9,116,248 B/op      218,631 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32              3      403,481,799 ns/op      79,153,216 B/op    1,941,993 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32    2      797,633,046 ns/op     126,031,980 B/op    3,042,134 allocs/op
--- FAIL: BenchmarkGrammarToParser/GCC_3.3.6_C++
    lr1_test.go:46: executing bison: signal: killed
        
        /tmp/golr-lr1-420347516.y: warning: 655 shift/reduce conflicts [-Wconflicts-sr]
        /tmp/golr-lr1-420347516.y: warning: 87 reduce/reduce conflicts [-Wconflicts-rr]
        /tmp/golr-lr1-420347516.y: note: rerun with option '-Wcounterexamples' to generate conflict counterexamples
        
BenchmarkGrammarToParser/GCC_4.2.4_Java-32            1    3,471,241,783 ns/op     250,731,248 B/op    6,123,496 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                  1    1,290,377,839 ns/op     202,733,176 B/op    4,944,552 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                 1   57,038,746,138 ns/op   1,505,382,432 B/op   35,940,118 allocs/op
--- FAIL: BenchmarkGrammarToParser/PostgreSQL_18.4
    lr1_test.go:46: executing bison: signal: killed
        
        
--- FAIL: BenchmarkGrammarToParser/Ruby_3.2.11
    lr1_test.go:46: executing bison: signal: killed
        
        
--- FAIL: BenchmarkGrammarToParser
FAIL
exit status 1
FAIL    github.com/backbone81/golr/internal/parsergen/core/lr1/bison    245.914s
FAIL
```
