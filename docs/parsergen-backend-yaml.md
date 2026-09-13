# Parser Generator Backend: YAML

This backend outputs a parser as a YAML document. See [parser generator backends](parsergen-backend.md) for the other
backends and for how to build one of your own on this output.

The document holds the grammar and the automaton as the core produced it: a state lists its kernel items, the
transitions it takes on a symbol, and the productions it reduces by together with the lookaheads which call for them.
The parse table is therefore uncompressed, unlike the packed lookup tables the language backends emit, which leaves a
consumer free to compress it whichever way suits its own target. The `grammar` member is what the
[YAML frontend](parsergen-frontend-yaml.md) reads, so a grammar taken out of this output can be fed back in. See the
data types in `internal/parsergen/backend` for details about the YAML structure.

## Example

The YAML output looks like this:

```yaml
grammar:
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
states:
- kernelItems:
  - productionIdx: 0
    position: 0
  transitionActions:
  - symbolRef:
      nonterminal: false
      index: 0
    stateIdx: 1
  reduceActions:
  - lookaheadSet:
    - 0
    productionIdx: 0
- kernelItems:
  - productionIdx: 0
    position: 1
  - productionIdx: 1
    position: 2
  transitionActions:
  - symbolRef:
      nonterminal: false
      index: 1
    stateIdx: 0
  - symbolRef:
      nonterminal: true
      index: 0
    stateIdx: 0
  reduceActions:
  - lookaheadSet:
    - 0
    - 1
    productionIdx: 0
  - lookaheadSet:
    - 1
    productionIdx: 1
  defaultReduceProductionIdx: 1
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/yaml
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32                28          67976296 ns/op        18636833 B/op     427917 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                    5         236457150 ns/op        164133075 B/op   3718052 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32          4         308486673 ns/op        240180888 B/op   5431046 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                   2         790850773 ns/op        878787828 B/op  19788733 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                  4         307934227 ns/op        263820460 B/op   5954990 allocs/op
BenchmarkFromParser/Go_1.5.4-32                        5         233771618 ns/op        202475723 B/op   4557011 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                       2         707271355 ns/op        795512872 B/op  17820022 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 1        11334317425 ns/op       15334454384 B/op        342128952 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                     2         734807645 ns/op        834694188 B/op  18705416 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/yaml      23.463s
```
