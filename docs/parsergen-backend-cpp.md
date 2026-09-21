# Parser Generator Backend: C++

This backend outputs a parser as C++ source code. See [parser generator backends](parsergen-backend.md) for what every
generated parser does and [scanner generator backend: C++](scannergen-backend-cpp.md) for the scanner side.

It targets C++17 and is a self contained header.

`--backend-cpp-namespace` sets the namespace, which defaults to `parser` and has to be the one the scanner was
generated into. `--backend-cpp-scanner-include` sets the header the token type is included from, which defaults to
`scanner.hpp`.

`Parser::parse` returns a `ParseResult` holding the tree and the errors, and can be called again with a different
scanner. It is a template over the scanner type rather than a function taking an interface, so any type with the
members the generated scanner has will do. Every node carries its span as `byte_offset` and `byte_length`, which the
scanner's `text` turns into its bytes.

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
BenchmarkFromParser/GNU_Bison_3.8.2-32          608     1,716,551 ns/op       536,522 B/op     8,618 allocs/op
BenchmarkFromParser/GCC_2.95.3_C-32             180     6,291,328 ns/op     2,763,803 B/op    29,858 allocs/op
BenchmarkFromParser/GCC_2.95.3_Objective_C-32   124     9,097,340 ns/op     4,601,872 B/op    41,457 allocs/op
BenchmarkFromParser/GCC_3.3.6_C++-32             43    29,006,969 ns/op    11,603,946 B/op    81,681 allocs/op
BenchmarkFromParser/GCC_4.2.4_Java-32           100    10,326,418 ns/op     4,622,337 B/op    44,475 allocs/op
BenchmarkFromParser/Go_1.5.4-32                 252     5,698,106 ns/op     2,697,234 B/op    28,926 allocs/op
BenchmarkFromParser/PHP_8.6.7-32                 63    17,015,554 ns/op     8,478,942 B/op    60,665 allocs/op
BenchmarkFromParser/PostgreSQL_18.4-32            2   540,262,297 ns/op   138,290,256 B/op   490,540 allocs/op
BenchmarkFromParser/Ruby_3.2.11-32               69    19,928,512 ns/op     9,811,740 B/op    62,891 allocs/op
PASS
ok      github.com/backbone81/golr/internal/parsergen/backend/cpp       11.391s
```
