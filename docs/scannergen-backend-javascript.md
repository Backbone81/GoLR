# Scanner Generator Backend: JavaScript

This backend outputs a scanner as JavaScript source code. See [scanner generator backends](scannergen-backend.md) for
what every generated scanner does and [parser generator backend: JavaScript](parsergen-backend-javascript.md) for the
parser side.

The output is an ES module and needs nothing beyond the language itself.

The scanner is `Scanner` and its tokens are the frozen `Token` object, with `tokenToString` and `isSkipped` alongside
it. `TokenSkipper` wraps a scanner to drop the tokens marked for skipping. Both offer the same members, which is what a
parser reads its tokens through. Lexemes are a `Uint8Array` viewing into the source rather than a copy of it.

## Example

[examples/calculator/javascript/](../examples/calculator/javascript/) is a calculator built on this backend. Its
scanner was generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend javascript \
  --backend-file-path parser/scanner.js
```
