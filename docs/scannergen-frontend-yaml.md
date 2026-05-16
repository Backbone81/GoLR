# Scanner Generator Frontend: YAML

This frontend describes the regular expressions of a scanner as a YAML document. See the data types in
`internal/scannergen/frontend` for details about the YAML structure.

## Example

The YAML input looks like this:

```yaml
- name: NUMBER
  regex:
    kind: Literal
    literal:
      text: "42"
- name: IDENT
  regex:
    kind: CharClass
    charClass:
      negate: true
      ranges:
      - low: 48
        high: 57
      - low: 97
        high: 122
```
