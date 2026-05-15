# Parser Generator Frontend: YAML

This frontend describes the context free grammar of a parser as YAML document. See the data types in
`internal/parsergen/frontend` for details about the YAML structure.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/frontend/yaml
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                                 13         103960693 ns/op        43509921 B/op    72150 allocs/op
BenchmarkToGrammar/GNU_GCC_2.95.3_C-32                                 1        1450592850 ns/op       758923416 B/op   332830 allocs/op
BenchmarkToGrammar/GNU_GCC_2.95.3_Objective_C-32                       1        2070322384 ns/op       999187432 B/op   436366 allocs/op
BenchmarkToGrammar/GNU_GCC_3.3.6_C++-32                                1        5359249041 ns/op      3676029616 B/op   931957 allocs/op
BenchmarkToGrammar/GNU_GCC_4.2.4_Java-32                               1        2325036569 ns/op      1160001896 B/op   478083 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                                         1        1344231639 ns/op       679062376 B/op   300697 allocs/op
PASS
ok      golr/internal/parsergen/frontend/yaml   14.125s
```
