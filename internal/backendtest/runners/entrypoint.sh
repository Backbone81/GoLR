#!/bin/sh
#
# The whole corpus for one language, in one container. This script is the entrypoint of every language and is the only
# place which knows how a case becomes a trace: it walks the corpus, generates the code, hands each case to the
# language.sh of the language and collects what that printed.
#
# It exists because the alternative is one container per case and per role - around 430 of them once every language has
# a runner - and because a loop copied into nine language scripts would be nine places to fix when generation changes.
# A language therefore provides nothing but a language.sh, which sets two variables and defines one or two functions:
#
#   SCANNER_FILE_NAME   the names the generated scanner and the generated parser are written under. The generator is
#   PARSER_FILE_NAME    pointed at them directly, so no language has to rename anything.
#
#   setup               runs once, before any case, in /work/<language>, above the case directories. It is optional:
#                       the default below does nothing, so a language which needs none defines only execute.
#   execute             runs once per case, in /work/<language>/<case>, with the generated sources, the input and the
#                       rest of the runner already there. It writes the two traces and nothing else.
#
# language.sh is sourced and not executed, which is what lets the work a language does once be done once: a compiled
# language builds its hand written runner in setup and compiles only the generated files per case, instead of paying for
# both twenty four times, and a toolchain which has to bring something up does it in setup and not per case. Being
# sourced is also why the file may hold no commands besides those two assignments. Others would run in the shell of this
# script, and a "set -e" among them would change how a failing case is treated here - see below. A function which wants
# that option sets it inside itself, where the subshell each call runs in contains it.
#
# The contract with the harness on the outside, all of it mounted from the repository and nothing staged:
#
#   /cases              the committed corpus, read only, one directory per case
#   /runners            this script and the runner of every language, read only
#   /work               the output, one directory per language, created here
#   golr                on the PATH, statically linked so it runs in any of the language images
#
# The language is the only argument, because a container has no other way of learning which service it is. It names both
# the backend to generate with and the runner to use.
#
# Per case the runner writes <case>/scanner.actual and <case>/parser.actual, and nothing else is captured: whatever the
# generator and the runner say goes to the output of this container, where the person who started the run can see it.
# Each case announces itself first, so that output belongs to a case and a long run says how far it has come.
#
# Two kinds of failure are deliberately told apart here, because they need opposite treatment:
#
#   - The run broke down - no corpus, no runner, no generator, a setup which failed. Nothing was produced, so there is
#     nothing to compare and the whole run has to fail. Those are the checks right below, and they exit non zero.
#   - One case failed to generate or crashed its runner. That is a defect in the backend under test, which is what this
#     harness exists to find, so it must be reported against that one case. Failing the run would hide the other twenty
#     three behind it and lose the name of the case which broke. Those failures are recorded and the sweep continues.
#
# That is also why there is no "set -e": it cannot tell the two apart.

set -u

language="$1"
languagePath="/runners/$language"

# The preconditions of the run. Each of these means the container was set up wrong rather than a backend being wrong,
# and every one of them would otherwise turn into twenty four identical and misleading case failures.
if [ ! -d /cases ]; then
    echo "the corpus is not mounted at /cases" >&2
    exit 1
fi

if ! command -v golr > /dev/null 2>&1; then
    echo "the golr generator is not on the PATH; it is mounted from the host and has to be statically linked" >&2
    exit 1
fi

if [ ! -f "$languagePath/language.sh" ]; then
    echo "there is no runner for $language at $languagePath" >&2
    exit 1
fi

# The setup a language does not define. Declaring it before language.sh is sourced is what makes the function optional
# without this script having to ask whether it exists.
setup() {
    :
}

# shellcheck source=/dev/null
. "$languagePath/language.sh"

if ! command -v execute > /dev/null 2>&1; then
    echo "$languagePath/language.sh defines no execute function, which is what turns one case into its traces" >&2
    exit 1
fi

if [ -z "${SCANNER_FILE_NAME:-}" ] || [ -z "${PARSER_FILE_NAME:-}" ]; then
    echo "$languagePath/language.sh sets no SCANNER_FILE_NAME and PARSER_FILE_NAME, which the generator writes to" >&2
    exit 1
fi

# Each language works below its own name, because the whole output tree is mounted rather than one directory per
# language. The Go side empties this before the run, so whatever is in here afterwards is what this run produced. It is
# also where the setup leaves whatever the cases go on to use, one directory above them.
mkdir -p "/work/$language" && cd "/work/$language" || exit 1

# The one time preparation of the language, before a single case exists. Whatever it has to say goes to standard error,
# so that it can never be mistaken for a trace, and it is left on the terminal rather than put in a log, because the
# only thing which reads it is the person whose "make test-backends" just stopped.
#
# A failing setup fails the whole run, unlike a failing case. It runs before anything has been generated, so it cannot
# be a defect in a backend - it is always the container or the toolchain being wrong, and every case after it would
# fail identically and bury the reason.
if ! setup >&2; then
    echo "the one time setup of $language failed, so no case can run" >&2
    exit 1
fi

for casePath in /cases/*/; do
    caseName=$(basename "$casePath")
    echo "--> $language/$caseName" >&2

    mkdir -p "$caseName"
    cp "$casePath/input.txt" "$caseName/"

    # The runner is copied into the case directory rather than run from where it is mounted, so that a case directory
    # holds everything that case needs, a runner can name its files without any path at all, and a language which has to
    # compile can write its output next to its source without touching a read only mount.
    #
    # Everything except language.sh, which is sourced once from where it is mounted. A copy of it in a case directory
    # would be a file which nothing runs.
    for runnerFile in "$languagePath"/*; do
        case "$runnerFile" in
            */language.sh) continue ;;
        esac
        cp "$runnerFile" "$caseName/"
    done

    # The core is pinned rather than left to the default, for the same reason scripts/generate.sh pins one: a change of
    # the default must not silently change what every backend is held to. It is the native Go core and never the bison
    # backed one, which shells out to a GNU Bison no language image carries, and which numbers symbols differently.
    #
    # The file names come from language.sh, so this script names no extension and no language ever renames anything.
    golr scanner \
        --frontend golr \
        --frontend-file-path "$casePath/spec.golr" \
        --core subset \
        --backend "$language" \
        --backend-file-path "$caseName/$SCANNER_FILE_NAME" \
        || continue

    golr parser \
        --frontend golr \
        --frontend-file-path "$casePath/spec.golr" \
        --core ielr1-golr \
        --backend "$language" \
        --backend-file-path "$caseName/$PARSER_FILE_NAME" \
        || continue

    # The subshell is what keeps a case from reaching the next one: it contains the working directory, and it contains
    # the shell options of a function which sets some.
    #
    # A runner which fails is not an error of this script. It leaves a missing or short trace behind, and reporting that
    # is the job of the comparison outside, which names the case and shows the diff.
    (cd "$caseName" && execute) || true
done

# The sweep completed, which is all this script promises. Whether the traces it produced are the right ones is decided
# outside, one case at a time. Without this the status would be that of the last case, so a single failing case at the
# end of the corpus would report as a broken run.
exit 0
