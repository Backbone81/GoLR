#!/bin/sh
#
# Builds and runs one case of the corpus in JavaScript, and prints its trace. The shared entrypoint.sh has already put
# the generated source, the input and this runner in the working directory and asks for one role at a time, so
# everything below is JavaScript specific and nothing else is.
#
# This file is sourced by entrypoint.sh, so it holds function definitions and nothing else. A command outside a function
# would run in the shell of the entrypoint, which walks the whole corpus and must not take on the shell options of one
# language. JavaScript defines no setup either: node runs the emitted file as it is, and the package.json next to it is
# copied per case like the rest of the runner.
#
# Nothing here may reach the network or install anything. The container runs with no network at all, which is what
# proves that generated GoLR code needs nothing but the bare language.

# One case and one role. The entrypoint calls this in a subshell of its own, which is what contains the "set -e" and
# lets the runner be exec'd.
execute() {
    set -e

    # The generated source arrives under a name which says nothing about the language, because the generating side is
    # the same for all of them. Giving it the extension node needs is this function's job. Together with the
    # package.json next to it, which says "type": "module", the .js file is loaded as an ECMAScript module.
    cp scanner.generated scanner.js

    # The trace and nothing else goes to standard output, so the runner replaces the subshell rather than running in
    # one whose output could be added to. JavaScript needs no compilation step, which is why this is the whole of it.
    exec node "${1}_runner.js" input.txt
}
