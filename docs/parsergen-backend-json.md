# Parser Generator Backend: JSON

This backend outputs a parser as a JSON document. See the data types in `internal/parsergen/backend` for details about
the JSON structure.

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
BenchmarkFromParser/GNU_Bison_3.8.2-32              1255            885990 ns/op          146567 B/op       3529 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  141           8097676 ns/op         2442451 B/op      37033 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32        100          11209389 ns/op         3296190 B/op      53183 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  25          47505688 ns/op        18384436 B/op     249642 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                 81          14199372 ns/op         5174323 B/op      70614 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      114           9586628 ns/op         2973308 B/op      45364 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/json      7.661s
```
