# Go

This example demonstrates the use of GoLR for parsing Go source code.

***IMPORTANT: This example is work in progress. The scanner is working but the parser still needs to be created.***

The generated scanner produces the same tokens as the official `go/scanner`. This is validated against the full source
code of the Go standard library. Roughly 7,700 Go source code files.

The generated scanner provides comparable performance to the official `go/scanner`. A benchmark with scanning
`net/http/server.go` which is roughly 130 KB in size delivers the following results:

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/examples/golang/parser
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkGolangScanner/Official_Go_Scanner-32              17208            687693 ns/op          125817 B/op       7266 allocs/op
BenchmarkGolangScanner/GoLR_Generated_Scanner-32           14458            825925 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/backbone81/golr/examples/golang/parser       23.794s
```

Note that the generated scanner does not cause any memory allocations during tokenization.
