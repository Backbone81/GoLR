# Roadmap

## General Topics

- Fix the naming of "golang" in the documentation. We should use Go where possible. Use only golang when it collides
  in go code with the reserved keyword.
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
- Make the table driven Go backend the default and drop the directly coded one, once the table driven backend has
  proven itself. It is already faster and much smaller, so the remaining question is whether anything is lost by giving
  up the readable per state functions of the directly coded backend.
- Add a table driven backend for Java and Rust as well, replacing their directly coded ones. The tables are language
  neutral, so each of those is a small driver instead of another encoding of the automaton.
- Introduce a test harness to test drive all generated scanners in their native language.

## Formater

- The GoLR formater should retain comments. Right now, comments are dropped because we parse the grammar file into
  a context-free grammar and regular expressions, then write those out again. As parsing the context-free grammar drops
  all comments, they are lost for writing out again. We need to look into mechanics to pass on dropped comments to the
  output.
