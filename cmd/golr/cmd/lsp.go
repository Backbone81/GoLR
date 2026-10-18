package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/lsp"
)

var (
	lspStdio           bool
	lspPipe            string
	lspSocket          int
	lspClientProcessID int
)

var lspCmd = &cobra.Command{
	Use:          "lsp",
	Short:        "Runs the language server for GoLR grammar files.",
	Long:         `Runs the language server for GoLR grammar files.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		options, err := lspOptions(cmd)
		if err != nil {
			return err
		}
		return lsp.Serve(cmd.Context(), options...)
	},
}

// lspOptions returns the language server options the flags ask for. The flags follow the arguments the language
// server protocol specification names for the channels.
func lspOptions(cmd *cobra.Command) ([]lsp.Option, error) {
	var options []lsp.Option
	channels := 0
	if cmd.Flags().Changed("stdio") {
		channels++
	}
	if cmd.Flags().Changed("pipe") {
		channels++
		options = append(options, lsp.WithPipe(lspPipe))
	}
	if cmd.Flags().Changed("socket") {
		channels++
		options = append(options, lsp.WithSocket(lspSocket))
	}
	if channels > 1 {
		return nil, errors.New("only one of --stdio, --pipe and --socket can be given")
	}

	if lspClientProcessID != 0 {
		options = append(options, lsp.WithClientProcessID(lspClientProcessID))
	}
	return options, nil
}

func init() {
	lspCmd.Flags().BoolVar(
		&lspStdio,
		"stdio",
		false,
		"talk to the editor over stdin and stdout, the default",
	)
	lspCmd.Flags().StringVar(
		&lspPipe,
		"pipe",
		"",
		"talk to the editor over the pipe it listens on, a Unix domain socket or a Windows named pipe",
	)
	lspCmd.Flags().IntVar(
		&lspSocket,
		"socket",
		0,
		"talk to the editor over the TCP port on 127.0.0.1 it listens on",
	)
	lspCmd.Flags().IntVar(
		&lspClientProcessID,
		"clientProcessId",
		0,
		"end the session when the editor process with this id is gone",
	)
	rootCmd.AddCommand(lspCmd)
}
