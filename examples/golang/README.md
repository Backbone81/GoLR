# Go

This example demonstrates the use of GoLR for parsing Go source code. It can be used to list all tokens and display a
parse tree of such code.

***IMPORTANT: The Go programming language has some ambiguities which are very hard to solve in the context of an LR(1)
grammar. The generated parser therefore accepts a superset of valid Go code. A separate pass over the parsed syntax
tree would need to ensure that the Go code is in fact valid Go code. That additional pass is not implemented in this
example.***

The generated scanner produces the same tokens as the official `go/scanner`. This is validated against the full source
code of the Go standard library. Roughly 6,500 Go source code files.

The generated parser can not be tested against the official `go/parser`, because the GoLR grammar produces a slightly
different parse tree. The parser is still validated against the full source code of the Go standard library to make
sure that those files parse successfully.

The performance of the generated scanner and parser can be compared with a benchmark to the official Go scanner and
parser:

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/examples/golang/parser
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGolangParser/Official_Go_Parser-32          228   4,867,542 ns/op   887,541 B/op   25,411 allocs/op
BenchmarkGolangParser/GoLR_Generated_Parser-32       852   1,176,418 ns/op     3,692 B/op        0 allocs/op
BenchmarkGolangScanner/Official_Go_Scanner-32      1,564     742,033 ns/op   125,846 B/op    7,273 allocs/op
BenchmarkGolangScanner/GoLR_Generated_Scanner-32   1,674     637,284 ns/op         0 B/op        0 allocs/op
PASS
ok      github.com/backbone81/golr/examples/golang/parser       4.353s
```

The benchmark uses the file `net/http/server.go` from the Go standard library as input. It is about 130 KB in size.

The generated scanner and parser are created once and reset for every iteration of the benchmark. The parser reuses
its internal memory between runs, which is why no allocations show up in the benchmark.

With that in mind, the generated scanner performs on par with the official `go/scanner`.

The difference between the parsers is mostly caused by garbage collection. Most of the time of the official parser is
spent on garbage collection triggered by its allocations, which also makes its timings vary noticeably between runs.
