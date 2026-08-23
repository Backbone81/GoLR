# Roadmap

## General Topics

- Fix the naming of "golang" in the documentation. We should use Go where possible. Use only golang when it collides
  in go code with the reserved keyword.
- Extend the documentation.
- Add benchmarks to all documentation.
- Add a benchmark mode which runs with the garbage collector disabled, by setting `GOGC=off` together with a
  `GOMEMLIMIT` high enough that it never triggers, as a make target next to the normal one. Measured on the Go grammar,
  a parse takes 8.9 ms with the collector on and 2.0 ms with it off, so roughly three quarters of the runtime is
  collection - the parse allocates about 1.8 GB/s, far more than the collector keeps up with, so the parsing goroutine
  is pulled into the collection itself through mutator assist. That cost is identical for every backend and swamps what
  a benchmark is meant to compare: the run to run spread drops from 26 % to about 2 % with the collector off, which
  turned a difference between the two Go parser backends that could not be quoted at all into a clear 23 %. Set the
  limit through the environment rather than in code, so the committed benchmark keeps measuring what a user actually
  gets, and discard the first few samples, which are consistently high.
- Publish Visual Studio Code extension for syntax highlighting of GoLR files
- Publish IntelliJ plugin for syntax highlighting of GoLR files
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
- Report nonterminals and productions which are useless, meaning they cannot be reached from the start symbol. GNU
  Bison warns with "nonterminal useless in grammar" and lists them in its report; we have no equivalent, and
  `RemoveUnreachableStates` covers states rather than grammar symbols. This is a diagnostic only and changes no
  generated output, so it does not have to wait for the backend switch. `FallthroughStmt` in `examples/golang` was such
  a symbol, and it went unnoticed until it was found by reading the grammar.

## Scanner Generator

- Allow scanner to parse case independent (accept lower case and upper case characters if specified in one case only)
- Introduce a test harness to test drive all generated scanners in their native language.

## Formater

- The GoLR formater should retain comments. Right now, comments are dropped because we parse the grammar file into
  a context-free grammar and regular expressions, then write those out again. As parsing the context-free grammar drops
  all comments, they are lost for writing out again. We need to look into mechanics to pass on dropped comments to the
  output.
