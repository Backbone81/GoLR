# Parser Generator Core: IELR(1) Bison

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the IELR(1) Bison core to work.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/ielr1/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          24      45,478,338 ns/op     2,030,080 B/op       46,830 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32             12     110,362,378 ns/op    12,960,702 B/op      313,092 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32    8     143,874,956 ns/op    18,378,025 B/op      443,790 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32             4     302,936,800 ns/op    74,788,266 B/op    1,812,256 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32            7     150,152,298 ns/op    21,982,441 B/op      529,901 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                 12     115,414,717 ns/op    14,684,495 B/op      352,694 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                 4     262,174,202 ns/op    54,024,902 B/op    1,289,309 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32           1   2,594,671,886 ns/op   820,041,232 B/op   19,761,286 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32               4     263,386,593 ns/op    58,899,534 B/op    1,429,766 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/bison  11.949s
```

**NOTE:** This benchmark includes writing out the grammar as GNU Bison grammar file, executing the GNU Bison executable
on that file and then loading the XML report to construct the parser tables. This means that these numbers cannot be
directly compared with the GoLR core, as that can execute the core directly in-memory in the same process.
