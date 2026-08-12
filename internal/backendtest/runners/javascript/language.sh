#!/bin/sh
#
# Everything JavaScript specific about running one case of the corpus. The shared entrypoint.sh generates the scanner
# and the parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all, which is what
# proves that generated GoLR code needs nothing but the bare language.

# The names the generator writes the two files under. The .js extension and the package.json next to it, which says
# "type": "module", are what make node load them as ECMAScript modules. scanner.js is also the module specifier the
# generated parser imports the token constants from, which is the default of --backend-javascript-scanner-module.
export SCANNER_FILE_NAME=scanner.js
export PARSER_FILE_NAME=parser.js

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e" and lets the runner be
# exec'd. JavaScript needs no compilation step, which is why this is the whole of it.
execute() {
    set -e
    exec node runner.js input.txt
}
