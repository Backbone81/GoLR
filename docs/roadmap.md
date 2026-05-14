# Roadmap

## General Topics

- Extend the documentation.
- Make the example for a Go parser work
- Add a simple example for a calculator
- Add benchmarks to all documentation.
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

## Scanner Generator

- Allow scanner to parse case independent (accept lower case and upper case characters if specified in one case only)
