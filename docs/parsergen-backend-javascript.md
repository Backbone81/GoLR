# Parser Generator Backend: JavaScript

This backend outputs a parser as JavaScript source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: JavaScript](scannergen-backend-javascript.md) for the
scanner side.

The output is an ES module and needs nothing beyond the language itself.

`--backend-javascript-scanner-module` sets the module specifier the token constants are imported from, which defaults
to `./scanner.js`.

`Parser.parse` takes a scanner and returns a result holding the tree and the errors, and can be called again with
another scanner. The tree is null when the parse could not be finished. Lexemes are a `Uint8Array` viewing into the
source rather than a copy of it.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` values (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It is
null on a terminal node.

## Example

[examples/calculator/javascript/](../examples/calculator/javascript/) is a calculator built on this backend. Its parser
was generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend javascript \
  --backend-file-path parser/parser.js
```
