# GNU Bison Grammar Parser

This example demonstrates how to generate a scanner and parser for processing GNU Bison grammar files. It can then
be used to list all tokens and display a parse tree of such a grammar.

The `spec` folder contains a description of tokens in the GoLR format alongside the official GNU Bison grammar used by
GNU Bison itself.

The spec is then transformed into a scanner and parser in the `parser` folder.

As the GNU Bison grammar has some specialities, a few manual additions are needed for the scanner and parser to work
correctly.

Note that this example is basically a copy of the internal GoLR GNU Bison frontend.
