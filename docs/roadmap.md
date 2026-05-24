# Roadmap

## General Topics

- GoLR needs to support @precedence on @empty productions for compatibility with GNU Bison languages.
- Extend the documentation.
- Add benchmarks to all documentation.
- Make the example for a Go parser work.
- Add an example for parsing ANTLR v4 grammar files.
- Add support for C backend
- Add support for C++ backend
- Add support for C# backend
- Add support for Java backend
- Add support for Python backend
- Add support for Rust backend
- Add support for JavaScript
- Add support for TypeScript

## Parser Generator

- Improve error handling for generated parser.
- Improve performance
- Compare performance with other generators like https://github.com/goccmack/gocc or Hyacc (https://hyacc.sourceforge.net/)
- The IELR(1) implementation should be native Go and not call out to Bison.
- Introduce strongly typed wrappers for general purpose AST nodes. That way, users don't rely on children being a 
  specific count, but can instead use named methods for directory accessing the correct child. Make sure this is a
  zero overhead abstraction over the AST nodes.

## Scanner Generator

- Improve error handling for generated scanner.
- Improve performance
- Check if any is excluding newlines.
- Add support for fragments - the possibility to declare rules which do not result in a token but can be referenced in
  other rules. This helps to re-use common pattern at several places without having to copy&paste the patterns
  everywhere.
- Allow scanner to parse case independent (accept lower case and upper case characters if specified in one case only)

## Formater

- The GoLR formater should retain comments. Right now, comments are dropped because we parse the grammar file into
  a context-free grammar and regular expressions, then write those out again. As parsing the context-free grammar drops
  all comments, they are lost for writing out again. We need to look into mechanics to pass on dropped comments to the
  output.
