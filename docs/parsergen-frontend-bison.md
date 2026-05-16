# Parser Generator Frontend: Bison

This frontend describes the context free grammar of a parser as a [GNU Bison](https://www.gnu.org/software/bison/) grammar document.

The following functionality is currently supported:

- %token
- %left
- %right
- %nonassoc
- %precedence
- rules
- %prec
- %start

Any not supported functionality is ignored.

The GNU Bison grammar parser is tested against a set of well known GNU Bison grammar files for several programming
languages, to make sure that it works correctly. The well known grammar files include GNU Bison, GCC C, GCC
Objective C, GCC C++, GCC Java and Go.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: golr/internal/parsergen/frontend/bison
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                                699           1610186 ns/op          473257 B/op       6235 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                                  1020           3314797 ns/op         1169322 B/op      17970 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32                         223           5782654 ns/op         1506658 B/op      23352 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                                  126           9164394 ns/op         2619912 B/op      41781 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                                 159           7745372 ns/op         2420908 B/op      24190 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                                       288           3487969 ns/op         1029408 B/op      16399 allocs/op
PASS
ok      golr/internal/parsergen/frontend/bison  9.202s
```
