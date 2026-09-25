package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	ielr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/ielr1/bison"
	lalr1golrcore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/golr"
	lr1golrcore "github.com/backbone81/golr/pkg/parsergen/core/lr1/golr"
	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/parsergen/backend"
	cbackend "github.com/backbone81/golr/pkg/parsergen/backend/c"
	cppbackend "github.com/backbone81/golr/pkg/parsergen/backend/cpp"
	csharpbackend "github.com/backbone81/golr/pkg/parsergen/backend/csharp"
	dotbackend "github.com/backbone81/golr/pkg/parsergen/backend/dot"
	golangbackend "github.com/backbone81/golr/pkg/parsergen/backend/golang"
	javabackend "github.com/backbone81/golr/pkg/parsergen/backend/java"
	javascriptbackend "github.com/backbone81/golr/pkg/parsergen/backend/javascript"
	jsonbackend "github.com/backbone81/golr/pkg/parsergen/backend/json"
	kotlinbackend "github.com/backbone81/golr/pkg/parsergen/backend/kotlin"
	pythonbackend "github.com/backbone81/golr/pkg/parsergen/backend/python"
	rustbackend "github.com/backbone81/golr/pkg/parsergen/backend/rust"
	typescriptbackend "github.com/backbone81/golr/pkg/parsergen/backend/typescript"
	yamlbackend "github.com/backbone81/golr/pkg/parsergen/backend/yaml"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	ielr1golrcore "github.com/backbone81/golr/pkg/parsergen/core/ielr1/golr"
	lalr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/bison"
	lr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lr1/bison"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/golr"
	jsonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/json"
	yamlfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/yaml"
	"github.com/backbone81/golr/pkg/utils"
)

var (
	parserFrontend         string
	parserFrontendFilePath string

	parserCore string

	parserBackend         string
	parserBackendFilePath string

	parserBackendCPrefix                 string
	parserBackendCScannerInclude         string
	parserBackendCSharpNamespace         string
	parserBackendCppNamespace            string
	parserBackendCppScannerInclude       string
	parserBackendGoPackageName           string
	parserBackendJavaPackageName         string
	parserBackendJavaScriptScannerModule string
	parserBackendKotlinPackageName       string
	parserBackendPythonScannerModule     string
	parserBackendRustScannerModule       string
	parserBackendTypeScriptScannerModule string

	parserVerbose          bool
	parserWithStateNumbers bool

	parserFailOnConflicts             bool
	parserFailOnShiftReduceConflicts  bool
	parserFailOnReduceReduceConflicts bool
	parserFailOnWarnings              bool
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

		// The conflicts are reported to stderr so they do not corrupt a backend which writes its output to stdout. The
		// conflicts the policy resolved on its own are only summarized unless --verbose also asks for the full listing, so
		// the report stays readable for a large grammar.
		reportConfig := conflict.ReportConfig{
			Verbose:          parserVerbose,
			WithStateNumbers: parserWithStateNumbers,
		}

		parser, conflicts, warnings, err := executeParserCore(grammar, parserCoreOptions()...)
		// The warnings come first, also when the core fails, so that none of them gets lost.
		if err := writeWarnings(os.Stderr, warnings); err != nil {
			return err
		}
		if err != nil {
			return reportUnresolvedConflicts(err, conflicts, reportConfig)
		}

		if err := conflict.WriteConflictReport(os.Stderr, parser.Grammar, conflicts, reportConfig); err != nil {
			return err
		}

		if err := executeParserBackend(parser); err != nil {
			return err
		}
		return nil
	},
}

// parserCoreOptions returns the core options the flags ask for.
func parserCoreOptions() []core.Option {
	var options []core.Option
	if parserFailOnConflicts {
		options = append(options, core.FailOnConflicts())
	}
	if parserFailOnShiftReduceConflicts {
		options = append(options, core.FailOnShiftReduceConflicts())
	}
	if parserFailOnReduceReduceConflicts {
		options = append(options, core.FailOnReduceReduceConflicts())
	}
	if parserFailOnWarnings {
		options = append(options, core.FailOnWarnings())
	}
	return options
}

// writeWarnings writes every warning on a line of its own.
func writeWarnings(w io.Writer, warnings []utils.Warning) error {
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(w, "warning: %s\n", warning.Error()); err != nil {
			return err
		}
	}
	return nil
}

// reportUnresolvedConflicts writes the report of the unresolved conflicts the error holds to stderr, headed by the
// counts of all conflicts. It returns the other errors of the error, followed by the count of the unresolved conflicts,
// so they are not printed a second time. An error without unresolved conflicts is returned unchanged.
func reportUnresolvedConflicts(err error, conflicts []conflict.Conflict, config conflict.ReportConfig) error {
	unresolvedConflictErrors := conflict.UnresolvedConflictErrors(err)
	if len(unresolvedConflictErrors) == 0 {
		return err
	}

	if err := conflict.WriteUnresolvedConflictReport(os.Stderr, conflicts, err, config); err != nil {
		return err
	}
	// The error which follows is separated from the last report by an empty line, so it does not read as part of it.
	if _, err := io.WriteString(os.Stderr, "\n"); err != nil {
		return err
	}
	countErr := fmt.Errorf("%d unresolved conflicts", len(unresolvedConflictErrors))
	if len(unresolvedConflictErrors) == 1 {
		countErr = errors.New("1 unresolved conflict")
	}
	return errors.Join(append(otherErrors(err), countErr)...)
}

