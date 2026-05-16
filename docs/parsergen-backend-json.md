# Parser Generator Backend: JSON

This backend outputs a parser as a JSON document. See the data types in `internal/parsergen/backend` for details about
the JSON structure.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/backend/json
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32                             56114             20791 ns/op             113 B/op          1 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                                17606             68163 ns/op             127 B/op          1 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32                      13008             92018 ns/op             132 B/op          1 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                                6990            169605 ns/op             187 B/op          1 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                              13010             91083 ns/op             132 B/op          1 allocs/op
BenchmarkFromParser/Go_1.5.4-32                                    18098             66510 ns/op             127 B/op          1 allocs/op
PASS
ok      golr/internal/parsergen/backend/json    7.935s
```
