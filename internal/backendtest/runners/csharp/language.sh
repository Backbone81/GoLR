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
# The build runs twice, once against the configuration "dotnet new console" writes and once against an older project:
# no nullable context, no implicit usings and an earlier language version. The two disagree about the generated code -
# a file which states its own nullable context compiles under both, one which relies on the project setting does not -
# so only checking one would let a regression through. The second build gets output directories of its own, because
# MSBuild would otherwise consider the first one up to date and skip it.
execute() {
    set -e

    dotnet build runner.csproj --verbosity quiet --nologo \
        -p:Nullable=disable \
        -p:ImplicitUsings=disable \
        -p:LangVersion=10 \
        -p:BaseOutputPath=baseline-bin/ \
        -p:BaseIntermediateOutputPath=baseline-obj/

    dotnet build runner.csproj --verbosity quiet --nologo

    exec dotnet bin/Debug/net10.0/runner.dll input.txt
}
