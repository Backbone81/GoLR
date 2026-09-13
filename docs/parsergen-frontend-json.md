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
BenchmarkToGrammar/GNU_Bison_3.8.2-32               3001            413945 ns/op          109298 B/op       1278 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                   633           1727760 ns/op          426299 B/op       5020 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32         722           2037867 ns/op          490587 B/op       6551 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                  247           4490687 ns/op          938916 B/op      12054 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                 601           2636229 ns/op          654525 B/op       7068 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                       948           1252045 ns/op          371682 B/op       4510 allocs/op
BenchmarkToGrammar/PHP_8.6.7-32                      722           2654659 ns/op          728500 B/op       8232 allocs/op
BenchmarkToGrammar/PostgreSQL_18.4-32                100          15953931 ns/op         3924757 B/op      47384 allocs/op
BenchmarkToGrammar/Ruby_3.2.11-32                    411           3231359 ns/op          808573 B/op       9255 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/json     12.599s
```
