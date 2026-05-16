# Parser Generator Frontend: JSON

This frontend describes the context free grammar of a parser as a JSON document. See the data types in
`internal/parsergen/frontend` for details about the JSON structure.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/frontend/json
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                              10000            217879 ns/op           62946 B/op        361 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                                  2005            512869 ns/op          142293 B/op        673 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32                        2040            584778 ns/op          209715 B/op        860 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                                 1279           1082091 ns/op          281381 B/op       1442 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                                1717            671748 ns/op          218895 B/op        850 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                                      4348            462063 ns/op          141651 B/op        620 allocs/op
PASS
ok      golr/internal/parsergen/frontend/json   8.966s
```
