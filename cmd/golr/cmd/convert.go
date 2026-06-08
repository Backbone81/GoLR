package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/backbone81/golr/pkg/convert"
	parserfrontend "github.com/backbone81/golr/pkg/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/golr"
	scannerfrontend "github.com/backbone81/golr/pkg/scannergen/frontend"
	"github.com/spf13/cobra"
)

var (
	convertInputFilePath    string
	convertInputFileFormat  string
	convertOutputFilePath   string
	convertOutputFileFormat string
)

var convertCmd = &cobra.Command{
	Use:          "convert",
	Short:        "Converts grammar files between different formats.",
	Long:         `Converts grammar files between different formats.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, grammar, err := executeInput()
		if err != nil {
			return err
		}
		if err := executeOutput(rules, grammar); err != nil {
			return err
		}
		return nil
	},
}

func executeInput() ([]scannerfrontend.Rule, parserfrontend.Grammar, error) {
	if convertInputFileFormat == "auto" {
		switch filepath.Ext(convertInputFilePath) {
		case ".y":
			convertInputFileFormat = "bison"
		case ".golr":
			convertInputFileFormat = "golr"
		default:
			return nil, parserfrontend.Grammar{}, fmt.Errorf("input file format cannot be detected from input file path extension")
		}
	}

	switch convertInputFileFormat {
	case "bison":
		if convertInputFilePath == "-" {
			grammar, err := bisonfrontend.ToGrammar(os.Stdin, "pipe")
			return nil, grammar, err
		}
		grammar, err := bisonfrontend.GrammarFromFile(convertInputFilePath)
		return nil, convert.BisonGrammar2GoLR(grammar), err
	case "golr":
		if convertInputFilePath == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, parserfrontend.Grammar{}, err
			}
			return golrfrontend.GrammarFromString(string(data))
		}
		return golrfrontend.GrammarFromFile(convertInputFilePath)
	default:
		return nil, parserfrontend.Grammar{}, fmt.Errorf("unsupported input file format %q", convertInputFileFormat)
	}
}

func executeOutput(rules []scannerfrontend.Rule, grammar parserfrontend.Grammar) error {
	if convertOutputFileFormat == "auto" {
		switch filepath.Ext(convertOutputFilePath) {
		case ".y":
			convertOutputFileFormat = "bison"
		case ".golr":
			convertOutputFileFormat = "golr"
		default:
			return fmt.Errorf("output file format cannot be detected from output file path extension")
		}
	}

	switch convertOutputFileFormat {
	case "bison":
		if convertOutputFilePath == "-" {
			return bisonfrontend.FromGrammar(os.Stdout, grammar)
		}
		return bisonfrontend.GrammarToFile(convertOutputFilePath, grammar)
	case "golr":
		if convertOutputFilePath == "-" {
			return golrfrontend.FromGrammar(os.Stdout, rules, grammar)
		}
		return golrfrontend.GrammarToFile(convertOutputFilePath, rules, grammar)
	default:
		return fmt.Errorf("unsupported output file format %q", convertInputFileFormat)
	}
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.PersistentFlags().StringVar(
		&convertInputFilePath,
		"input-file-path",
		"",
		"The grammar file to read. Can be '-' to read from stdin.",
	)
	if err := convertCmd.MarkPersistentFlagRequired("input-file-path"); err != nil {
		panic(err)
	}

	convertCmd.PersistentFlags().StringVar(
		&convertInputFileFormat,
		"input-file-format",
		"auto",
		"The format of the grammar file to read. Format auto derives the format from the file extension. One of: auto, bison, golr.",
	)

	convertCmd.PersistentFlags().StringVar(
		&convertOutputFilePath,
		"output-file-path",
		"",
		"The grammar file to write. Can be '-' to write to stdout.",
	)
	if err := convertCmd.MarkPersistentFlagRequired("output-file-path"); err != nil {
		panic(err)
	}

	convertCmd.PersistentFlags().StringVar(
		&convertOutputFileFormat,
		"output-file-format",
		"auto",
		"The format of the grammar file to write. Format auto derives the format from the file extension. One of: auto, bison, golr.",
	)
}
