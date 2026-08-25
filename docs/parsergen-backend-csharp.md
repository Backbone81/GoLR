# Parser Generator Backend: C#

This backend outputs a parser as C# source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C#](scannergen-backend-csharp.md) for the scanner side.

`--backend-csharp-namespace` sets the namespace, which defaults to `Parser` and has to be the one the scanner was
generated into.

`Parser.Parse` takes an `ITokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished. Lexemes are a
`ReadOnlyMemory<byte>` over the source rather than a copy of it.

## Example

[examples/calculator/csharp/](../examples/calculator/csharp/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend csharp \
  --backend-csharp-namespace Calculator.Parser \
  --backend-file-path Parser/Parser.cs
```
