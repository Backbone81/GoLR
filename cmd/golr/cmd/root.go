package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "golr",
	Short: "GoLR is a parser generator for LR(1) grammars.",
	Long:  `GoLR is a parser generator for LR(1) grammars.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
