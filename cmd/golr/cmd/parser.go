package cmd

import (
	"fmt"
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
	pythonbackend "github.com/backbone81/golr/pkg/parsergen/backend/python"
	rustbackend "github.com/backbone81/golr/pkg/parsergen/backend/rust"
	typescriptbackend "github.com/backbone81/golr/pkg/parsergen/backend/typescript"
	yamlbackend "github.com/backbone81/golr/pkg/parsergen/backend/yaml"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/pkg/parsergen/core/ielr1/golr"
	lalr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lalr1/bison"
	lr1bisoncore "github.com/backbone81/golr/pkg/parsergen/core/lr1/bison"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/golr"
	jsonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/json"
	yamlfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/yaml"
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
	parserBackendPythonScannerModule     string
	parserBackendRustScannerModule       string
	parserBackendTypeScriptScannerModule string

	parserVerbose bool
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

		parser, conflicts, err := executeParserCore(grammar)
		if err != nil {
			return err
		}

		// The conflicts are reported to stderr so they do not corrupt a backend which writes its output to stdout. The
		// conflicts the policy resolved on its own are only summarized unless --verbose asks for the full listing, so the
		// report stays readable for a large grammar which leans on precedence declarations.
		printConflictReport(os.Stderr, parser.Grammar, conflicts, parserVerbose)

		if err := executeParserBackend(parser); err != nil {
			return err
		}
		return nil
	},
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

func executeParserCore(grammar frontend.Grammar) (backend.Parser, []conflict.Conflict, error) {
	switch parserCore {
	case "ielr1", "ielr1-golr":
		return ielr1golrcore.GrammarToParser(grammar)
	case "ielr1-bison":
		return ielr1bisoncore.GrammarToParser(grammar)
	case "lalr1", "lalr1-golr":
		return lalr1golrcore.GrammarToParser(grammar)
	case "lalr1-bison":
		return lalr1bisoncore.GrammarToParser(grammar)
	case "lr1", "lr1-golr":
		return lr1golrcore.GrammarToParser(grammar)
	case "lr1-bison":
		return lr1bisoncore.GrammarToParser(grammar)
	default:
		return backend.Parser{}, nil, fmt.Errorf("unsupported parser core %q", parserCore)
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
			" javascript, json, null, python, rust, typescript, yaml.",
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
		"List every conflict the parser generator resolved on its own, instead of only summarizing them.",
	)
}
