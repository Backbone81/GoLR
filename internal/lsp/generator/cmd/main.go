// Command cmd writes the Go files generated from the meta model of the Language Server Protocol.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/internal/lsp/generator"
)

var (
	metaModelPath string
	outputDir     string
)

var rootCmd = &cobra.Command{
	Use:          "generator",
	Short:        "Generates the Go types of the Language Server Protocol from its meta model.",
	Long:         `Generates the Go types of the Language Server Protocol from its meta model.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		//nolint:gosec // the path needs to be user provided
		data, err := os.ReadFile(metaModelPath)
		if err != nil {
			return fmt.Errorf("reading the meta model: %w", err)
		}

		model, err := generator.ParseMetaModel(data)
		if err != nil {
			return err
		}

		return generator.MetaModelToDirectory(outputDir, model)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&metaModelPath,
		"meta-model",
		"internal/lsp/spec/metaModel.json",
		"the meta model to read",
	)
	rootCmd.PersistentFlags().StringVar(
		&outputDir,
		"output",
		"internal/lsp/protocol",
		"the directory to write the generated files into",
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
