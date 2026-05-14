# Parser Generator Core: IELR(1)

This core generates an LR(1) parser from a context free grammar. It applies the IELR(1) algorithm as described in the paper
["The IELR(1) algorithm for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" by Joel E. Denny and Brian A. Malloy](https://doi.org/10.1016/j.scico.2009.08.001).

The IELR(1) implementation delegates the parser generation to GNU Bison for now. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison binary needs to be available
on your system for the IELR(1) core to work. While this is a simple way to make IELR(1) quickly available for GoLR, the
long term goal is to provide an IELR(1) implementation which is written natively in Go.
