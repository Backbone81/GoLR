#!/bin/sh
#
# Everything Java specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the compiler in
# the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. Java takes no choice here: a public type has to live in a file
# named after it, so these follow from the types the two backends emit.
export SCANNER_FILE_NAME=Scanner.java
export PARSER_FILE_NAME=Parser.java

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The release is pinned rather than left to the compiler of the image, so that the generated code is held to the version
# the backend targets instead of to whatever the container happens to ship. "-Xlint:all -Werror" then fails the case on
# a warning, which is stricter than any default a user builds with, so code which passes here compiles quietly for them.
#
# The generated files declare a package and the runner does not, so the class files land in ./parser and ./Runner.class.
# Naming the sources explicitly is what lets them sit flat in the case directory all the same.
execute() {
    set -e

    javac --release 17 -Xlint:all -Werror -d . Scanner.java Parser.java Runner.java

    exec java -cp . Runner input.txt
}
