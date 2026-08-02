# Roadmap

## General Topics

- Extend the documentation.
- Add benchmarks to all documentation.
- Publish Visual Studio Code extension for syntax highlighting of GoLR files
- Publish IntelliJ plugin for syntax highlighting of GoLR files
- Add support for C backend
- Add support for C++ backend
- Add support for C# backend
- Add support for Java backend
- Add support for Python backend
- Add support for Rust backend
- Add support for JavaScript
- Add support for TypeScript
- Support extended Backus-Naur form for GoLR productions
- Reduce shift/reduce conflict of the golang example to 0.

## Parser Generator

- The reduction lookahead builder of the IELR(1) core computes the always follows and the goto follows for every goto
  up front. Check if computing them lazily, only for the gotos which a reduce action actually traces back to, is
  faster. A significant number of follow sets could be left uncomputed that way. Note that the later IELR(1) phases
  read both tables for arbitrary gotos, so laziness has to survive those accesses.
- Introduce strongly typed wrappers for general purpose parse nodes. That way, users don't rely on children being a 
  specific count, but can instead use named methods for directly accessing the correct child. Make sure this is a
  zero overhead abstraction over the parse nodes.

## Scanner Generator

- Allow scanner to parse case independent (accept lower case and upper case characters if specified in one case only)

## Formater

- The GoLR formater should retain comments. Right now, comments are dropped because we parse the grammar file into
  a context-free grammar and regular expressions, then write those out again. As parsing the context-free grammar drops
  all comments, they are lost for writing out again. We need to look into mechanics to pass on dropped comments to the
  output.
