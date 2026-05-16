# Scanner Generator Backend: JSON

This backend outputs a scanner as a JSON document. See the data types in `internal/scannergen/backend` for details about
the JSON structure.

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
          "charRange": {
            "low": 48,
            "high": 57
          },
          "stateIdx": 0
        },
        {
          "charRange": {
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
