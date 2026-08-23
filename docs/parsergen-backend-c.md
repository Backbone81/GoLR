# Parser Generator Backend: C

This backend outputs a parser as C source code. The generated C code is a table driven parser which holds the parsing
decisions in lookup tables. It targets C99 and is a self contained header which carries its own implementation the same
way the generated scanner does, behind `<PREFIX>_PARSER_IMPLEMENTATION`.

The parse tree and the errors are owned by the result and are released together with `<prefix>_parse_result_free`. An
allocation which fails ends the parse and is reported as an error of kind `<PREFIX>_ERROR_KIND_OUT_OF_MEMORY`.

The parser uses the token type and the scanner of the generated scanner. Use `--backend-c-scanner-include` to set the
header it includes them from, which defaults to `scanner.h`, and `--backend-c-prefix` to set the prefix, which has to
be the one the scanner was generated with.
