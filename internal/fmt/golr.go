package fmt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// GoLR parses the GoLR grammar from the given reader and writes the formatted version to the given writer.
func GoLR(reader io.Reader, writer io.Writer, filePath string) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Format: GoLR").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	scanner := golrparser.NewTokenSkipper(
		golrparser.NewScanner(data, filePath),
	)

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(scanner)
	if err != nil {
		return err
	}

	walker := golr.NewTreeWalker()
	rules, grammar, err := walker.BuildGrammar(rootNode)
	if err != nil {
		return err
	}

	return golr.FromGrammar(writer, rules, grammar)
}

// GoLRFile parses the GoLR grammar from the given input file path and writes the formatted version to the given output
// file path. Input and output file path can be the same. A temporary file is used to ensure that any parsing errors
// do not lead to an empty input file.
func GoLRFile(inputFilePath string, outputFilePath string) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	input, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("opening GoLR file %q: %w", inputFilePath, err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing input file: %w", closeErr))
		}
	}()

	tempPath, err := formatToTempFile(input, inputFilePath, filepath.Dir(outputFilePath))
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			// On success, we do not remove the temporary file as it has been renamed.
			return
		}
		if removeErr := os.Remove(tempPath); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("removing temporary file: %w", removeErr))
		}
	}()

	if err = os.Rename(tempPath, outputFilePath); err != nil {
		return fmt.Errorf("replacing %q with formatted output: %w", outputFilePath, err)
	}
	return nil
}

// formatToTempFile formats the GoLR grammar from reader into a new temporary file in dir and returns its path.
// The caller is responsible for removing or renaming the file. On error the temporary file is cleaned up automatically.
//
//nolint:nonamedreturns // Named returns are required to deal with errors in defer statements.
func formatToTempFile(reader io.Reader, filePath string, dir string) (tempPath string, err error) {
	tempFile, err := os.CreateTemp(dir, ".golrfmt-*")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %w", err)
	}
	defer func() {
		if err == nil {
			// On success, we do not remove the temporary file.
			return
		}
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("removing temporary file: %w", removeErr))
		}
	}()
	defer func() {
		if closeErr := tempFile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing temporary file: %w", closeErr))
		}
	}()

	if err = GoLR(reader, tempFile, filePath); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

// GoLRString parses the GoLR grammar from the given string and returns the formatted version.
func GoLRString(input string) (string, error) {
	var builder strings.Builder
	if err := GoLR(strings.NewReader(input), &builder, "in-memory"); err != nil {
		return "", err
	}
	return builder.String(), nil
}
