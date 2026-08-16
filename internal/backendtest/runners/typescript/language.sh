#!/bin/sh
#
# Everything TypeScript specific about running one case of the corpus. The shared entrypoint.sh generates the scanner
# and the parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all, which is what
# proves that generated GoLR code needs nothing but the bare language. The compiler is in the image.

# The names the generator writes the two files under. The package.json next to them says "type": "module", which is what
# makes the emitted JavaScript load as ECMAScript modules. The generated parser imports the token constants from
# "./scanner.js", which is what TypeScript spells the scanner.ts here as.
export SCANNER_FILE_NAME=scanner.ts
export PARSER_FILE_NAME=parser.ts

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The type check runs against the tsconfig.json "tsc --init" writes, which holds the generated code to the configuration
# a normal project is scaffolded with. That one is the strictest of the checks available: it turns on more than --strict
# does on its own, noUncheckedIndexedAccess and exactOptionalPropertyTypes among it. The same run emits, which is what
# the runner is then run as: TypeScript is compiled rather than stripped, because only the compiler resolves the
# "./scanner.js" the parser imports.
execute() {
    set -e

    tsc --project tsconfig.json

    exec node runner.js input.txt
}
