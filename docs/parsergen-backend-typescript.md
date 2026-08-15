# Parser Generator Backend: TypeScript

This backend outputs a parser as TypeScript source code. The generated TypeScript code is a table driven parser which
holds the parsing decisions in lookup tables.

The parser imports the token constants from the generated scanner. Use `--backend-typescript-scanner-module` to set the
module specifier, which defaults to `./scanner.js`.
