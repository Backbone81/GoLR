package backendtest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// The harness reaches every language toolchain through docker compose. Starting the containers is the job of the
// test-backends target of the Makefile, which knows which languages to run because LANGUAGE is its own focus; what is
// left here is finding the traces that run left behind.
const (
	// workPath is the root of what the containers see, relative to the repository root. It is below tmp/, which is
	// gitignored, so nothing a container produces can be committed by accident.
	workPath = "tmp/backendtest"

	// runnerPath holds one directory per language, with the hand written program which prints the trace and the
	// entrypoint script which builds and runs it. It is also what the harness reads to find out which languages it
	// can test, so a language exists exactly when its runner does.
	runnerPath = "internal/backendtest/runners"
)

// The two traces a run produces. They are separate files because the scanner trace and the parser trace are never
// interleaved into one stream.
const (
	ScannerRole = "scanner"
	ParserRole  = "parser"

	// TraceFileSuffix is appended to a role for the trace a runner wrote.
	TraceFileSuffix = ".actual"
)

// Languages returns every language the harness can test, which is every language with a runner directory. The list is
// read from disk rather than written out here, so that adding a language is adding a directory and never editing Go.
// The name of that directory is at once the compose service, the name of the backend the CLI generates with and the
// Ginkgo label which selects the language.
func Languages() ([]string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filepath.Join(root, runnerPath))
	if err != nil {
		return nil, fmt.Errorf("reading the runner directory: %w", err)
	}

	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name())
		}
	}
	return result, nil
}

// WorkPath returns the path on the host of a directory which a container wrote its traces to. It only names the path:
// creating the directories is the job of the Makefile and of the harness script inside the container, which is what
// keeps a run from being started by the act of looking for its output.
func WorkPath(pathElements ...string) (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{root, workPath}, pathElements...)...), nil
}

// repositoryRoot finds the repository by walking up from the working directory until it sees the module file. The
// harness needs the root to find the traces below tmp/, and a test runs in the directory of its own
// package, so the root cannot be assumed to be the working directory.
func repositoryRoot() (string, error) {
	result, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determining the working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(result, "go.mod")); err == nil {
			return result, nil
		}

		parent := filepath.Dir(result)
		if parent == result {
			return "", errors.New("no go.mod found in any parent directory, so the repository root is unknown")
		}
		result = parent
	}
}
