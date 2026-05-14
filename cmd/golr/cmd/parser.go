package cmd

import (
	"fmt"
	"golr/pkg/parsergen/backend"
	golangbackend "golr/pkg/parsergen/backend/golang"
	jsonbackend "golr/pkg/parsergen/backend/json"
	yamlbackend "golr/pkg/parsergen/backend/yaml"
	ielr1core "golr/pkg/parsergen/core/ielr1"
	"golr/pkg/parsergen/frontend"
	bisonfrontend "golr/pkg/parsergen/frontend/bison"
	jsonfrontend "golr/pkg/parsergen/frontend/json"
	yamlfrontend "golr/pkg/parsergen/frontend/yaml"

	"github.com/spf13/cobra"
)

var (
	parserFrontend         string
	parserFrontendFilePath string

	parserCore string

	parserBackend         string
	parserBackendFilePath string

	parserBackendGoPackageName string
)

var parserCmd = &cobra.Command{
	Use:          "parser",
	Short:        "Generates a LR(1) parser.",
	Long:         `Generates a LR(1) parser.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		grammar, err := executeParserFrontend()
		if err != nil {
			return err
		}

		parser, err := executeParserCore(grammar)
		if err != nil {
			return err
		}

		if err := executeParserBackend(parser); err != nil {
			return err
		}
		return nil
	},
}

func executeParserFrontend() (frontend.Grammar, error) {
	switch parserFrontend {
	case "bison":
		return bisonfrontend.GrammarFromFile(parserFrontendFilePath)
	case "json":
		return jsonfrontend.GrammarFromFile(parserFrontendFilePath)
	case "yaml":
		return yamlfrontend.GrammarFromFile(parserFrontendFilePath)
	default:
		return frontend.Grammar{}, fmt.Errorf("unsupported parser frontend %q", parserFrontend)
	}
}

func executeParserCore(grammar frontend.Grammar) (backend.Parser, error) {
	switch parserCore {
	case "ielr1":
		return ielr1core.GrammarToParser(grammar)
	default:
		return backend.Parser{}, fmt.Errorf("unsupported parser core %q", parserCore)
	}
}

func executeParserBackend(parser backend.Parser) error {
	switch parserBackend {
	case "go":
		return golangbackend.ParserToFile(parserBackendFilePath, parser, golangbackend.Config{
			PackageName: parserBackendGoPackageName,
		})
	case "json":
		return jsonbackend.ParserToFile(parserBackendFilePath, parser)
	case "yaml":
		return yamlbackend.ParserToFile(parserBackendFilePath, parser)
	default:
		return fmt.Errorf("unsupported parser backend %q", parserBackend)
	}
}

func init() {
	rootCmd.AddCommand(parserCmd)

	parserCmd.PersistentFlags().StringVar(
		&parserFrontend,
		"frontend",
		"bison",
		"The frontend to use for reading the context free grammar. One of: bison, json, yaml.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserFrontendFilePath,
		"frontend-file-path",
		"",
		"The file path to read the context free grammar from.",
	)
	if err := parserCmd.MarkPersistentFlagRequired("frontend-file-path"); err != nil {
		panic(err)
	}

	parserCmd.PersistentFlags().StringVar(
		&parserCore,
		"core",
		"ielr1",
		"The core to use for generating the parser from the context free grammar. One of: ielr1.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackend,
		"backend",
		"go",
		"The backend to use for writing the parser. One of: go, json, yaml.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserBackendFilePath,
		"backend-file-path",
		"",
		"The file path to write the parser to.",
	)
	if err := parserCmd.MarkPersistentFlagRequired("backend-file-path"); err != nil {
		panic(err)
	}

	parserCmd.PersistentFlags().StringVar(
		&parserBackendGoPackageName,
		"backend-go-package-name",
		"parser",
		"The Go package name to use for the generated Go code.",
	)
}
