# Scanner Generator Backend: Java

This backend outputs a scanner as Java source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: Java](parsergen-backend-java.md) for the parser side.

It targets Java 17.

`--backend-java-package-name` sets the package the file declares, which defaults to `parser`.

The scanner is the `Scanner` class, its tokens are the nested `Scanner.Token` enum, and the nested
`Scanner.TokenSkipper` wraps a scanner to drop the tokens marked for skipping. Both implement `TokenSource`, which is
what a parser reads its tokens through. Lexemes are a `ByteBuffer` over the source rather than a copy of it.

## Example

[examples/calculator/java/](../examples/calculator/java/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend java \
  --backend-file-path parser/Scanner.java
```
