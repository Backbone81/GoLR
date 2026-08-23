# Scanner Generator Backend: C

This backend outputs a scanner as C source code. The generated C code is a table driven scanner which holds the
automaton in lookup tables. It targets C99, is a self contained header and includes nothing beyond the standard
library.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. In case of conflicts between rules,
the rule which was specified earlier has priority over those rules specified later.

The header carries its own implementation: include it wherever the scanner is used, and define
`<PREFIX>_SCANNER_IMPLEMENTATION` before including it in exactly one translation unit, which is where the tables and
the function bodies are emitted.

Use `--backend-c-prefix` to set the prefix every generated name carries, which defaults to `parser`.
