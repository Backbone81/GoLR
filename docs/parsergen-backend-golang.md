# Parser Generator Backend: Golang

This backend outputs a parser as Go source code. The generated Go code is a directly coded parser which does not use
a dedicated parsing table, but has the parsing decisions encoded directly in code.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/backend/golang
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32                                93          22936917 ns/op          5173371 B/op     91049 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                                   13          86056179 ns/op         32289215 B/op    605661 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32                          9         120342117 ns/op         50471566 B/op    885634 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                                   3         342840333 ns/op        167407458 B/op   2662183 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                                 10         101395806 ns/op         44605676 B/op    762858 allocs/op
BenchmarkFromParser/Go_1.5.4-32                                       12          92483503 ns/op         42463811 B/op    687514 allocs/op
PASS
ok      golr/internal/parsergen/backend/golang  8.313s
```
