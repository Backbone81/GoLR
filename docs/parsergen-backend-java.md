# Parser Generator Backend: Java

This backend outputs a parser as Java source code. The generated Java code is a table driven parser which holds the
parsing decisions in lookup tables.

The parser uses the token constants and the scanner interface of the generated scanner. Use
`--backend-java-package-name` to set the package, which defaults to `parser` and has to be the one the scanner was
generated into.
