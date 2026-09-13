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
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   15          73295579 ns/op         9116248 B/op     218631 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                       3         403481799 ns/op        79153216 B/op    1941993 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             2         797633046 ns/op        126031980 B/op   3042134 allocs/op
--- FAIL: BenchmarkGrammarToParser/GCC_3.3.6_C++
    lr1_test.go:46: executing bison: signal: killed
        
        /tmp/golr-lr1-420347516.y: warning: 655 shift/reduce conflicts [-Wconflicts-sr]
        /tmp/golr-lr1-420347516.y: warning: 87 reduce/reduce conflicts [-Wconflicts-rr]
        /tmp/golr-lr1-420347516.y: note: rerun with option '-Wcounterexamples' to generate conflict counterexamples
        
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     1        3471241783 ns/op        250731248 B/op   6123496 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                           1        1290377839 ns/op        202733176 B/op   4944552 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          1        57038746138 ns/op       1505382432 B/op 35940118 allocs/op
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