// otherErrors returns the errors of the error tree which are not an unresolved conflict, in the order they were
// joined.
func otherErrors(err error) []error {
	//nolint:errorlint // errors.As stops at the first match, but every error of a join is needed. The recursion unwraps.
	switch typedErr := err.(type) {
	case conflict.UnresolvedConflictError:
		return nil
	case interface{ Unwrap() []error }:
		var result []error
		for _, wrappedErr := range typedErr.Unwrap() {
			result = append(result, otherErrors(wrappedErr)...)
		}
		return result
	}
	return []error{err}
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
			_, grammar, err := golrfrontend.ToGrammar(os.Stdin, "pipe")
			return grammar, err
		}
		_, grammar, err := golrfrontend.GrammarFromFile(parserFrontendFilePath)
		return grammar, err
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

func executeParserCore(grammar frontend.Grammar, options ...core.Option) (
	backend.Parser,
	[]conflict.Conflict,
	[]utils.Warning,
	error,
) {
	switch parserCore {
	case "ielr1", "ielr1-golr":
		return ielr1golrcore.GrammarToParser(grammar, options...)
	case "ielr1-bison":
		return ielr1bisoncore.GrammarToParser(grammar, options...)
	case "lalr1", "lalr1-golr":
		return lalr1golrcore.GrammarToParser(grammar, options...)
	case "lalr1-bison":
		return lalr1bisoncore.GrammarToParser(grammar, options...)
	case "lr1", "lr1-golr":
		return lr1golrcore.GrammarToParser(grammar, options...)
	case "lr1-bison":
		return lr1bisoncore.GrammarToParser(grammar, options...)
	default:
		return backend.Parser{}, nil, nil, fmt.Errorf("unsupported parser core %q", parserCore)
	}
}

