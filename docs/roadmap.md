# Roadmap

## General Topics

- Extend the documentation.
- Add benchmarks to all documentation.
- Add an example for parsing simple mathematical expressions.
- Introduce a proper tutorial for setting up scanner and parser with more details than the getting started.
- Make the example for a Go parser work.
- Add an example for parsing ANTLR v4 grammar files.
- Improve error handling for generated scanner and parser.
- Add support for C backend
- Add support for C++ backend
- Add support for C# backend
- Add support for Java backend
- Add support for Python backend
- Add support for Rust backend
- Add support for JavaScript
- Add support for TypeScript

## Parser Generator

- Compare performance with other generators like https://github.com/goccmack/gocc or Hyacc (https://hyacc.sourceforge.net/)
- The IELR(1) implementation should be native Go and not call out to Bison.
- Introduce strongly typed wrappers for general purpose AST nodes. That way, users don't rely on children being a 
  specific count, but can instead use named methods for directory accessing the correct child. Make sure this is a
  zero overhead abstraction over the AST nodes.

## Scanner Generator

- Allow scanner to parse case independent (accept lower case and upper case characters if specified in one case only)
