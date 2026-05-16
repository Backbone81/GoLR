# Scanner Generator Frontend: JSON

This frontend describes the regular expressions of a scanner as a JSON document. See the data types in
`internal/scannergen/frontend` for details about the JSON structure.

## Example

The JSON input looks like this:

```json
[
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
]
```
