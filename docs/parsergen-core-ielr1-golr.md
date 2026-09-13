# Parser Generator Core: IELR(1) GoLR

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation is a native Go implementation.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/core/ielr1/golr
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGrammarToParser/GNU_Bison_3.8.2-32          1,132     1,135,589 ns/op       402,098 B/op       5,274 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_C-32                37    30,916,133 ns/op     8,418,460 B/op     234,910 allocs/op
BenchmarkGrammarToParser/GCC_2.95.3_Objective_C-32      27    38,979,613 ns/op    10,217,952 B/op     259,720 allocs/op
BenchmarkGrammarToParser/GCC_3.3.6_C++-32               10   101,529,643 ns/op    48,111,604 B/op     936,241 allocs/op
BenchmarkGrammarToParser/GCC_4.2.4_Java-32              14    87,890,290 ns/op    25,037,587 B/op     907,317 allocs/op
BenchmarkGrammarToParser/Go_1.5.4-32                    37    44,238,401 ns/op    13,197,800 B/op     390,335 allocs/op
BenchmarkGrammarToParser/PHP_8.6.7-32                    7   157,727,338 ns/op    82,275,042 B/op   3,123,022 allocs/op
BenchmarkGrammarToParser/PostgreSQL_18.4-32              4   299,335,323 ns/op   181,124,936 B/op   4,056,245 allocs/op
BenchmarkGrammarToParser/Ruby_3.2.11-32                 12    99,888,308 ns/op    49,797,668 B/op   1,462,854 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/core/ielr1/golr   10.924s
```
