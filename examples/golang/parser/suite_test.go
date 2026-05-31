package parser_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRegex(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Examples: Golang: Parser Suite")
}

func stdLibSourceDir(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
	if err != nil {
		return "", err
	}

	return filepath.Join(strings.TrimSpace(string(out)), "src"), nil
}

func stdLibSourceFilePaths(ctx context.Context) ([]string, error) {
	sourceDir, err := stdLibSourceDir(ctx)
	if err != nil {
		return nil, err
	}

	var files []string
	if err := filepath.WalkDir(sourceDir, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if dirEntry.IsDir() || filepath.Ext(dirEntry.Name()) != ".go" {
			return nil
		}

		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}

func stdLibBenchmarkSourceFilePath(ctx context.Context) ([]byte, error) {
	sourceDir, err := stdLibSourceDir(ctx)
	if err != nil {
		return nil, err
	}

	// The net/http/server.go is a nice big source code file with roughly 130KB. This is perfect for benchmarks.
	filePath := filepath.Join(sourceDir, "net", "http", "server.go")
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return source, nil
}
