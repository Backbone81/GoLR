# Parser Generator Backend: C++

This backend outputs a parser as C++ source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C++](scannergen-backend-cpp.md) for the scanner side.

It targets C++17 and is a self contained header.

`--backend-cpp-namespace` sets the namespace, which defaults to `parser` and has to be the one the scanner was
generated into. `--backend-cpp-scanner-include` sets the header the token type is included from, which defaults to
`scanner.hpp`.

`Parser::parse` returns a `ParseResult` holding the tree and the errors, and can be called again with a different
scanner. It is a template over the scanner type rather than a function taking an interface, so any type with the
members the generated scanner has will do. The tree holds `std::string_view` lexemes into the source, which therefore
has to outlive it.

Every nonterminal node's `production` field names the alternative it was reduced by, as one of the generated
`Production` enumerators (`ProductionExpression1`, ... - `@name` in the grammar overrides the auto-generated name). It
is `std::nullopt` on a terminal node.

## Tracing

`Parser::set_trace` takes a `TraceFunc` (`std::function<void(std::string_view)>`), called with one line for every
parser action. Set it after construction; pass an empty `std::function` for no tracing. See
[parser generator backends](parsergen-backend.md) for the line format.

## Example

[examples/calculator/cpp/](../examples/calculator/cpp/) is a calculator built on this backend. Its parser was generated
with:

```sh
golr parser \
  --frontend golr \
  --frontend-file-path calculator.golr \
  --backend cpp \
  --backend-cpp-namespace calculator::parser \
  --backend-file-path parser/parser.hpp
```

## Benchmarks

```text
goos: linux
goarch: amd64
pkg: github.com/backbone81/golr/internal/parsergen/backend/cpp
cpu: Intel(R) Core(TM) i9-14900K
BenchmarkFromParser/GNU_Bison_3.8.2-32               608           1716551 ns/op          536522 B/op       8618 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32                  180           6291328 ns/op         2763803 B/op      29858 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32        124           9097340 ns/op         4601872 B/op      41457 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32                  43          29006969 ns/op        11603946 B/op      81681 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32                100          10326418 ns/op         4622337 B/op      44475 allocs/op
BenchmarkFromParser/Go_1.5.4-32                      252           5698106 ns/op         2697234 B/op      28926 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                      63          17015554 ns/op         8478942 B/op      60665 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32                 2         540262297 ns/op        138290256 B/op    490540 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32                    69          19928512 ns/op         9811740 B/op      62891 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/cpp       11.391s
```
