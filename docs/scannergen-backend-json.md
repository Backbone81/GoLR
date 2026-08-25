# Scanner Generator Backend: JSON

This backend outputs a scanner as a JSON document. See [scanner generator backends](scannergen-backend.md) for the
other backends and for how to build one of your own on this output.

The document holds the rules and the DFA as the core constructed it: a state says which rule it accepts, if any, and
lists its transitions as byte ranges. The automaton is therefore uncompressed, unlike the byte classes and packed
lookup tables the language backends emit, which leaves a consumer free to compress it whichever way suits its own
target. The `rules` member is what the [JSON frontend](scannergen-frontend-json.md) reads, so the rules taken out of
this output can be fed back in. See the data types in `internal/scannergen/backend` for details about the JSON
structure.

## Example

The JSON output looks like this:

```json
{
  "rules": [
    {
      "name": "NUMBER",
      "regex": {
        "kind": "Literal",
        "literal": {
          "text": "42"
        }
      }
    },
    {
      "name": "IDENT",
      "skip": true,
      "regex": {
        "kind": "CharClass",
        "charClass": {
          "negate": true,
          "ranges": [
            {
              "low": 48,
              "high": 57
            },
            {
              "low": 97,
              "high": 122
            }
          ]
        }
      }
    }
  ],
  "states": [
    {
      "ruleIdx": 0
    },
    {
      "ruleIdx": 1,
      "accept": true,
      "transitions": [
        {
          "byteRange": {
            "low": 48,
            "high": 57
          },
          "stateIdx": 0
        },
        {
          "byteRange": {
            "low": 97,
            "high": 122
          },
          "stateIdx": 1
        }
      ]
    }
  ]
}
```
