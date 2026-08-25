# Parser Generator Backend: TypeScript

This backend outputs a parser as TypeScript source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: TypeScript](scannergen-backend-typescript.md) for the
scanner side.

The output is an ES module and needs nothing beyond the language itself.

`--backend-typescript-scanner-module` sets the module specifier the token constants are imported from, which defaults
to `./scanner.js`.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished. Lexemes are a `Uint8Array` viewing
into the source rather than a copy of it.

## Example

[examples/calculator/typescript/](../examples/calculator/typescript/) is a calculator built on this backend. Its parser
was generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend typescript \
  --backend-file-path parser/parser.ts
```
