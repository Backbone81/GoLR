#!/bin/sh
#
# Everything C++ specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the compiler in
# the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. Both are headers, because golr writes one file per invocation and
# a self contained header is what C++ has to offer for that. scanner.hpp is also what the generated parser includes,
# which is the default of --backend-cpp-scanner-include, and the quoted include finds it as the sibling it is.
export SCANNER_FILE_NAME=scanner.hpp
export PARSER_FILE_NAME=parser.hpp

# One case, in one compile. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The generated code is header only, so compiling the runner compiles both generated files with it and there is nothing
# to build separately.
#
# C++17 is the standard the backend targets, and it is named explicitly rather than left to the compiler, whose default
# moves with its version. The warnings go well past what any project is scaffolded with: -Wall -Wextra -Wpedantic is
# the usual strict set, and -Wconversion with -Wsign-conversion additionally reports every implicit narrowing and every
# implicit change of signedness, which is the mistake a table driven driver makes when it indexes with a plain char.
# -Werror is what ends the case over any of them, so code which passes here compiles quietly rather than merely
# compiling.
execute() {
    set -e

    g++ -std=c++17 -Wall -Wextra -Wpedantic -Wconversion -Wsign-conversion -Werror -o runner runner.cpp

    exec ./runner input.txt
}
