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
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   15       74,360,470 ns/op       9,079,371 B/op      218,632 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                       3      399,133,264 ns/op      78,724,744 B/op    1,941,992 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             2      766,864,608 ns/op     124,384,164 B/op    3,042,123 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      1   61,810,949,308 ns/op   1,445,274,488 B/op   35,339,556 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     1    3,359,567,144 ns/op     249,652,608 B/op    6,123,503 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                           1    1,183,783,369 ns/op     201,685,496 B/op    4,944,557 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          1   54,220,088,738 ns/op   1,501,047,536 B/op   35,933,192 allocs/op
--- FAIL: BenchmarkGrammarToParser/PostgreSQL_18.4
    lr1_test.go:45: executing bison: signal: killed
        
        
--- FAIL: BenchmarkGrammarToParser
FAIL
exit status 1
FAIL    github.com/backbone81/golr/internal/parsergen/core/lr1/bison    184.506s
FAIL
```
