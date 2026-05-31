package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/parsergen/backend"
	dotbackend "github.com/backbone81/golr/pkg/parsergen/backend/dot"
	golangbackend "github.com/backbone81/golr/pkg/parsergen/backend/golang"
	jsonbackend "github.com/backbone81/golr/pkg/parsergen/backend/json"
	yamlbackend "github.com/backbone81/golr/pkg/parsergen/backend/yaml"
	ielr1core "github.com/backbone81/golr/pkg/parsergen/core/ielr1"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/golr"
	jsonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/json"
	yamlfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/yaml"
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
		if parserFrontendFilePath == "-" {
			return bisonfrontend.ToGrammar(os.Stdin, "pipe")
		}
		return bisonfrontend.GrammarFromFile(parserFrontendFilePath)
	case "golr":
		if parserFrontendFilePath == "-" {
			return golrfrontend.ToGrammar(os.Stdin, "pipe")
		}
		return golrfrontend.GrammarFromFile(parserFrontendFilePath)
	case "json":
		if parserFrontendFilePath == "-" {
			return jsonfrontend.ToGrammar(os.Stdin)
		}
		return jsonfrontend.GrammarFromFile(parserFrontendFilePath)
	case "yaml":
		if parserFrontendFilePath == "-" {
			return yamlfrontend.ToGrammar(os.Stdin)
		}
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
	case "dot":
		if parserBackendFilePath == "-" {
			return dotbackend.FromParser(os.Stdout, parser)
		}
		return dotbackend.ParserToFile(parserBackendFilePath, parser)
	case "go":
		if parserBackendFilePath == "-" {
			return golangbackend.FromParser(os.Stdout, parser, golangbackend.Config{
				PackageName: parserBackendGoPackageName,
			})
		}
		return golangbackend.ParserToFile(parserBackendFilePath, parser, golangbackend.Config{
			PackageName: parserBackendGoPackageName,
		})
	case "json":
		if parserBackendFilePath == "-" {
			return jsonbackend.FromParser(os.Stdout, parser)
		}
		return jsonbackend.ParserToFile(parserBackendFilePath, parser)
	case "null":
		// Nothing to do.
		return nil
	case "yaml":
		if parserBackendFilePath == "-" {
			return yamlbackend.FromParser(os.Stdout, parser)
		}
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
		"golr",
		"The frontend to use for reading the context free grammar. One of: bison, golr, json, yaml.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserFrontendFilePath,
		"frontend-file-path",
		"",
		"The file path to read the context free grammar from. Can be '-' to read from stdin.",
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
		"The backend to use for writing the parser. One of: dot, go, json, null, yaml.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserBackendFilePath,
		"backend-file-path",
		"",
		"The file path to write the parser to. Can be '-' to write to stdout.",
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
