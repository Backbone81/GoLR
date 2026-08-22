#!/usr/bin/env bash

set -ex

# This shell script is executing commands to generate code and do some sanity checks.

# Generate the internal parsergen GNU Bison frontend parser.
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path internal/parsergen/frontend/bison/spec/bison.golr \
  --backend go \
  --backend-file-path internal/parsergen/frontend/bison/parser/scanner.go
go run ./cmd/golr parser \
  --frontend bison \
  --frontend-file-path internal/parsergen/frontend/bison/spec/bison-3.8.2.y \
  --backend go \
  --backend-file-path internal/parsergen/frontend/bison/parser/parser.go

# Generate the internal parsergen GoLR frontend parser.
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path internal/parsergen/frontend/golr/spec/golr.golr \
  --backend go \
  --backend-file-path internal/parsergen/frontend/golr/parser/scanner.go
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path internal/parsergen/frontend/golr/spec/golr.golr \
  --backend go \
  --backend-file-path internal/parsergen/frontend/golr/parser/parser.go

# Copy files from the internal parsergen GNU Bison frontend to the examples folder.
cp internal/parsergen/frontend/bison/spec/*.go examples/bison/spec
cp internal/parsergen/frontend/bison/spec/*.golr examples/bison/spec
cp internal/parsergen/frontend/bison/spec/*.y examples/bison/spec
cp internal/parsergen/frontend/bison/spec/*.txt examples/bison/spec
cp internal/parsergen/frontend/bison/spec/LICENSES examples/bison/spec
cp internal/parsergen/frontend/bison/parser/*.go examples/bison/parser

# Generate the example calculator parser.
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend go \
  --backend-file-path examples/calculator/golang/parser/scanner.go
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend go \
  --backend-file-path examples/calculator/golang/parser/parser.go
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend java \
  --backend-file-path examples/calculator/java/parser/Scanner.java
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend java \
  --backend-file-path examples/calculator/java/parser/Parser.java
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend rust \
  --backend-file-path examples/calculator/rust/src/parser/scanner.rs
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend rust \
  --backend-file-path examples/calculator/rust/src/parser/parser.rs
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend python \
  --backend-file-path examples/calculator/python/parser/scanner.py
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend python \
  --backend-python-scanner-module .scanner \
  --backend-file-path examples/calculator/python/parser/parser.py
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend javascript \
  --backend-file-path examples/calculator/javascript/parser/scanner.js
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend javascript \
  --backend-file-path examples/calculator/javascript/parser/parser.js
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend typescript \
  --backend-file-path examples/calculator/typescript/parser/scanner.ts
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend typescript \
  --backend-file-path examples/calculator/typescript/parser/parser.ts
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend cpp \
  --backend-cpp-namespace calculator::parser \
  --backend-file-path examples/calculator/cpp/parser/scanner.hpp
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend cpp \
  --backend-cpp-namespace calculator::parser \
  --backend-file-path examples/calculator/cpp/parser/parser.hpp
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend csharp \
  --backend-csharp-namespace Calculator.Parser \
  --backend-file-path examples/calculator/csharp/Parser/Scanner.cs
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/calculator/calculator.golr \
  --backend csharp \
  --backend-csharp-namespace Calculator.Parser \
  --backend-file-path examples/calculator/csharp/Parser/Parser.cs

# Generate the example Go parser.
go run ./cmd/golr scanner \
  --frontend golr \
  --frontend-file-path examples/golang/spec/golang.golr \
  --backend go \
  --backend-file-path examples/golang/parser/scanner.go
go run ./cmd/golr parser \
  --frontend golr \
  --frontend-file-path examples/golang/spec/golang.golr \
  --backend go \
  --backend-file-path examples/golang/parser/parser.go

# Let's make sure that our examples folder does not reference any internal package.
if grep -r 'github.com/backbone81/golr/internal/' examples/; then
  echo "ERROR: examples/ must not reference internal packages"
  exit 1
fi

# Let's make sure that our cmd folder does not reference any internal package.
if grep -r 'github.com/backbone81/golr/internal/' cmd/; then
  echo "ERROR: cmd/ must not reference internal packages"
  exit 1
fi
