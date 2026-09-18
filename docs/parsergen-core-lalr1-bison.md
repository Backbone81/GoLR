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
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          26      45,171,778 ns/op     2,021,140 B/op       46,347 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32             10     103,851,683 ns/op    12,930,591 B/op      311,504 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32    9     138,734,468 ns/op    18,282,704 B/op      441,626 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32             4     272,628,536 ns/op    74,146,460 B/op    1,795,395 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32            8     140,266,836 ns/op    21,877,251 B/op      527,625 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                 10     113,088,494 ns/op    14,636,024 B/op      351,287 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                 5     244,593,421 ns/op    53,931,297 B/op    1,286,690 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32           1   2,070,072,536 ns/op   819,238,632 B/op   19,733,198 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32               4     252,278,171 ns/op    58,845,542 B/op    1,426,598 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/lalr1/bison  11.146s
```

**NOTE:** This benchmark includes writing out the grammar as GNU Bison grammar file, executing the GNU Bison executable
on that file and then loading the XML report to construct the parser tables. This means that these numbers cannot be
directly compared with the GoLR core, as that can execute the core directly in-memory in the same process.
