package backendtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The harness reaches every language toolchain through docker compose. This file knows the language names and the
// invocation; it deliberately knows no image tag, because docker-compose.yaml is the single place where those live.
const (
	// composeFilePath is the compose file, relative to the repository root.
	composeFilePath = "internal/backendtest/docker-compose.yaml"

	// workPath is the one directory a container can see, relative to the repository root. It is below tmp/, which is
	// gitignored, so nothing a container produces can be committed by accident.
	workPath = "tmp/backendtest"

	// ContainerWorkPath is where workPath is mounted inside every container, and the working directory of every
	// command the harness runs. A command therefore names its files relative to this and never needs a host path.
	ContainerWorkPath = "/work"

	// enabledEnvVar makes a run start containers. The container tests are not part of "make test": they need docker
	// and take far longer than the unit tests, so they are opt in and "make test-backends" is what opts in.
	enabledEnvVar = "BACKEND_TESTS"
)

// Languages is every language the harness has a toolchain for, and the name of that language's service in
// docker-compose.yaml. C and C++ share an image but not a service, so a language is always addressed by its own name.
var Languages = []string{
	"go",
	"java",
	"rust",
	"javascript",
	"typescript",
	"python",
	"c",
	"cpp",
	"csharp",
}

// Enabled reports whether this run may start containers. It is the single reader of the environment variable, so the
// tests which skip and the message which explains the skip can never disagree about what turns them on.
func Enabled() bool {
	return os.Getenv(enabledEnvVar) == "1"
}

// EnabledHint is what to tell a reader who wants the skipped container tests to run.
func EnabledHint() string {
	return fmt.Sprintf("run 'make test-backends', or set %s=1, to run the language backends in containers", enabledEnvVar)
}

// WorkPath returns the path on the host of a directory the containers see below ContainerWorkPath. The directory is
// created, because docker creates a missing bind mount source itself and does so as root, which the container then
// cannot write to as the invoking user.
func WorkPath(elem ...string) (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", err
	}

	result := filepath.Join(append([]string{root, workPath}, elem...)...)
	//nolint:gosec // World readable is deliberate: this holds generated test code below the gitignored tmp/.
	if err := os.MkdirAll(result, 0o755); err != nil {
		return "", fmt.Errorf("creating the working directory of the containers: %w", err)
	}
	return result, nil
}

// Run executes command in the container of the given language and returns what it wrote to standard output. The
// command runs in ContainerWorkPath, so it names its files relative to that. Standard error is kept out of the result
// and reported with a failure instead, because standard output is what the trace comparison reads and a diagnostic
// mixed into it would corrupt the comparison.
//
// The deadline comes from the caller rather than from a constant here, because the first run of a language pulls its
// image and takes minutes while every run after that takes seconds, and no single constant serves both.
func Run(ctx context.Context, language string, command ...string) (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", err
	}
	if _, err := WorkPath(); err != nil {
		return "", err
	}

	//nolint:prealloc // A hand written capacity would have to be kept in sync with the number of flags below.
	args := []string{
		"compose",
		"--file", filepath.Join(root, composeFilePath),

		// A relative bind mount resolves against the project directory, which defaults to the directory holding the
		// compose file. The mounts are written relative to the repository root, so the root has to be named here.
		"--project-directory", root,

		"run",

		// The container exists for one command, and the harness starts one per case.
		"--rm",

		// A service has no dependencies, and asking compose not to look for any keeps a future one from being started
		// silently.
		"--no-deps",

		// Without this, compose allocates a pseudo terminal, which rewrites every LF on the way out to CRLF and would
		// corrupt every trace this harness compares.
		"--no-TTY",

		language,
	}
	args = append(args, command...)

	//nolint:gosec // The arguments are built right here and controlled by ourselves.
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = root

	// The two variables the compose file interpolates are set here rather than expected from the environment. Compose
	// turns an unset variable into the empty string, so a forgotten export would leave "user: ':'" and a failure which
	// says nothing about the cause.
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("GOLR_UID=%d", os.Getuid()),
		fmt.Sprintf("GOLR_GID=%d", os.Getgid()),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf(
			"running '%s' in the %s container: %w\n%s",
			strings.Join(command, " "),
			language,
			err,
			stderr.String(),
		)
	}
	return stdout.String(), nil
}

// repositoryRoot finds the repository by walking up from the working directory until it sees the module file. The
// harness needs the root for the compose file and for the bind mounts, and a test runs in the directory of its own
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
