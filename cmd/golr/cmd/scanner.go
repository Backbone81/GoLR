package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/scannergen/backend"
	cppbackend "github.com/backbone81/golr/pkg/scannergen/backend/cpp"
	csharpbackend "github.com/backbone81/golr/pkg/scannergen/backend/csharp"
	dotbackend "github.com/backbone81/golr/pkg/scannergen/backend/dot"
	golangbackend "github.com/backbone81/golr/pkg/scannergen/backend/golang"
	golangtablebackend "github.com/backbone81/golr/pkg/scannergen/backend/golangtable"
	javabackend "github.com/backbone81/golr/pkg/scannergen/backend/java"
	javascriptbackend "github.com/backbone81/golr/pkg/scannergen/backend/javascript"
	jsonbackend "github.com/backbone81/golr/pkg/scannergen/backend/json"
	pythonbackend "github.com/backbone81/golr/pkg/scannergen/backend/python"
	rustbackend "github.com/backbone81/golr/pkg/scannergen/backend/rust"
	typescriptbackend "github.com/backbone81/golr/pkg/scannergen/backend/typescript"
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

	scannerBackendGoPackageName   string
	scannerBackendJavaPackageName string
	scannerBackendCSharpNamespace string
	scannerBackendCppNamespace    string
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
			rules, _, err := golrfrontend.ToRules(os.Stdin, "pipe")
			return rules, err
		}
		rules, _, err := golrfrontend.RulesFromFile(scannerFrontendFilePath)
		return rules, err
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
	case "cpp":
		if scannerBackendFilePath == "-" {
			return cppbackend.FromDFA(os.Stdout, dfa, cppbackend.Config{
				Namespace: scannerBackendCppNamespace,
			})
		}
		return cppbackend.DFAToFile(scannerBackendFilePath, dfa, cppbackend.Config{
			Namespace: scannerBackendCppNamespace,
		})
	case "csharp":
		if scannerBackendFilePath == "-" {
			return csharpbackend.FromDFA(os.Stdout, dfa, csharpbackend.Config{
				Namespace: scannerBackendCSharpNamespace,
			})
		}
		return csharpbackend.DFAToFile(scannerBackendFilePath, dfa, csharpbackend.Config{
			Namespace: scannerBackendCSharpNamespace,
		})
	case "dot":
		if scannerBackendFilePath == "-" {
			return dotbackend.FromDFA(os.Stdout, dfa)
		}
		return dotbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "go-direct":
		if scannerBackendFilePath == "-" {
			return golangbackend.FromDFA(os.Stdout, dfa, golangbackend.Config{
				PackageName: scannerBackendGoPackageName,
			})
		}
		return golangbackend.DFAToFile(scannerBackendFilePath, dfa, golangbackend.Config{
			PackageName: scannerBackendGoPackageName,
		})
	case "go", "go-table":
		if scannerBackendFilePath == "-" {
			return golangtablebackend.FromDFA(os.Stdout, dfa, golangtablebackend.Config{
				PackageName: scannerBackendGoPackageName,
			})
		}
		return golangtablebackend.DFAToFile(scannerBackendFilePath, dfa, golangtablebackend.Config{
			PackageName: scannerBackendGoPackageName,
		})
	case "java":
		if scannerBackendFilePath == "-" {
			return javabackend.FromDFA(os.Stdout, dfa, javabackend.Config{
				PackageName: scannerBackendJavaPackageName,
			})
		}
		return javabackend.DFAToFile(scannerBackendFilePath, dfa, javabackend.Config{
			PackageName: scannerBackendJavaPackageName,
		})
	case "javascript":
		if scannerBackendFilePath == "-" {
			return javascriptbackend.FromDFA(os.Stdout, dfa)
		}
		return javascriptbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "json":
		if scannerBackendFilePath == "-" {
			return jsonbackend.FromDFA(os.Stdout, dfa)
		}
		return jsonbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "null":
		// Nothing to do.
		return nil
	case "python":
		if scannerBackendFilePath == "-" {
			return pythonbackend.FromDFA(os.Stdout, dfa)
		}
		return pythonbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "rust":
		if scannerBackendFilePath == "-" {
			return rustbackend.FromDFA(os.Stdout, dfa)
		}
		return rustbackend.DFAToFile(scannerBackendFilePath, dfa)
	case "typescript":
		if scannerBackendFilePath == "-" {
			return typescriptbackend.FromDFA(os.Stdout, dfa)
		}
		return typescriptbackend.DFAToFile(scannerBackendFilePath, dfa)
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
		"golr",
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
		"The backend to use for writing the scanner. One of: cpp, csharp, dot, go, go-direct, go-table, java,"+
			" javascript, json, null, python, rust, typescript, yaml.",
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
	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendJavaPackageName,
		"backend-java-package-name",
		"parser",
		"The Java package name to use for the generated Java code.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendCSharpNamespace,
		"backend-csharp-namespace",
		"Parser",
		"The C# namespace to use for the generated C# code.",
	)
	scannerCmd.PersistentFlags().StringVar(
		&scannerBackendCppNamespace,
		"backend-cpp-namespace",
		"parser",
		"The C++ namespace to use for the generated C++ code.",
	)
}
