package cmd

import (
	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/lsp"
)

var (
	lspStdio bool
)

var lspCmd = &cobra.Command{
	Use:          "lsp",
	Short:        "Runs the language server for GoLR grammar files.",
	Long:         `Runs the language server for GoLR grammar files.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	// This command is hidden until the language server provides functionality comparable to the intellij and vscode
	// plugins.
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return lsp.Serve(cmd.Context())
	},
}

func init() {
	lspCmd.PersistentFlags().BoolVar(
		&lspStdio,
		"stdio",
		true,
		"talk to the editor over stdin and stdout",
	)
	rootCmd.AddCommand(lspCmd)
}
