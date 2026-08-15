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
# The type check runs twice, once against the tsconfig.json "tsc --init" writes and once against --strict alone. The two
# disagree about the generated code, so only checking one would let a regression through. The first also emits, which is
# what the runner is then run as: TypeScript is compiled rather than stripped, because only the compiler resolves the
# "./scanner.js" the parser imports.
execute() {
    set -e

    tsc --project tsconfig.json
    tsc --noEmit --ignoreConfig --strict --target esnext --module nodenext \
        "$SCANNER_FILE_NAME" "$PARSER_FILE_NAME" runner.ts node.d.ts

    exec node runner.js input.txt
}
