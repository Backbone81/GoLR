package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/scannergen/backend"
	golangbackend "github.com/backbone81/golr/pkg/scannergen/backend/golang"
	jsonbackend "github.com/backbone81/golr/pkg/scannergen/backend/json"
	yamlbackend "github.com/backbone81/golr/pkg/scannergen/backend/yaml"
	subsetcore "github.com/backbone81/golr/pkg/scannergen/core/subset"
	"github.com/backbone81/golr/pkg/scannergen/frontend"
	golrfrontend "github.com/backbone81/golr/pkg/scannergen/frontend/golr"
	jsonfrontend "github.com/backbone81/golr/pkg/scannergen/frontend/json"
	yamlfrontend "github.com/backbone81/golr/pkg/scannergen/frontend/yaml"
)

var (
	scannerFrontend         string
	scannerFrontendFilePath string

	scannerCore string

	scannerBackend         string
	scannerBackendFilePath string

	scannerBackendGoPackageName string
)

var scannerCmd = &cobra.Command{
	Use:          "scanner",
	Short:        "Generates a DFA scanner.",
	Long:         `Generates a DFA scanner.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := executeScannerFrontend()
		if err != nil {
			return err
		}

		dfa, err := executeScannerCore(rules)
		if err != nil {
			return err
		}

		if err := executeScannerBackend(dfa); err != nil {
			return err
		}
		return nil
	},
}

func executeScannerFrontend() ([]frontend.Rule, error) {
	switch scannerFrontend {
	case "golr":
		if scannerFrontendFilePath == "-" {
			return golrfrontend.ToRules(os.Stdin, "pipe")
		}
		return golrfrontend.RulesFromFile(scannerFrontendFilePath)
	case "json":
		if scannerFrontendFilePath == "-" {
			return jsonfrontend.ToRules(os.Stdin)
		}
		return jsonfrontend.RulesFromFile(scannerFrontendFilePath)
	case "yaml":
		if scannerFrontendFilePath == "-" {
			return yamlfrontend.ToRules(os.Stdin)
		}
		return yamlfrontend.RulesFromFile(scannerFrontendFilePath)
	default:
		return nil, fmt.Errorf("unsupported scanner frontend %q", scannerFrontend)
	}
}

func executeScannerCore(rules []frontend.Rule) (backend.DFA, error) {
	switch scannerCore {
	case "subset":
		return subsetcore.RulesToDFA(rules), nil
	default:
		return backend.DFA{}, fmt.Errorf("unsupported scanner core %q", scannerCore)
	}
}

func executeScannerBackend(dfa backend.DFA) error {
	switch scannerBackend {
	case "go":
		if scannerBackendFilePath == "-" {
			return golangbackend.FromDFA(os.Stdout, dfa, golangbackend.Config{
				PackageName: scannerBackendGoPackageName,
			})
		}
		return golangbackend.DFAToFile(scannerBackendFilePath, dfa, golangbackend.Config{
			PackageName: scannerBackendGoPackageName,
		})
	case "json":
		if scannerBackendFilePath == "-" {
			return jsonbackend.FromDFA(os.Stdout, dfa)
		}
		return jsonbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "null":
		// Nothing to do.
		return nil
	case "yaml":
		if scannerBackendFilePath == "-" {
			return yamlbackend.FromDFA(os.Stdout, dfa)
		}
		return yamlbackend.DFAToFile(scannerBackendFilePath, dfa)
	default:
		return fmt.Errorf("unsupported scanner backend %q", scannerBackend)
	}
}

func init() {
	rootCmd.AddCommand(scannerCmd)

	scannerCmd.PersistentFlags().StringVar(
		&scannerFrontend,
		"frontend",
		"yaml",
		"The frontend to use for reading the regular expressions. One of: golr, json, yaml.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerFrontendFilePath,
		"frontend-file-path",
		"",
		"The file path to read the regular expressions from. Can be '-' to read from stdin.",
	)
	if err := scannerCmd.MarkPersistentFlagRequired("frontend-file-path"); err != nil {
		panic(err)
	}

	scannerCmd.PersistentFlags().StringVar(
		&scannerCore,
		"core",
		"subset",
		"The core to use for generating the scanner from the regular expressions. One of: subset.",
	)

	scannerCmd.PersistentFlags().StringVar(
		&scannerBackend,
		"backend",
		"go",
		"The backend to use for writing the scanner. One of: go, json, null, yaml.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendFilePath,
		"backend-file-path",
		"",
		"The file path to write the scanner to. Can be '-' to write to stdout.",
	)
	if err := scannerCmd.MarkPersistentFlagRequired("backend-file-path"); err != nil {
		panic(err)
	}

	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendGoPackageName,
		"backend-go-package-name",
		"parser",
		"The Go package name to use for the generated Go code.",
	)
}
