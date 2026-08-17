# GoLR

GoLR is a modern tool for generating high-performance parsers based on LR(1) grammars. It combines the expressive power
of full LR(1) parsing with the efficiency of modern algorithms like
[IELR(1)](https://doi.org/10.1016/j.scico.2009.08.001), making it ideal for both interactive and production use.

For years, LR(1) grammars were seen as too resource-intensive compared to simpler LALR(1) approaches. However,
advancements like [IELR(1)](https://doi.org/10.1016/j.scico.2009.08.001) have changed the game, offering LALR(1)-like
performance without sacrificing the richness of LR(1). This tool brings those benefits to you in a highly modular and
extensible way.

The architecture of GoLR separates the frontend from the core and the backend. The frontend is responsible for reading
the context free grammar from different input formats. The core constructs the LR(1) parser from the grammar. The
backend finally outputs the parser into different output formats.

The generated parser constructs a parse tree which you can then walk and execute semantic actions
accordingly.

For more details about how this project came to be, see the documentation about [motivation](docs/motivation.md). For
details about how GoLR makes sure that the generated parser tables are right, see the documentation about
[correctness](docs/correctness.md).

## Getting Started

Install the GoLR binary either with your Go toolchain:

```shell
go install github.com/backbone81/golr/cmd/golr@latest
```

Or download a prebuilt binary from the releases section and make it available in your shell.

This example assumes a context free grammar in a GNU Bison grammar file `grammar.y`. Run GoLR to generate a Go parser
from it:

```sh
golr parser \
  --frontend bison \
  --frontend-file-path grammar.y \
  --backend-file-path parser/parser.go
```

This generates a `parser/parser.go` file in the default parser package.

You then need to provide a scanner that produces tokens for the parser. The generated ParserScanner interface documents
what the parser expects from the scanner. You can use the GoLR scanner generator to generate a scanner for you. Note
that the scanner tokens need to be defined in the same package as the parser. Otherwise, the parser will reference
tokens which do not exist.

Once you have a scanner, parsing works like this:

```go
scanner := parser.NewScanner() // your scanner implementation

p := parser.NewParser()
rootNode, err := p.Parse(scanner)
if err != nil {
    log.Fatal(err)
}
```

You can then walk the parse tree from the root node.

## Examples

See the [Calculator Example](examples/calculator/README.md) for a simple and complete example about how to use GoLR.

See the `examples` directory for parsers generated with GoLR.

## Command Line Parameters

The GoLR CLI supports several command line parameters. Use `--help` for a help screen.

The main top level sub-commands are `parser` to generate an LR(1) parser and `scanner` to generate an DFA scanner:

```text
GoLR is a parser generator for LR(1) grammars.

Usage:
  golr [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  convert     Converts grammar files between different formats.
  fmt         Pretty prints GoLR grammar files.
  help        Help about any command
  parser      Generates a LR(1) parser.
  scanner     Generates a DFA scanner.
  selftest    Checks the IELR(1) parser core against a canonical LR(1) oracle.

Flags:
  -h, --help   help for golr

Use "golr [command] --help" for more information about a command.
```

The `parser` sub-command allows selecting frontend, core and backend for the parser:

```text
Generates a LR(1) parser.

Usage:
  golr parser [flags]

Flags:
      --backend string                   The backend to use for writing the parser. One of: dot, go, go-direct, go-table, json, null, yaml. (default "go")
      --backend-file-path string         The file path to write the parser to. Can be '-' to write to stdout.
      --backend-go-package-name string   The Go package name to use for the generated Go code. (default "parser")
      --core string                      The core to use for generating the parser from the context free grammar. One of: ielr1, ielr1-golr, ielr1-bison, lalr1, lalr1-golr, lalr1-bison, lr1, lr1-golr, lr1-bison. (default "ielr1")
      --frontend string                  The frontend to use for reading the context free grammar. One of: bison, golr, json, yaml. (default "golr")
      --frontend-file-path string        The file path to read the context free grammar from. Can be '-' to read from stdin.
  -h, --help                             help for parser
  -v, --verbose                          List every conflict the parser generator resolved on its own, instead of only summarizing them.
```

The `scanner` sub-command allows for selecting frontend, core and backend for the scanner:

```text
Generates a DFA scanner.

Usage:
  golr scanner [flags]

Flags:
      --backend string                     The backend to use for writing the scanner. One of: dot, go, go-direct, go-table, java, json, null, rust, yaml. (default "go")
      --backend-file-path string           The file path to write the scanner to. Can be '-' to write to stdout.
      --backend-go-package-name string     The Go package name to use for the generated Go code. (default "parser")
      --backend-java-package-name string   The Java package name to use for the generated Java code. (default "parser")
      --core string                        The core to use for generating the scanner from the regular expressions. One of: subset. (default "subset")
      --frontend string                    The frontend to use for reading the regular expressions. One of: golr, json, yaml. (default "golr")
      --frontend-file-path string          The file path to read the regular expressions from. Can be '-' to read from stdin.
  -h, --help                               help for scanner
```

The `fmt` sub-command allows to pretty print GoLR grammar files:

```text
Pretty prints GoLR grammar files. All comments will be removed.

Usage:
  golr fmt [file...] [flags]

Flags:
  -h, --help   help for fmt
```

Note that due to limitations in the current implementation, the pretty printer for GoLR grammar files currently drops
all comments from the file. This will be fixed in the future.

The `convert` sub-command converts GNU Bison grammar files to GoLR grammar files:

```text
Converts grammar files between different formats.

Usage:
  golr convert [flags]

Flags:
  -h, --help                      help for convert
      --input-file-path string    The GNU Bison grammar file to convert. Can be '-' to read from stdin.
      --output-file-path string   The GoLR grammar file to write. Can be '-' to write to stdout. (default "-")
```

Note that the conversion is incomplete most of the time. GNU Bison grammar files do not describe regular expressions
of tokens, for example. But the conversion can be a starting point to have a GoLR grammar quickly with only a few
manual corrections needed.

The `selftest` sub-command checks the IELR(1) parser core against a canonical LR(1) oracle:

```text
Checks the IELR(1) parser core against a canonical LR(1) oracle.

Usage:
  golr selftest [flags]

Flags:
      --duration duration                          How long to keep checking grammars, for example 30m or 8h. Zero means no time limit.
      --failure-dir string                         The directory to dump a failing grammar and its action traces into. Empty only reports the seed.
      --grammar-count int                          The number of grammars to check in total. Zero keeps checking until the duration is up or the run is interrupted.
  -h, --help                                       help for selftest
      --inputs-per-grammar int                     The number of generated sentences each grammar is checked with. (default 16)
      --max-nonterminal-count int                  The largest number of nonterminals a generated grammar may have. Zero uses the generator default.
      --max-production-count-per-nonterminal int   The largest number of productions a generated nonterminal may have. Zero uses the generator default.
      --max-rhs-symbol-count int                   The largest number of symbols on the right hand side of a generated production. Zero uses the generator default.
      --max-terminal-count int                     The largest number of terminals a generated grammar may have. Zero uses the generator default.
      --progress-interval duration                 How often to print a progress line. (default 10s)
      --stop-on-failure                            End the whole run as soon as one grammar fails, instead of counting the failure and carrying on.
      --workers int                                The number of grammars to check concurrently. Defaults to the number of CPU cores.
```

This is a soak test for GoLR itself, not a step of generating a parser. Random grammars are turned into an IELR(1) and
a canonical LR(1) parser table, and both tables are driven through the same generated sentences. They have to take the
identical sequence of LR actions every time, because that is what IELR(1) guarantees: the same language and the same
parses as canonical LR(1), only with fewer states. Any disagreement is a bug in the IELR(1) core.

A run saturates every core it is given and keeps going until the grammar count or the duration is reached, so a run
with neither only ends on Ctrl-C. Interrupting it is fine at any point, the summary still covers everything checked so
far:

```shell
golr selftest --duration 8h --failure-dir ./selftest-failures | tee selftest.log
```

See the documentation about [correctness](docs/correctness.md) for where this fits into the overall verification of the
IELR(1) implementation.

## Parser Generator

The parser generator constructs an LR(1) parser from a context free grammar. Please be aware of the known
[limitations](docs/limitations.md).

### Parser Generator Frontends

These frontends are currently supported:

- [Bison](docs/parsergen-frontend-bison.md)
- [GoLR](docs/parsergen-frontend-golr.md)
- [DSL](docs/parsergen-frontend-dsl.md)
- [JSON](docs/parsergen-frontend-json.md)
- [YAML](docs/parsergen-frontend-yaml.md)

Are you missing a frontend for your use case? Use the JSON frontend of GoLR to input the grammar as JSON and implement
your own frontend by loading whatever format you need and output the JSON. You do not need to do that in Go. Any
programming language which is able to load your format and can output JSON can be used for such a custom frontend. And
with outputting JSON to stdout, the output of your own frontend application can be piped into GoLR for maximum
flexibility.

### Parser Generator Cores

These cores are currently supported:

- [IELR(1)](docs/parsergen-core-ielr1.md)
- [IELR(1) GoLR](docs/parsergen-core-ielr1-golr.md)
- [IELR(1) Bison](docs/parsergen-core-ielr1-bison.md)
- [LALR(1)](docs/parsergen-core-lalr1.md)
- [LALR(1) GoLR](docs/parsergen-core-lalr1-golr.md)
- [LALR(1) Bison](docs/parsergen-core-lalr1-bison.md)
- [LR(1)](docs/parsergen-core-lr1.md)
- [LR(1) GoLR](docs/parsergen-core-lr1-golr.md)
- [LR(1) Bison](docs/parsergen-core-lr1-bison.md)

### Parser Generator Backends

These backends are currently supported:

- [C#](docs/parsergen-backend-csharp.md)
- [DOT](docs/parsergen-backend-dot.md)
- [Go](docs/parsergen-backend-golang.md)
- [Go Direct](docs/parsergen-backend-golang-direct.md)
- [Go Table](docs/parsergen-backend-golang-table.md)
- [Java](docs/parsergen-backend-java.md)
- [JavaScript](docs/parsergen-backend-javascript.md)
- [JSON](docs/parsergen-backend-json.md)
- [Null](docs/parsergen-backend-null.md)
- [Rust](docs/parsergen-backend-rust.md)
- [TypeScript](docs/parsergen-backend-typescript.md)
- [YAML](docs/parsergen-backend-yaml.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the parser as JSON and implement
your own backend by loading the JSON and output it in whatever format you need. You do not even need to do that
in Go. Any programming language which is able to load JSON can be used for such a custom backend. And with outputting
JSON to stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.

## Scanner Generator

The scanner generator constructs a DFA scanner from regular expressions.

### Scanner Generator Frontends

These frontends are currently supported:

- [DSL](docs/scannergen-frontend-dsl.md)
- [GoLR](docs/scannergen-frontend-golr.md)
- [JSON](docs/scannergen-frontend-json.md)
- [YAML](docs/scannergen-frontend-yaml.md)

Are you missing a frontend for your use case? Use the JSON frontend of GoLR to input the regular expressions as JSON and
implement your own frontend by loading whatever format you need and output the JSON. You do not need to do that in Go.
Any programming language which is able to load your format and can output JSON can be used for such a custom frontend.
And with outputting JSON to stdout, the output of your own frontend application can be piped into GoLR for maximum
flexibility.

### Scanner Generator Cores

These cores are currently supported:

- [Subset](docs/scannergen-core-subset.md)

### Scanner Generator Backends

These backends are currently supported:

- [C#](docs/scannergen-backend-csharp.md)
- [DOT](docs/scannergen-backend-dot.md)
- [Go](docs/scannergen-backend-golang.md)
- [Go Direct](docs/scannergen-backend-golang-direct.md)
- [Go Table](docs/scannergen-backend-golang-table.md)
- [Java](docs/scannergen-backend-java.md)
- [JavaScript](docs/scannergen-backend-javascript.md)
- [JSON](docs/scannergen-backend-json.md)
- [Null](docs/scannergen-backend-null.md)
- [Rust](docs/scannergen-backend-rust.md)
- [TypeScript](docs/scannergen-backend-typescript.md)
- [YAML](docs/scannergen-backend-yaml.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the scanner as JSON and implement
your own backend by loading the JSON and output it in whatever format you need. You do not even need to do that
in Go. Any programming language which is able to load JSON can be used for such a custom backend. And with outputting
JSON to stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.

## Correctness

A parser generator fails quietly. A reduction lookahead set which is one terminal too large turns a perfectly good
grammar into one with a conflict, and one which is one terminal too small produces a parser that builds, runs and then
rejects a sentence the grammar clearly derives. Nothing crashes, and the damage surfaces only in whoever uses the
generated parser. IELR(1) makes this worse: it is a five phase algorithm whose output is deliberately not comparable to
any table you could write down by hand, so "diff it against the expected result" is not available as a test strategy.

What IELR(1) does guarantee is behavioral — an IELR(1) parser accepts the same language and produces the same parses as
a canonical LR(1) parser under the same conflict resolution policy, only with fewer states. GoLR builds its verification
on that guarantee, in overlapping layers:

- **The grammars from the [IELR(1) paper](https://doi.org/10.1016/j.scico.2009.08.001).** Parser tables, follow kernel
  items, annotations and item lookahead sets are pinned against the definitions of the paper for the small grammars its
  figures were constructed from.
- **Real-world grammars cross-checked against GNU Bison.** The grammars of GNU Bison, GCC's C, Objective-C, C++ and
  Java, Go, PHP and PostgreSQL are each built with the GoLR LALR(1) and IELR(1) cores and with GNU Bison itself — the
  reference implementation whose authors wrote the paper. Where a grammar is LALR(1), all four tables must agree on the
  state count. Where it is not, both implementations must split, and GoLR must land within 2% of Bison's state count.
- **Differential testing against canonical LR(1).** Random grammars, generated from scenarios deliberately biased
  toward the shapes where LALR(1) and canonical LR(1) diverge, are turned into an IELR(1) and a canonical LR(1) table.
  Both tables are then driven through sentences derived from the grammar itself and have to take the identical sequence
  of LR actions, step for step. Every run additionally asserts the size invariant
  `|LALR(1)| <= |IELR(1)| <= |canonical LR(1)|`, and the corpus measures itself so it cannot pass vacuously on grammars
  which never exercise the splitting.
- **The `golr selftest` soak test.** The same comparison, running across every CPU core for hours instead of seconds.
  Corpora of millions of grammars are routine, and a single seed reproduces any failure it finds.
- **Mutation testing of the test suite itself.** Around 50 deliberate bugs, each derived from a specific definition in
  the paper, were injected one at a time to confirm the self-test actually notices. Every mutation which changes a parse
  was detected, most within a few dozen grammars.

The [correctness](docs/correctness.md) documentation describes each layer in detail, including what these checks
deliberately do not cover.

## Roadmap

See the [roadmap](docs/roadmap.md) for topics which will be addressed in the future.

## License

GoLR is licensed under the [Apache License, Version 2.0](LICENSE).

The parsers and scanners GoLR generates for you are exempt from that license. The
[GoLR Output Exception](LICENSE.OUTPUT) gives you unlimited permission to use, modify and distribute the generated
files under terms of your choosing. You do not need to place them under the Apache License, ship a copy of the license
with them, or attribute GoLR in them.

The exception covers the generated output only. GoLR itself, including its code-generation templates, stays under the
Apache License.
