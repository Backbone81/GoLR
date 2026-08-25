# Parser Generator Backend: Python

This backend outputs a parser as Python source code. See [parser generator backends](parsergen-backend.md) for what
every generated parser does and [scanner generator backend: Python](scannergen-backend-python.md) for the scanner side.

It targets Python 3.10 and imports nothing beyond the standard library.

`--backend-python-scanner-module` sets the module the token constants are imported from, which defaults to `scanner`. A
leading dot makes the import relative, which is what a scanner and a parser sitting in the same package need.

`Parser.parse` takes a `TokenSource` and returns a `ParseResult` holding the tree and the errors, and can be called
again with another scanner. The tree is `None` when the parse could not be finished. Errors are returned rather than
raised, so reporting many of them costs nothing. Lexemes are `bytes` sliced out of the source.

## Example

[examples/calculator/python/](../examples/calculator/python/) is a calculator built on this backend. Its parser was
generated with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend python \
  --backend-python-scanner-module .scanner \
  --backend-file-path parser/parser.py
```
