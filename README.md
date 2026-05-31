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

The generated parser constructs an abstract syntax tree which you can then walk and execute semantic actions
accordingly.

For more details about how this project came to be, see the documentation about [motivation](docs/motivation.md).

## Getting Started

Install the GoLR binary either with your Go toolchain:

```shell
go install github.com/backbone81/golr/cmd/golr@latest
```

Or download a prebuilt binary from the releases section and make it available in your shell.

**IMPORTANT: The current [IELR(1)](https://doi.org/10.1016/j.scico.2009.08.001) core uses GNU Bison in the background to do the parser construction, make sure
that a recent version of GNU Bison v3 is available in your shell.**

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

You can then walk the abstract syntax tree from the root node.

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
      --backend string                   The backend to use for writing the parser. One of: dot, go, json, null, yaml. (default "go")
      --backend-file-path string         The file path to write the parser to. Can be '-' to write to stdout.
      --backend-go-package-name string   The Go package name to use for the generated Go code. (default "parser")
      --core string                      The core to use for generating the parser from the context free grammar. One of: ielr1. (default "ielr1")
      --frontend string                  The frontend to use for reading the context free grammar. One of: bison, golr, json, yaml. (default "golr")
      --frontend-file-path string        The file path to read the context free grammar from. Can be '-' to read from stdin.
  -h, --help                             help for parser
```

The `scanner` sub-command allows for selecting frontend, core and backend for the scanner:

```text
Generates a DFA scanner.

Usage:
  golr scanner [flags]

Flags:
      --backend string                   The backend to use for writing the scanner. One of: dot, go, json, null, yaml. (default "go")
      --backend-file-path string         The file path to write the scanner to. Can be '-' to write to stdout.
      --backend-go-package-name string   The Go package name to use for the generated Go code. (default "parser")
      --core string                      The core to use for generating the scanner from the regular expressions. One of: subset. (default "subset")
      --frontend string                  The frontend to use for reading the regular expressions. One of: golr, json, yaml. (default "golr")
      --frontend-file-path string        The file path to read the regular expressions from. Can be '-' to read from stdin.
  -h, --help                             help for scanner
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

### Parser Generator Backends

These backends are currently supported:

- [DOT](docs/parsergen-backend-dot.md)
- [Go](docs/parsergen-backend-golang.md)
- [JSON](docs/parsergen-backend-json.md)
- [Null](docs/parsergen-backend-null.md)
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

- [DOT](docs/scannergen-backend-dot.md)
- [Go](docs/scannergen-backend-golang.md)
- [JSON](docs/scannergen-backend-json.md)
- [Null](docs/scannergen-backend-null.md)
- [YAML](docs/scannergen-backend-yaml.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the scanner as JSON and implement
your own backend by loading the JSON and output it in whatever format you need. You do not even need to do that
in Go. Any programming language which is able to load JSON can be used for such a custom backend. And with outputting
JSON to stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.

## Roadmap

See the [roadmap](docs/roadmap.md) for topics which will be addressed in the future.
