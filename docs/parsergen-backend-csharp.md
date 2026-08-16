# Parser Generator Backend: C#

This backend outputs a parser as C# source code. The generated C# code is a table driven parser which holds the parsing
decisions in lookup tables.

The parser uses the token constants and the scanner interface of the generated scanner. Use `--backend-csharp-namespace`
to set the namespace, which defaults to `Parser` and has to be the one the scanner was generated into.
