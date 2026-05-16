# Parser Generator Frontend: YAML

This frontend describes the context free grammar of a parser as YAML document. See the data types in
`internal/parsergen/frontend` for details about the YAML structure.

## Example

The YAML input looks like this:

```yaml
terminals:
- name: PLUS
- name: STAR
  alias: "\"*\""
  associativity: left
  precedence: 1
nonterminals:
- name: expr
- name: term
  alias: "\"term\""
productions:
- nonterminalIdx: 0
  symbolRefs:
  - nonterminal: true
    index: 0
  - nonterminal: false
    index: 0
  - nonterminal: true
    index: 1
- nonterminalIdx: 1
  symbolRefs:
  - nonterminal: true
    index: 1
  - nonterminal: false
    index: 1
  - nonterminal: true
    index: 1
  precedenceTerminalIdx: 1
startNonterminalIdx: 1
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/frontend/yaml
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                 22          52785697 ns/op        43458478 B/op     70649 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                     2        1226356110 ns/op       783261032 B/op    329970 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32           1        2093521358 ns/op      1022589232 B/op    431555 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                    1        5743500257 ns/op      3785791800 B/op    932319 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                   1        2466666292 ns/op      1105932032 B/op    462939 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                         1        1413885689 ns/op       701657528 B/op    297750 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/yaml     15.611s
```
