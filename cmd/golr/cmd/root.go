package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"golr/internal/parsergen/backend/golang"
	"golr/internal/parsergen/core/ielr1"
	frontend2 "golr/internal/parsergen/frontend"
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
		// if err := bison.BuildLALR1("tmp/bison-3.8.2.y", "tmp/output-lalr1.xml"); err != nil {
		//	return err
		//}
		//if err := bison.BuildIELR1("tmp/bison-3.8.2.y", "tmp/output-ielr1.xml"); err != nil {
		//	return err
		//}
		//if err := bison.BuildLR1("tmp/bison-3.8.2.y", "tmp/output-lr1.xml"); err != nil {
		//	return err
		//}

		// report, err := bison.LoadBisonXMLReportFromFile("tmp/output-ielr1.xml")
		//if err != nil {
		//	return err
		//}

		parser, err := ielr1.GrammarToParser(frontend2.Grammar{})
		if err != nil {
			return err
		}
		// if err := yaml.ParserToFile("tmp/parser.yaml", parser); err != nil {
		//	return err
		//}
		if err := golang.ParserToFile("examples/bison/parser/parser.go", parser, golang.Config{PackageName: "parser"}); err != nil {
			return err
		}
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
		"backend",
		"yaml",
		"The backend to use for writing the parser.",
	)
}
