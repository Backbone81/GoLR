#!/bin/sh
#
# Everything C# specific about running one case of the corpus. The shared entrypoint.sh generates the scanner and the
# parser, puts them and the input in the working directory, and calls execute there.
#
# This file is sourced by entrypoint.sh, so it holds the two file names and the function definitions and nothing else. A
# command outside a function would run in the shell of the entrypoint, which walks the whole corpus and must not take on
# the shell options of one language.
#
# Nothing here may reach the network or install anything. The container runs with no network at all, and the build
# restores from what the SDK image already carries, which is what proves that generated GoLR code needs nothing but the
# bare language.

# The names the generator writes the two files under. The project next to them compiles whatever C# it finds, so the
# names only have to be the ones a C# developer expects for the types in them.
export SCANNER_FILE_NAME=Scanner.cs
export PARSER_FILE_NAME=Parser.cs

# One case. The entrypoint calls this in a subshell of its own, which contains the "set -e".
#
# The build goes through "dotnet msbuild" rather than "dotnet build", which restores and builds just the same, so that
# a case which passes prints nothing at all. "dotnet build" appends a console logger parameter of its own which asks
# for the summary, and it wins over the one asked for here, so the "Build succeeded" block survives even the quiet
# verbosity there and scrolls past for every case of the corpus. Nothing is lost with the summary: errors and warnings
# are reported where they happen and not by it, which the quiet verbosity keeps as well.
execute() {
    set -e

    # The SDK verifies its workload manifests on the first build in a container and reports that it could not, because
    # the image carries none and the container has no network to fetch them over. Nothing built here uses a workload,
    # so the check is turned off instead of left to announce a problem which is none.
    export DOTNET_SKIP_WORKLOAD_INTEGRITY_CHECK=true

    dotnet msbuild runner.csproj -restore -target:Build -nologo -verbosity:quiet -consoleLoggerParameters:NoSummary

    exec dotnet bin/Debug/net10.0/runner.dll input.txt
}
