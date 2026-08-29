# Scanner Generator Backend: Kotlin

This backend outputs a scanner as Kotlin source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: Kotlin](parsergen-backend-kotlin.md) for the parser side.

It targets Kotlin 2.0 on JVM 17.

`--backend-kotlin-package-name` sets the package the file declares, which defaults to `parser`.

The scanner is the `Scanner` class, its tokens are the `Token` enum, and `TokenSkipper` wraps a scanner to drop the
tokens marked for skipping. Both implement `TokenSource`, which is what a parser reads its tokens through, and the
skipper hands every member but `next` on to the scanner it wraps by class delegation. What a scanner reports is
properties rather than methods, so it is `scanner.token` and not `scanner.token()`. Lexemes are a `ByteBuffer` over the
source rather than a copy of it.

## Example

[examples/calculator/kotlin/](../examples/calculator/kotlin/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend kotlin \
  --backend-file-path parser/Scanner.kt
```
