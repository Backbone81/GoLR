# Go

This example demonstrates the use of GoLR for parsing Go source code. It can be used to list all tokens and display an
abstract syntax tree of such code.

***IMPORTANT: The Go programming language has some ambiguities which are very hard to solve in the context of an LR(1)
grammar. The generated parser therefore accepts a superset of valid Go code. A separate pass over the parsed syntax
tree would need to ensure that the Go code is in fact valid Go code. That additional pass is not implemented int his
example.***

The generated scanner produces the same tokens as the official `go/scanner`. This is validated against the full source
code of the Go standard library. Roughly 6,500 Go source code files.

The generated parser can not be tested against the official `go/parser`, because the GoLR grammar produces a slightly
different syntax tree. The parser is still validated against the full source code of the Go standard library to make
sure that those files parse successfully.

The performance of the generated scanner and parser can be compared with a benchmark to the official Go scanner and
parser:

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/examples/golang/parser
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGolangParser/Official_Go_Parser-32                  270           4557704 ns/op         1142887 B/op      29010 allocs/op
BenchmarkGolangParser/GoLR_Generated_Parser-32               100          11662332 ns/op         3696479 B/op      85514 allocs/op
BenchmarkGolangScanner/Official_Go_Scanner-32               1557            706204 ns/op          125891 B/op       7266 allocs/op
BenchmarkGolangScanner/GoLR_Generated_Scanner-32            1290            867959 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/golr/examples/golang/parser       4.636s
```

The benchmark uses the file `net/http/server.go` from the Go standard library as input. It is about 130 KB in size.
The generated scanner provides comparable performance to the official `go/scanner` without causing any memory
allocations during tokenization.
