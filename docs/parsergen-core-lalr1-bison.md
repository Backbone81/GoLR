# Parser Generator Core: LALR(1) Bison

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the LALR(1) Bison core to work.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/lalr1/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32                   22      46,951,635 ns/op     1,984,004 B/op       46,349 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                      10     102,373,566 ns/op    12,799,727 B/op      311,503 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32             9     118,005,471 ns/op    18,151,924 B/op      441,627 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32                      4     271,169,122 ns/op    73,834,662 B/op    1,795,396 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32                     8     138,376,020 ns/op    21,745,520 B/op      527,624 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                          10     101,787,070 ns/op    14,537,484 B/op      351,286 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                          5     231,060,138 ns/op    53,766,534 B/op    1,286,441 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32                    1   1,972,647,684 ns/op   817,989,984 B/op   19,732,074 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/bison  9.498s
```
