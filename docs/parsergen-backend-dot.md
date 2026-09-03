# Parser Generator Backend: DOT

This backend outputs a parser as a DOT document, for looking at the state machine while working on a grammar. See
[parser generator backends](parsergen-backend.md) for the other backends.

A node is a state, labelled with its index and its kernel items, each prefixed with its production's name (`@name` in
the grammar overrides the auto-generated name) and with a `•` marking how far the parse has come in the production. An
edge is a transition, labelled with the symbol it is taken on: solid for a shift on a terminal, dashed for a goto on a
nonterminal. Reductions and their lookaheads are not in the graph.

Render it with Graphviz:

```sh
golr parser --frontend golr --frontend-file-path calculator.golr --backend dot --backend-file-path parser.dot
dot -Tsvg parser.dot -o parser.svg
```
