# Parser Generator Backend: YAML

This backend outputs a parser as a YAML document. See the data types in `internal/parsergen/backend` for details about
the YAML structure.

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
BenchmarkFromParser/GNU_Bison_3.8.2-32                34          78649496 ns/op        22390411 B/op     510692 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                    3         342472295 ns/op        253107877 B/op   5699288 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32          3         397419209 ns/op        365594642 B/op   8236027 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                   1        1442258120 ns/op       1783031000 B/op  40054543 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                  3         460239385 ns/op        496505978 B/op  11179925 allocs/op
BenchmarkFromParser/Go_1.5.4-32                        3         343616448 ns/op        315817293 B/op   7090116 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/yaml      9.664s
```
