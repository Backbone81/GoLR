package bison

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ErrVersionParseError = errors.New("could not parse bison version")

type BisonVersion struct {
	Major int
	Minor int
	Patch int
}

func Version() (BisonVersion, error) {
	stdout, stderr, err := execute("--version")
	if err != nil {
		return BisonVersion{}, errors.Join(err, errors.New(stderr))
	}

	lines := strings.Split(stdout, "\n")
	if len(lines) < 1 {
		return BisonVersion{}, ErrVersionParseError
	}

	firstLine := lines[0]
	words := strings.Split(firstLine, " ")
	if len(words) < 2 {
		return BisonVersion{}, ErrVersionParseError
	}
	if words[0] != "bison" {
		return BisonVersion{}, ErrVersionParseError
	}

	version := words[len(words)-1]
	versionParts := strings.Split(version, ".")
	if len(versionParts) < 2 {
		return BisonVersion{}, ErrVersionParseError
	}
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return BisonVersion{}, ErrVersionParseError
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return BisonVersion{}, ErrVersionParseError
	}
	var patch int
	if len(versionParts) == 3 {
		patch, err = strconv.Atoi(versionParts[2])
		if err != nil {
			return BisonVersion{}, ErrVersionParseError
		}
	}
	return BisonVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func BuildLALR1(grammarFilePath string, automatonFilePath string) error {
	return build(grammarFilePath, automatonFilePath, "lalr")
}

func BuildIELR1(grammarFilePath string, automatonFilePath string) error {
	return build(grammarFilePath, automatonFilePath, "ielr")
}

func BuildLR1(grammarFilePath string, automatonFilePath string) error {
	return build(grammarFilePath, automatonFilePath, "canonical-lr")
}

func build(grammarFilePath string, automatonFilePath string, parserType string) error {
	version, err := Version()
	if err != nil {
		return err
	}
	if version.Major != 3 {
		return errors.New("expected bison version 3")
	}

	args := []string{
		"--warnings=no-other",
		"-Werror=conflicts-rr",
		"--output=/dev/null",
		"--report-file=/dev/null",
		"--xml=" + automatonFilePath,
		"--define=lr.type=" + parserType,
	}
	if version.Minor > 7 {
		args = append(args, "--header=/dev/null")
	}
	args = append(args, grammarFilePath)
	if _, _, err := execute(args...); err != nil {
		return err
	}
	return nil
}

func execute(args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bison", args...) //nolint:gosec // The variables are fine and controlled by ourselves.

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("executing bison: %w", err)
		err = errors.Join(err, errors.New(stdout.String()))
		err = errors.Join(err, errors.New(stderr.String()))
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}
