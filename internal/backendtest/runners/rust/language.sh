#!/bin/sh
#
# Everything Rust specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the compiler in
# the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. They are the module names runner.rs declares, and the generated
# parser reaches the generated scanner through the module path those give it.
export SCANNER_FILE_NAME=scanner.rs
export PARSER_FILE_NAME=parser.rs

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# An edition changes the language itself, and the generated files are dropped into a crate whose edition GoLR does not
# choose. The one compiled here is the edition "cargo new" writes into a fresh manifest, so it is the edition the
# generated code meets in a user's crate unless that user goes out of their way to pick another one. "-D warnings"
# fails the case on a warning, so code which passes here compiles quietly rather than merely compiling.
#
# rustc is given the crate root and finds the two generated modules beside it, so nothing has to be renamed and there is
# no manifest.
execute() {
    set -e

    rustc --edition 2024 -D warnings -o runner runner.rs

    exec ./runner input.txt
}
