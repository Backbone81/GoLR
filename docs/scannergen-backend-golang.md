# Scanner Generator Backend: Golang

This backend outputs a scanner as Go source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: Go](parsergen-backend-golang.md) for the parser side.

`--backend-go-package-name` sets the package the file declares, which defaults to `parser`.

The scanner is `Scanner` and its tokens are the `Token` constants. `TokenSkipper` wraps a scanner to drop the tokens
marked for skipping. Both implement `TokenSource`, which is what a parser reads its tokens through. Lexemes are a
`[]byte` viewing into the source rather than a copy of it.

## Example

[examples/calculator/golang/](../examples/calculator/golang/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend go \
  --backend-file-path parser/scanner.go
```
