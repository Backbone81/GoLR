package cmd

import (
	"os"

	"github.com/spf13/cobra"

	golrconvert "github.com/backbone81/golr/pkg/convert"
)

var (
	convertInputFilePath  string
	convertOutputFilePath string
)

var convertCmd = &cobra.Command{
	Use:          "convert",
	Short:        "Converts grammar files between different formats.",
	Long:         `Converts grammar files between different formats.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if convertInputFilePath != "-" && convertOutputFilePath != "-" {
			return golrconvert.BisonToGoLRFile(convertInputFilePath, convertOutputFilePath)
		}

		reader := os.Stdin
		if convertInputFilePath != "-" {
			//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
			inputFile, err := os.Open(convertInputFilePath)
			if err != nil {
				return err
			}
			defer inputFile.Close() //nolint:errcheck
			reader = inputFile
		}

		writer := os.Stdout
		if convertOutputFilePath != "-" {
			//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
			outputFile, err := os.Create(convertOutputFilePath)
			if err != nil {
				return err
			}
			defer outputFile.Close() //nolint:errcheck
			writer = outputFile
		}

		return golrconvert.BisonToGoLR(reader, writer, convertInputFilePath)
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.PersistentFlags().StringVar(
		&convertInputFilePath,
		"input-file-path",
		"",
		"The GNU Bison grammar file to convert. Can be '-' to read from stdin.",
	)
	if err := convertCmd.MarkPersistentFlagRequired("input-file-path"); err != nil {
		panic(err)
	}

	convertCmd.PersistentFlags().StringVar(
		&convertOutputFilePath,
		"output-file-path",
		"-",
		"The GoLR grammar file to write. Can be '-' to write to stdout.",
	)
}
