#!/bin/sh
#
# Everything Go specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the toolchain in
# the image is all there is, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. They go into a package of their own because the generated code
# declares "package parser" while the runner beside it is "package main", and two packages cannot share a directory.
# That is also what a Go project which keeps its generated parser apart looks like, so the layout is the one a user
# has. The go.mod copied in beside the runner names the module runner, which makes the package runner/parser.
export SCANNER_FILE_NAME=parser/scanner.go
export PARSER_FILE_NAME=parser/parser.go

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# "go run" compiles and runs in one step and leaves no binary behind in the case directory. The compilation costs
# nothing after the first case: the build cache is shared by the whole corpus through the GOCACHE the compose service
# sets, so the standard library is compiled once and not twenty four times.
execute() {
    set -e

    exec go run . input.txt
}
