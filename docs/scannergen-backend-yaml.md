# Scanner Generator Backend: YAML

This backend outputs a scanner as a YAML document. See the data types in `internal/scannergen/backend` for details about
the YAML structure.

## Example

The YAML output looks like this:

```yaml
rules:
- name: NUMBER
  regex:
    kind: Literal
    literal:
      text: "42"
- name: IDENT
  skip: true
  regex:
    kind: CharClass
    charClass:
      negate: true
      ranges:
      - low: 48
        high: 57
      - low: 97
        high: 122
states:
- ruleIdx: 0
- ruleIdx: 1
  accept: true
  transitions:
  - charRange:
      low: 48
      high: 57
    stateIdx: 0
  - charRange:
      low: 97
      high: 122
    stateIdx: 1
```
