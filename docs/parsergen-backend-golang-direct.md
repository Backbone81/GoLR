# Parser Generator Backend: Golang Direct

This backend outputs a parser as Go source code. The generated Go code is a directly coded parser which does not use
a dedicated parsing table, but has the parsing decisions encoded directly in code.

## Error Recovery

`Parse` returns a parse tree and an error. A grammar which marks places to resume at with the error symbol makes the
parser continue after a syntax error instead of stopping at the first one, so both return values can be set at the same
time:

- The error joins every syntax error the parse found with `errors.Join`, in the order they were found. A parse without
  any error returns `nil`, and a parse with a single error joins that one error, so the shape of the returned error
  never depends on how many errors there were. Every one of them wraps `ErrSyntax` and is an `*Error` carrying the
  position it was found at, so `errors.Is(err, ErrSyntax)` tells a syntax error apart from an `ErrInternal`, and
  `errors.As` gets at the position of the first one - both look into a join. Printing the joined error lists all of
  them, one per line. To reach every single error instead of the first, unwrap the join with
  `err.(interface{ Unwrap() []error })`.
- The tree is the tree of the input as far as the parser could make sense of it, with a node for the error symbol at
  every place the parse resumed at. Those nodes carry no lexeme; what the recovery dropped is not in the tree. A tree is
  only returned when the parse reached its end. If the recovery gave up in the middle there is no tree, only the errors.

The recovery works the way section 9 "Error Recovery" of "LR Parsing" by Aho and Johnson and section 7 "Error Handling"
of the yacc report describe it, which is also what GNU Bison does:

1. The error is reported.
2. The parser pops its stack until it reaches a state which can shift the error symbol, and shifts the symbol there.
   Everything the popped states had parsed is dropped with them. Where the parse resumes is therefore decided by the
   position the error symbol takes in the production: a production `"{" error "}"` resumes at the closing brace of the
   block.
3. The token which caused the error is kept, because the resumed state is usually waiting for exactly it - the `"}"` in
   the example above. Only when the parse fails on that same token again is the token discarded and the search for a
   state to resume in repeated. That is what makes recovery terminate: every round either gets the parse going again or
   consumes one token, and the end of input cannot be consumed.
4. Errors are not reported again until three tokens have been shifted, so that a single mistake in the input does not
   produce an avalanche of messages which are all consequences of the first one.

A parser for a grammar which marks no place to resume at cannot recover. It reports the first syntax error it hits and
returns no tree, and it does not carry the countdown of tokens to shift in its shift actions at all.

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
