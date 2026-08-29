// The runner is a module of its own so that the Go tooling of the repository never looks at it. It imports the
// generated parser package, which does not exist until a case has been generated, so a package in the main module
// could not be built, vetted or linted. A directory with a go.mod in it is excluded from the patterns of the module
// around it, which is what keeps "go build ./..." and "go test ./internal/..." working.
//
// The module is named runner because that is what the generated package is imported through: the entrypoint writes the
// scanner and the parser to parser/, which makes them runner/parser.
module runner

// The version the repository itself is on. It is the language version the generated code is compiled under here, and
// it has to be one the image below its own toolchain can build.
go 1.25

toolchain go1.26.7
