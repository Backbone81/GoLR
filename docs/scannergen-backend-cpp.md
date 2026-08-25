# Scanner Generator Backend: C++

This backend outputs a scanner as C++ source code. See [scanner generator backends](scannergen-backend.md) for what
every generated scanner does and [parser generator backend: C++](parsergen-backend-cpp.md) for the parser side.

It targets C++17 and is a self contained header which includes nothing beyond the standard library.

`--backend-cpp-namespace` sets the namespace the header declares everything in, which defaults to `parser`.

The scanner is `Scanner` and its tokens are the `Token` enum class. `TokenSkipper` wraps a scanner to drop the tokens
marked for skipping; it is a template rather than an implementation of an interface, so anything offering the same
members can take the place of a scanner. Lexemes are a `std::string_view` into the source, which therefore has to
outlive them.

## Example

[examples/calculator/cpp/](../examples/calculator/cpp/) is a calculator built on this backend. Its scanner was
generated with:

```sh
golr scanner \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend cpp \
  --backend-cpp-namespace calculator::parser \
  --backend-file-path parser/scanner.hpp
```
