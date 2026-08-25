# Scanner Generator Backend: Python

This backend outputs a scanner as Python source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: Python](parsergen-backend-python.md) for the parser side.

It targets Python 3.10 and imports nothing beyond the standard library.

The scanner is `Scanner` and its tokens are the `Token` enum, with `is_skipped` alongside it. `TokenSkipper` wraps a
scanner to drop the tokens marked for skipping. Both satisfy the `TokenSource` protocol, which is what a parser reads
its tokens through. Lexemes are `bytes` sliced out of the source.

## Example

[examples/calculator/python/](../examples/calculator/python/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend python \
  --backend-file-path parser/scanner.py
```
