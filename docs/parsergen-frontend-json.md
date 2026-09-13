# Parser Generator Frontend: JSON

This frontend describes the context free grammar of a parser as a JSON document. See the data types in
`internal/parsergen/frontend` for details about the JSON structure.

## Example

The JSON input looks like this:

```json
{
  "terminals": [
    {
      "name": "PLUS"
    },
    {
      "name": "STAR",
      "alias": "\"*\"",
      "associativity": "left",
      "precedence": 1
    }
  ],
  "nonterminals": [
    {
      "name": "expr"
    },
    {
      "name": "term",
      "alias": "\"term\""
    }
  ],
  "productions": [
    {
      "nonterminalIdx": 0,
      "symbolRefs": [
        {
          "nonterminal": true,
          "index": 0
        },
        {
          "nonterminal": false,
          "index": 0
        },
        {
          "nonterminal": true,
          "index": 1
        }
      ]
    },
    {
      "nonterminalIdx": 1,
      "symbolRefs": [
        {
          "nonterminal": true,
          "index": 1
        },
        {
          "nonterminal": false,
          "index": 1
        },
        {
          "nonterminal": true,
          "index": 1
        }
      ],
      "precedenceTerminalIdx": 1
    }
  ],
  "startNonterminalIdx": 1
}
```

## Error Symbol

A terminal named `$error` is the error symbol, which marks the places in the grammar where the parser is
allowed to resume after a syntax error. Reference it from a production right-hand side like any other
terminal:

```json
{
  "terminals": [
    {
      "name": "$error"
    },
    {
      "name": "SEMI"
    }
  ]
}
```

It may be declared at any position in the terminal list, and a grammar which does not use error recovery
leaves it out entirely. No scanner ever produces it: the parser produces it itself while recovering. It
corresponds to `@error` in the GoLR format and to the `error` token of GNU Bison.

The leading dollar sign is what makes the name reserved, so an ordinary terminal must not be given it.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/frontend/json
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32          3,001      413,945 ns/op     109,298 B/op    1,278 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32               633    1,727,760 ns/op     426,299 B/op    5,020 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32     722    2,037,867 ns/op     490,587 B/op    6,551 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32              247    4,490,687 ns/op     938,916 B/op   12,054 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32             601    2,636,229 ns/op     654,525 B/op    7,068 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                   948    1,252,045 ns/op     371,682 B/op    4,510 allocs/op
BenchmarkToGrammar/PHP_8.6.7-32                  722    2,654,659 ns/op     728,500 B/op    8,232 allocs/op
BenchmarkToGrammar/PostgreSQL_18.4-32            100   15,953,931 ns/op   3,924,757 B/op   47,384 allocs/op
BenchmarkToGrammar/Ruby_3.2.11-32                411    3,231,359 ns/op     808,573 B/op    9,255 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/json     12.599s
```
