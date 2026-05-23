package cmd

import (
	"github.com/spf13/cobra"

	golrfmt "github.com/backbone81/golr/pkg/fmt"
)

var fmtCmd = &cobra.Command{
	Use:          "fmt [file...]",
	Short:        "Pretty prints GoLR grammar files.",
	Long:         `Pretty prints GoLR grammar files. All comments will be removed.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, filePath := range args {
			if err := golrfmt.GoLRFile(filePath, filePath); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)
}
