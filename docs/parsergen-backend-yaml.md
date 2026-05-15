# Parser Generator Backend: YAML

This backend outputs a parser as a YAML document. See the data types in `internal/parsergen/backend` for details about
the YAML structure.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/backend/yaml
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32                                20          87341128 ns/op          22565518 B/op   515121 allocs/op
BenchmarkFromParser/GNU_GCC_2.95.3_C-32                                3         333425621 ns/op         253420722 B/op  5708362 allocs/op
BenchmarkFromParser/GNU_GCC_2.95.3_Objective_C-32                      3         400683658 ns/op         365832709 B/op  8246917 allocs/op
BenchmarkFromParser/GNU_GCC_3.3.6_C++-32                               1        1382845169 ns/op        1783626880 B/op 40070471 allocs/op
BenchmarkFromParser/GNU_GCC_4.2.4_Java-32                              3         470571238 ns/op         496778629 B/op 11191575 allocs/op
BenchmarkFromParser/Go_1.5.4-32                                        3         345187723 ns/op         316222085 B/op  7099273 allocs/op
PASS
ok      golr/internal/parsergen/backend/yaml    8.686s
```
