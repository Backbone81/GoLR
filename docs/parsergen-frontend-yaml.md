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

## Error Symbol

A terminal named `$error` is the error symbol, which marks the places in the grammar where the parser is
allowed to resume after a syntax error. Reference it from a production right-hand side like any other
terminal:

```yaml
terminals:
- name: $error
- name: SEMI
```

It may be declared at any position in the terminal list, and a grammar which does not use error recovery
leaves it out entirely. No scanner ever produces it: the parser produces it itself while recovering. It
corresponds to `@error` in the GoLR format and to the `error` token of GNU Bison.

The leading dollar sign is what makes the name reserved, so an ordinary terminal must not be given it.

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/frontend/yaml
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkToGrammar/GNU_Bison_3.8.2-32                 13         127159859 ns/op        43480624 B/op      71008 allocs/op
BenchmarkToGrammar/GCC_2.95.3_C-32                     1        1661144907 ns/op        783196248 B/op    330789 allocs/op
BenchmarkToGrammar/GCC_2.95.3_Objective_C-32           1        2195132424 ns/op        1022467160 B/op   432960 allocs/op
BenchmarkToGrammar/GCC_3.3.6_C++-32                    1        4900333068 ns/op        3785718672 B/op   934715 allocs/op
BenchmarkToGrammar/GCC_4.2.4_Java-32                   1        2460520581 ns/op        1106144848 B/op   464361 allocs/op
BenchmarkToGrammar/Go_1.5.4-32                         1        1624526628 ns/op        701544384 B/op    298087 allocs/op
BenchmarkToGrammar/PHP_8.6.7-32                        1        4084674560 ns/op        2466333696 B/op   631357 allocs/op
BenchmarkToGrammar/PostgreSQL_18.4-32                  1        37437390394 ns/op       59118112624 B/op         6949441 allocs/op
BenchmarkToGrammar/Ruby_3.2.11-32                      1        4263318820 ns/op        2778751224 B/op   708313 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/frontend/yaml     60.916s
```
