package cmd

import (
	"fmt"
	"golr/pkg/scannergen/backend"
	golangbackend "golr/pkg/scannergen/backend/golang"
	jsonbackend "golr/pkg/scannergen/backend/json"
	yamlbackend "golr/pkg/scannergen/backend/yaml"
	subsetcore "golr/pkg/scannergen/core/subset"
	"golr/pkg/scannergen/frontend"
	jsonfrontend "golr/pkg/scannergen/frontend/json"
	yamlfrontend "golr/pkg/scannergen/frontend/yaml"

	"github.com/spf13/cobra"
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
	case "json":
		return jsonfrontend.RulesFromFile(scannerFrontendFilePath)
	case "yaml":
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
		return golangbackend.DFAToFile(scannerBackendFilePath, dfa, golangbackend.Config{
			PackageName: scannerBackendGoPackageName,
		})
	case "json":
		return jsonbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "yaml":
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
		"The frontend to use for reading the regular expressions. One of: json, yaml.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerFrontendFilePath,
		"frontend-file-path",
		"",
		"The file path to read the regular expressions from.",
	)

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
		"The backend to use for writing the scanner. One of: go, json, yaml.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendFilePath,
		"backend-file-path",
		"",
		"The file path to write the scanner to.",
	)

	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendGoPackageName,
		"backend-go-package-name",
		"parser",
		"The Go package name to use for the generated Go code.",
	)
}