func executeParserBackend(parser backend.Parser) error {
	switch parserBackend {
	case "c":
		if parserBackendFilePath == "-" {
			return cbackend.FromParser(os.Stdout, parser, cbackend.Config{
				Prefix:         parserBackendCPrefix,
				ScannerInclude: parserBackendCScannerInclude,
			})
		}
		return cbackend.ParserToFile(parserBackendFilePath, parser, cbackend.Config{
			Prefix:         parserBackendCPrefix,
			ScannerInclude: parserBackendCScannerInclude,
		})
	case "cpp":
		if parserBackendFilePath == "-" {
			return cppbackend.FromParser(os.Stdout, parser, cppbackend.Config{
				Namespace:      parserBackendCppNamespace,
				ScannerInclude: parserBackendCppScannerInclude,
			})
		}
		return cppbackend.ParserToFile(parserBackendFilePath, parser, cppbackend.Config{
			Namespace:      parserBackendCppNamespace,
			ScannerInclude: parserBackendCppScannerInclude,
		})
	case "csharp":
		if parserBackendFilePath == "-" {
			return csharpbackend.FromParser(os.Stdout, parser, csharpbackend.Config{
				Namespace: parserBackendCSharpNamespace,
			})
		}
		return csharpbackend.ParserToFile(parserBackendFilePath, parser, csharpbackend.Config{
			Namespace: parserBackendCSharpNamespace,
		})
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
	case "java":
		if parserBackendFilePath == "-" {
			return javabackend.FromParser(os.Stdout, parser, javabackend.Config{
				PackageName: parserBackendJavaPackageName,
			})
		}
		return javabackend.ParserToFile(parserBackendFilePath, parser, javabackend.Config{
			PackageName: parserBackendJavaPackageName,
		})
	case "javascript":
		if parserBackendFilePath == "-" {
			return javascriptbackend.FromParser(os.Stdout, parser, javascriptbackend.Config{
				ScannerModule: parserBackendJavaScriptScannerModule,
			})
		}
		return javascriptbackend.ParserToFile(parserBackendFilePath, parser, javascriptbackend.Config{
			ScannerModule: parserBackendJavaScriptScannerModule,
		})
	case "json":
		if parserBackendFilePath == "-" {
			return jsonbackend.FromParser(os.Stdout, parser)
		}
		return jsonbackend.ParserToFile(parserBackendFilePath, parser)
	case "kotlin":
		if parserBackendFilePath == "-" {
			return kotlinbackend.FromParser(os.Stdout, parser, kotlinbackend.Config{
				PackageName: parserBackendKotlinPackageName,
			})
		}
		return kotlinbackend.ParserToFile(parserBackendFilePath, parser, kotlinbackend.Config{
			PackageName: parserBackendKotlinPackageName,
		})
	case "null":
		// Nothing to do.
		return nil
	case "python":
		if parserBackendFilePath == "-" {
			return pythonbackend.FromParser(os.Stdout, parser, pythonbackend.Config{
				ScannerModule: parserBackendPythonScannerModule,
			})
		}
		return pythonbackend.ParserToFile(parserBackendFilePath, parser, pythonbackend.Config{
			ScannerModule: parserBackendPythonScannerModule,
		})
	case "rust":
		if parserBackendFilePath == "-" {
			return rustbackend.FromParser(os.Stdout, parser, rustbackend.Config{
				ScannerModule: parserBackendRustScannerModule,
			})
		}
		return rustbackend.ParserToFile(parserBackendFilePath, parser, rustbackend.Config{
			ScannerModule: parserBackendRustScannerModule,
		})
	case "typescript":
		if parserBackendFilePath == "-" {
			return typescriptbackend.FromParser(os.Stdout, parser, typescriptbackend.Config{
				ScannerModule: parserBackendTypeScriptScannerModule,
			})
		}
		return typescriptbackend.ParserToFile(parserBackendFilePath, parser, typescriptbackend.Config{
			ScannerModule: parserBackendTypeScriptScannerModule,
		})
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
		"The core to use for generating the parser from the context free grammar. One of: ielr1, ielr1-golr, ielr1-bison, lalr1, lalr1-golr, lalr1-bison, lr1, lr1-golr, lr1-bison.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackend,
		"backend",
		"go",
		"The backend to use for writing the parser. One of: c, cpp, csharp, dot, go, java,"+
			" javascript, json, kotlin, null, python, rust, typescript, yaml.",
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
		golangbackend.DefaultPackageName,
		"The Go package name to use for the generated Go code.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendCPrefix,
		"backend-c-prefix",
		cbackend.DefaultPrefix,
		"The prefix to put in front of every name the generated C code declares. Has to be the one the scanner was"+
			" generated with.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserBackendCScannerInclude,
		"backend-c-scanner-include",
		cbackend.DefaultScannerInclude,
		"The header the generated C parser includes the token type from.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendCppNamespace,
		"backend-cpp-namespace",
		cppbackend.DefaultNamespace,
		"The C++ namespace to use for the generated C++ code. Has to be the one the scanner was generated into.",
	)
	parserCmd.PersistentFlags().StringVar(
		&parserBackendCppScannerInclude,
		"backend-cpp-scanner-include",
		cppbackend.DefaultScannerInclude,
		"The header the generated C++ parser includes the token type from.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendCSharpNamespace,
		"backend-csharp-namespace",
		csharpbackend.DefaultNamespace,
		"The C# namespace to use for the generated C# code. Has to be the one the scanner was generated into.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendJavaPackageName,
		"backend-java-package-name",
		javabackend.DefaultPackageName,
		"The Java package name to use for the generated Java code. Has to be the one the scanner was generated into.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendJavaScriptScannerModule,
		"backend-javascript-scanner-module",
		javascriptbackend.DefaultScannerModule,
		"The module specifier the generated JavaScript parser imports the token constants from.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendKotlinPackageName,
		"backend-kotlin-package-name",
		kotlinbackend.DefaultPackageName,
		"The Kotlin package name to use for the generated Kotlin code. Has to be the one the scanner was generated"+
			" into.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendPythonScannerModule,
		"backend-python-scanner-module",
		pythonbackend.DefaultScannerModule,
		"The module the generated Python parser imports the token constants from.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendRustScannerModule,
		"backend-rust-scanner-module",
		rustbackend.DefaultScannerModule,
		"The module path the generated Rust parser takes the token type from.",
	)

	parserCmd.PersistentFlags().StringVar(
		&parserBackendTypeScriptScannerModule,
		"backend-typescript-scanner-module",
		typescriptbackend.DefaultScannerModule,
		"The module specifier the generated TypeScript parser imports the token constants from.",
	)

	parserCmd.PersistentFlags().BoolVarP(
		&parserVerbose,
		"verbose",
		"v",
		false,
		"List every conflict the parser generator resolved on its own.",
	)

	parserCmd.PersistentFlags().BoolVar(
		&parserFailOnConflicts,
		"fail-on-conflicts",
		false,
		"Fail if a shift/reduce or reduce/reduce conflict is not resolved by precedence or associativity.",
	)
	parserCmd.PersistentFlags().BoolVar(
		&parserFailOnShiftReduceConflicts,
		"fail-on-sr-conflicts",
		false,
		"Fail if a shift/reduce conflict is not resolved by precedence or associativity.",
	)
	parserCmd.PersistentFlags().BoolVar(
		&parserFailOnReduceReduceConflicts,
		"fail-on-rr-conflicts",
		false,
		"Fail if a reduce/reduce conflict is not resolved by precedence or associativity.",
	)
	parserCmd.PersistentFlags().BoolVar(
		&parserFailOnWarnings,
		"fail-on-warnings",
		false,
		"Fail if there are warnings.",
	)

	parserCmd.PersistentFlags().BoolVar(
		&parserWithStateNumbers,
		"with-state-number",
		false,
		"Output the state number with every conflicted state.",
	)
}
