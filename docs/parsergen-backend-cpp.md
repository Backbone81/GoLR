# Parser Generator Backend: C++

This backend outputs a parser as C++ source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C++](scannergen-backend-cpp.md) for the scanner side.

It targets C++17 and is a self contained header.

`--backend-cpp-namespace` sets the namespace, which defaults to `parser` and has to be the one the scanner was
generated into. `--backend-cpp-scanner-include` sets the header the token type is included from, which defaults to
`scanner.hpp`.

`Parser::parse` returns a `ParseResult` holding the tree and the errors, and can be called again with a different
scanner. It is a template over the scanner type rather than a function taking an interface, so any type with the
members the generated scanner has will do. The tree holds `std::string_view` lexemes into the source, which therefore
has to outlive it.

## Example

[examples/calculator/cpp/](../examples/calculator/cpp/) is a calculator built on this backend. Its parser was generated
with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend cpp \
  --backend-cpp-namespace calculator::parser \
  --backend-file-path parser/parser.hpp
```
