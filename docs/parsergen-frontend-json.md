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

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/frontend/json
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32               3567            359375 ns/op          105541 B/op       1278 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                   786           1466872 ns/op          410241 B/op       5020 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32         614           1706725 ns/op          474533 B/op       6551 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                  328           3738118 ns/op          906438 B/op      12054 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                1040           2084589 ns/op          638473 B/op       7068 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                       967           1368425 ns/op          388373 B/op       4511 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/json     8.232s
```
