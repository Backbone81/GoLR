# Parser Generator Backend: JSON

This backend outputs a parser as a JSON document. See [parser generator backends](parsergen-backend.md) for the other
backends and for how to build one of your own on this output.

The document holds the grammar and the automaton as the core produced it: a state lists its kernel items, the
transitions it takes on a symbol, and the productions it reduces by together with the lookaheads which call for them.
The parse table is therefore uncompressed, unlike the packed lookup tables the language backends emit, which leaves a
consumer free to compress it whichever way suits its own target. The `grammar` member is what the
[JSON frontend](parsergen-frontend-json.md) reads, so a grammar taken out of this output can be fed back in. See the
data types in `internal/parsergen/backend` for details about the JSON structure.

## Example

The JSON output looks like this:

```json
{
  "grammar": {
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
  },
  "states": [
    {
      "kernelItems": [
        {
          "productionIdx": 0,
          "position": 0
        }
      ],
      "transitionActions": [
        {
          "symbolRef": {
            "nonterminal": false,
            "index": 0
          },
          "stateIdx": 1
        }
      ],
      "reduceActions": [
        {
          "lookaheadSet": [
            0
          ],
          "productionIdx": 0
        }
      ]
    },
    {
      "kernelItems": [
        {
          "productionIdx": 0,
          "position": 1
        },
        {
          "productionIdx": 1,
          "position": 2
        }
      ],
      "transitionActions": [
        {
          "symbolRef": {
            "nonterminal": false,
            "index": 1
          },
          "stateIdx": 0
        },
        {
          "symbolRef": {
            "nonterminal": true,
            "index": 0
          },
          "stateIdx": 0
        }
      ],
      "reduceActions": [
        {
          "lookaheadSet": [
            0,
            1
          ],
          "productionIdx": 0
        },
        {
          "lookaheadSet": [
            1
          ],
          "productionIdx": 1
        }
      ],
      "defaultReduceProductionIdx": 1
    }
  ]
}
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/json
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32              1608            687661 ns/op          111014 B/op       3073 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  169           6211822 ns/op         1333793 B/op      24759 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32        147           7582221 ns/op         2101428 B/op      35826 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  42          27222768 ns/op         8313694 B/op     124965 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                147           7962243 ns/op         2376128 B/op      38587 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      183           6430090 ns/op         1638429 B/op      29809 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      46          24310522 ns/op         8139051 B/op     111328 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 3         411708596 ns/op        172058906 B/op   2044996 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    50          23489561 ns/op         7495097 B/op     116584 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/json      11.116s
```
