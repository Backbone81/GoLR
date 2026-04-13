package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	frontend string
	core     string
	backend  string
)

var rootCmd = &cobra.Command{
	Use:          "golr",
	Short:        "GoLR is a parser generator for LR(1) grammars.",
	Long:         `GoLR is a parser generator for LR(1) grammars.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&frontend,
		"frontend",
		"yaml",
		"The frontend to use for reading the grammar.",
	)
	rootCmd.PersistentFlags().StringVar(
		&core,
		"core",
		"ielr1",
		"The core to use for generating the parser from the grammar.",
	)
	rootCmd.PersistentFlags().StringVar(
		&backend,
		"core",
		"yaml",
		"The backend to use for writing the parser.",
	)
}
