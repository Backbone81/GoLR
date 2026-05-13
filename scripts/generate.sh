#!/usr/bin/env bash

set -ex

# This shell script is executing commands to generate code and do some sanity checks.

# Generate the internal parsergen GNU Bison frontend parser.
go run ./internal/parsergen/frontend/bison/spec/export/

# Copy files from the internal parsergen GNU Bison frontend to the examples folder.
cp internal/parsergen/frontend/bison/spec/*.go examples/bison/spec
cp internal/parsergen/frontend/bison/spec/*.y examples/bison/spec
cp internal/parsergen/frontend/bison/spec/*.txt examples/bison/spec
cp internal/parsergen/frontend/bison/spec/LICENSES examples/bison/spec
cp internal/parsergen/frontend/bison/parser/*.go examples/bison/parser

# Make sure that all generated code of the examples are actually updated-
go run ./examples/bison/spec/export/
go run ./examples/golang/spec/export/
go run ./examples/golang/parser/export/

# Let's make sure that our examples folder does not reference any internal package.
if grep -r '/internal/' examples/; then
  echo "ERROR: examples/ must not reference internal packages"
  exit 1
fi
