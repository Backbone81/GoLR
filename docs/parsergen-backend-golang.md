# Parser Generator Backend: Golang

This backend outputs a parser as Go source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Go](scannergen-backend-golang.md) for the scanner side.

`--backend-go-package-name` sets the package the file declares, which defaults to `parser`.

`Parse` takes a `TokenSource` and returns a parse tree and an error, and can be called again with another scanner. Both
can be set at the same time when the parse recovered from an error:

- The error joins every syntax error the parse found with `errors.Join`, in the order they were found. A parse without
  any error returns `nil`, and a parse with a single error joins that one error, so the shape of the returned error
  never depends on how many there were. Every one of them wraps `ErrSyntax` and is an `*Error` carrying the position it
  was found at, so `errors.Is(err, ErrSyntax)` tells a syntax error apart from an `ErrInternal`, and `errors.As` gets
  at the position of the first one - both look into a join. Printing the joined error lists all of them, one per line.
  To reach every single error instead of the first, unwrap the join with `err.(interface{ Unwrap() []error })`.
- The tree is the zero `Node` when the parse could not be finished. Its lexemes are a `[]byte` viewing into the source
  rather than a copy of it.

The nodes of the tree are handed out from an arena which the parser reuses, which is what keeps a parse from allocating
once per node. A tree therefore stays valid only until the next call to `Parse` on the same parser, which hands the
same memory out again. Where the trees of two parses have to be alive at the same time, parse them with a parser each.

## Example

[examples/calculator/golang/](../examples/calculator/golang/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend go \
  --backend-file-path parser/parser.go
```
