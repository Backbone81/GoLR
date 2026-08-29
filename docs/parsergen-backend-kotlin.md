# Parser Generator Backend: Kotlin

This backend outputs a parser as Kotlin source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: Kotlin](scannergen-backend-kotlin.md) for the scanner side.

It targets Kotlin 2.0 on JVM 17.

`--backend-kotlin-package-name` sets the package the file declares, which defaults to `parser` and has to be the one the
scanner was generated into.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is null when the parse could not be finished, which the Kotlin type system states
rather than the documentation. A `ParseSymbol` is a sealed interface over `TerminalSymbol` and `NonterminalSymbol`, so a
`when` over a node symbol needs no else branch. Lexemes are a `ByteBuffer` over the source rather than a copy of it.

## Example

[examples/calculator/kotlin/](../examples/calculator/kotlin/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend kotlin \
  --backend-file-path parser/Parser.kt
```
