# Parser Generator Frontend: DSL

This frontend describes the context free grammar of a parser in a domain specific language with Go. Import
`golr/pkg/parsergen/frontend/dsl` and use the functions and data types of that package to describe the rules of your
parser.

## Error Symbol

A terminal named `$error` is the error symbol, which marks the places in the grammar where the parser is
allowed to resume after a syntax error. Declare it like any other terminal and use it on a production right-hand
side:

```go
errorSymbol := grammar.Terminal("$error")
```

It may be declared at any point, and a grammar which does not use error recovery leaves it out entirely. No
scanner ever produces it: the parser produces it itself while recovering. It corresponds to `@error` in the
GoLR format and to the `error` token of GNU Bison.

The leading dollar sign is what makes the name reserved, so an ordinary terminal must not be given it.
