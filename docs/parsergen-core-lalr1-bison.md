# Parser Generator Core: LALR(1) Bison

This core generates an LALR(1) parser from a context free grammar.

The LALR(1) implementation delegates the parser generation to GNU Bison. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison v3 binary needs to be available
on your system for the LALR(1) Bison core to work.
