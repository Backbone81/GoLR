# Scanner Generator Backend: C#

This backend outputs a scanner as C# source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: C#](parsergen-backend-csharp.md) for the parser side.

`--backend-csharp-namespace` sets the namespace the generated code lives in, which defaults to `Parser`.

The scanner is `Scanner` and its tokens are the `Token` enum, extended by `TokenExtensions` with `IsSkipped` and
`ToDisplayString`. `TokenSkipper` wraps a scanner to drop the tokens marked for skipping. Both implement
`ITokenSource`, which is what a parser reads its tokens through. Lexemes are a `ReadOnlyMemory<byte>` over the source
rather than a copy of it.

## Example

[examples/calculator/csharp/](../examples/calculator/csharp/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend csharp \
  --backend-csharp-namespace Calculator.Parser \
  --backend-file-path Parser/Scanner.cs
```
