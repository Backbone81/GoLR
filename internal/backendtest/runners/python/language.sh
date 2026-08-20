#!/bin/sh
#
# Everything Python specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and
# the parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all and the interpreter
# in the image is the whole toolchain, which is what proves that generated GoLR code needs nothing but the bare
# language.

# The names the generator writes the two files under. scanner.py is also the module the generated parser imports the
# token constants from, which is the default of --backend-python-scanner-module. The two sit next to runner.py with no
# package around them, so an import finds them on the path of the script.
export SCANNER_FILE_NAME=scanner.py
export PARSER_FILE_NAME=parser.py

# One case, in one run of the interpreter. The entrypoint calls this in a subshell of its own, which contains the
# "set -e".
#
# There is no compile step before it, because Python has none to offer: importing the generated files is what compiles
# them. The interpreter is instead given the tightest settings it has, so that anything short of a clean run ends the
# case and gets looked at:
#
#   -X dev    turns on the runtime checks which are off by default, among them ResourceWarning
#   -bb       makes comparing bytes to str an error, which is the mistake a runner would make by handing the generated
#             scanner text where it wants the bytes of the input
#   -W error  fails on any warning at all, whether it comes from compiling the generated files on import or from
#             running them. An unrecognized escape sequence in a generated string literal is a SyntaxWarning, and this
#             is what ends the case over it instead of printing it and carrying on.
execute() {
    set -e

    exec python -X dev -bb -W error runner.py input.txt
}
