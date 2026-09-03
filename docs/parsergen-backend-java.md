# Parser Generator Backend: Java

This backend outputs a parser as Java source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: Java](scannergen-backend-java.md) for the scanner side.

It targets Java 17.

`--backend-java-package-name` sets the package the file declares, which defaults to `parser` and has to be the one the
scanner was generated into.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` record holding the tree and the errors, and can be
called again with another scanner. The tree is null when the parse could not be finished. Nodes, errors and symbols are
records too, and lexemes are a `ByteBuffer` over the source rather than a copy of it.

Every nonterminal node's `production` component names the alternative it was reduced by, as one of the generated
`Production` values (`PRODUCTION_EXPRESSION1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Example

[examples/calculator/java/](../examples/calculator/java/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend java \
  --backend-file-path parser/Parser.java
```
