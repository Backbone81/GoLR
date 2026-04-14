# GoLR

GoLR is a modern tool for generating high-performance parsers based on LR(1) grammars. It combines the expressive power
of full LR(1) parsing with the efficiency of modern algorithms like
[IELR(1)](https://doi.org/10.1016/j.scico.2009.08.001), making it ideal for both interactive and production use.

For years, LR(1) grammars were seen as too resource-intensive compared to simpler LALR(1) approaches. However,
advancements like [IELR(1)](https://doi.org/10.1016/j.scico.2009.08.001) have changed the game, offering LALR(1)-like
performance without sacrificing the richness of LR(1). This tool brings those benefits to you in a highly modular and
extensible way.

The architecture of GoLR separates the frontend from the core and the backend. The frontend is responsible for reading
the grammar from different input formats. The core constructs the LR(1) parser from the grammar. The backend finally
outputs the parser into different output formats.

## Parser Generator

The parser generator constructs an LR(1) parser from a context free grammar. Please be aware of the known
[limitations](docs/limitations.md).

### Parser Generator Frontends

These frontends are currently supported:

- [JSON](docs/parsergen-frontend-json.md)
- [YAML](docs/parsergen-frontend-yaml.md)

Are you missing a frontend for your use case? Use the JSON frontend of GoLR to input the grammar as JSON and implement
your own frontend by loading whatever format you need and outputting the JSON. You do not need to do that in Go. Any
programming language which is able to load your format and can output JSON can be used for such a custom frontend. And
with outputting JSON to stdout, the output of your own frontend application can be piped into GoLR for maximum
flexibility.

### Parser Generator Cores

These cores are currently supported:

- [IELR(1)](docs/parsergen-core-ielr1.md)

### Parser Generator Backends

These backends are currently supported:

- [JSON](docs/parsergen-backend-json.md)
- [YAML](docs/parsergen-backend-yaml.md)
- [NULL](docs/parsergen-backend-null.md)

Are you missing a backend for your use case? Use the JSON backend of GoLR to output the parser as JSON and implement
your own backend by loading the JSON and outputting it in whatever format you need. You do not even need to do that
in Go. Any programming language which is able to load JSON can be used for such a custom backend. And with outputting
JSON to stdout, the output of GoLR can be piped into your own backend application for maximum flexibility.
