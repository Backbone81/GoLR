# Scanner Generator Backend: TypeScript

This backend outputs a scanner as TypeScript source code. See [scanner generator backends](scannergen-backend.md) for
what every generated scanner does and [parser generator backend: TypeScript](parsergen-backend-typescript.md) for the
parser side.

The output is an ES module and needs nothing beyond the language itself.

The scanner is `Scanner` and its tokens are the frozen `Token` object together with the `Token` type, with
`tokenToString` and `isSkipped` alongside them. `TokenSkipper` wraps a scanner to drop the tokens marked for skipping.
Both implement the `TokenSource` interface, which is what a parser reads its tokens through. Lexemes are a `Uint8Array`
viewing into the source rather than a copy of it.

## Example

[examples/calculator/typescript/](../examples/calculator/typescript/) is a calculator built on this backend. Its
scanner was generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend typescript \
  --backend-file-path parser/scanner.ts
```
