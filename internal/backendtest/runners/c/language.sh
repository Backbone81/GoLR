#!/bin/sh
#
# Everything C specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the compiler in
# the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. Both are headers which carry their implementation behind a macro,
# because golr writes one file per invocation and that is what C has to offer for it: a header may be included anywhere,
# while the tables and the function bodies are emitted in the one translation unit which asks for them. scanner.h is
# also what the generated parser includes, which is the default of --backend-c-scanner-include.
export SCANNER_FILE_NAME=scanner.h
export PARSER_FILE_NAME=parser.h

# One case, in one compile. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The runner defines both implementation macros, so compiling it compiles both generated files with it and there is
# nothing to build separately.
#
# C99 is the standard the backend targets, and it is named explicitly rather than left to the compiler, whose default
# moves with its version. The warnings go well past what any project is scaffolded with: -Wall -Wextra -Wpedantic is
# the usual strict set, and -Wconversion with -Wsign-conversion additionally reports every implicit narrowing and every
# implicit change of signedness, which is the mistake a table driven driver makes when it indexes with a plain char.
# -Werror is what ends the case over any of them, so code which passes here compiles quietly rather than merely
# compiling.
#
# The sanitizers are what no other language needs: C is the only backend which manages its own memory, so a leak, a use
# after free or an out of bounds read is a defect in the generated parser which no trace could show - a trace says what
# the parser computed and never what it did to the heap on the way. They belong on this one build rather than on a
# second one, because they change what the program reports about itself and never what it computes, so the traces are
# the same either way. -fno-sanitize-recover is what turns a finding into a failed case instead of a line nobody reads.
execute() {
    set -e

    gcc -std=c99 -g -fsanitize=address,undefined -fno-sanitize-recover=all \
        -Wall -Wextra -Wpedantic -Wconversion -Wsign-conversion -Werror -o runner runner.c

    exec ./runner input.txt
}
