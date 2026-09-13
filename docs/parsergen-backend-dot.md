# Parser Generator Backend: DOT

This backend outputs a parser as a DOT document, for looking at the state machine while working on a grammar. See
[parser generator backends](parsergen-backend.md) for the other backends.

A node is a state, labelled with its index and its kernel items, each prefixed with its production's name (`@name` in
the grammar overrides the auto-generated name) and with a `•` marking how far the parse has come in the production. An
edge is a transition, labelled with the symbol it is taken on: solid for a shift on a terminal, dashed for a goto on a
nonterminal. Reductions and their lookaheads are not in the graph.

Render it with Graphviz:

```sh
golr parser --frontend golr --frontend-file-path calculator.golr --backend dot --backend-file-path parser.dot
dot -Tsvg parser.dot -o parser.svg
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/dot
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32          170       8,711,825 ns/op       3,613,409 B/op       33,686 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32               8     135,100,326 ns/op      35,327,378 B/op      360,646 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32     4     282,721,430 ns/op      76,789,260 B/op      630,719 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32              2     854,661,228 ns/op     204,785,796 B/op    2,204,317 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32             6     208,647,603 ns/op      71,443,925 B/op      592,249 allocs/op
BenchmarkFromParser/Go_1.5.4-32                   9     118,357,505 ns/op      35,448,981 B/op      372,510 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                  7     343,500,047 ns/op     120,577,950 B/op    1,328,627 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            1   6,083,815,314 ns/op   2,929,461,440 B/op   37,788,528 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                3     435,542,674 ns/op     125,006,984 B/op    1,544,339 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/dot       18.630s
```
