#!/bin/sh
#
# Everything Kotlin specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and
# the parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the compiler in
# the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. Kotlin ties neither a file name nor a directory to what a file
# declares, but the generated files declare a package and laying them out along it is what a reader of the emitted code
# expects.
export SCANNER_FILE_NAME=parser/Scanner.kt
export PARSER_FILE_NAME=parser/Parser.kt

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The release is pinned rather than left to the compiler of the image, so that the generated code is held to the version
# the backend targets instead of to whatever the container happens to ship: -jvm-target picks the bytecode and
# -Xjdk-release the class library the emitted code may reach into. "-Werror" then fails the case on a warning, which is
# stricter than any default a user builds with, so code which passes here compiles quietly for them.
#
# The classes go into a directory of their own, so that the class files of the generated package can never be mistaken
# for the sources next to them. Running goes through the kotlin launcher, because that is what puts the Kotlin standard
# library on the class path; a top level main function lands in the class named after its file.
execute() {
    set -e

    kotlinc -jvm-target 17 -Xjdk-release=17 -Werror -d classes parser/Scanner.kt parser/Parser.kt Runner.kt

    exec kotlin -classpath classes RunnerKt input.txt
}
