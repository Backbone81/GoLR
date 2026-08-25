# Scanner Generator Backend: DOT

This backend outputs a scanner as a DOT document, for looking at the DFA while working on a set of rules. See
[scanner generator backends](scannergen-backend.md) for the other backends.

A node is a state of the automaton, drawn as a double circle when it accepts a rule, and the arrow with no source marks
the state a scan starts in. An edge is a transition, labelled with the characters it is taken on.

Render it with Graphviz:

```sh
golr scanner --frontend golr --frontend-file-path calculator.golr --backend dot --backend-file-path scanner.dot
dot -Tsvg scanner.dot -o scanner.svg
```
