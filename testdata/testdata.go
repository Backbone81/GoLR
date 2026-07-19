package testdata

import "embed"

//go:embed *.y
var grammarFS embed.FS

// WellKnownGrammar describes the details of a well known grammar usually used for testing and benchmarking.
type WellKnownGrammar struct {
	// Title provides a human-readable name for the well-known grammar.
	Title        string

	// FileName provides the on-disk file name for the well-known grammar.
	FileName     string

	// Content provides the content of the well-known grammar.
	Content      []byte

	// Terminals provides the number of terminals this grammar contains.
	Terminals    int

	// Nonterminals provides the number of nonterminals this grammar contains.
	Nonterminals int

	// Productions provides the number of productions this grammar contains.
	Productions  int
}

// WellKnownGrammars provides a list of well-known grammars to use for tests and benchmarks.
var WellKnownGrammars = []WellKnownGrammar{
	{
		Title:    "GNU Bison 3.8.2",
		FileName: "bison-3.8.2.y",

		// All %token declarations + error token
		Terminals: 58 + 1,

		// All left hand sides of productions
		// Note that grammar_declaration shows up twice and must be counted only once.
		Nonterminals: 38,

		// All productions + alternatives
		// Note that not all alternatives start in the first column. symbols.1 has an alternative which is indented
		// and therefore easy to overlook with a regex search in the grammar file.
		Productions: 39 + 80,
	},
	{
		Title:    "GCC 2.95.3 C",
		FileName: "gcc-2.95.3-c.y",

		// All %token declarations + error token + %left + %right + %nonassoc + char literals
		Terminals: 47 + 1 + 19 + 7 + 2 + 6,

		// All left hand sides of productions
		// Note that the production for all_iter_stmt_with_decl is commented out and needs to be removed from the
		// list for a correct count.
		Nonterminals: 117,

		// All productions + alternatives
		// Note that some alternatives are commented out and need to be removed from the count.
		Productions: 117 + 247,
	},
	{
		Title:    "GCC 2.95.3 Objective C",
		FileName: "gcc-2.95.3-objc.y",

		// All %token declarations + error token + %left + %right + %nonassoc + char literals
		Terminals: 47 + 1 + 19 + 7 + 2 + 6,

		// All left hand sides of productions
		// Note that the production for all_iter_stmt_with_decl is commented out and needs to be removed from the
		// list for a correct count.
		Nonterminals: 162,

		// All productions + alternatives
		// Note that some alternatives are commented out and need to be removed from the count.
		Productions: 162 + 340,
	},
	{
		Title:    "GCC 3.3.6 C++",
		FileName: "gcc-3.3.6-cpp.y",

		// All %token declarations + %left + %right + %nonassoc + char literals
		// Note that some terminals show up as duplicates between %token and %nonassoc or %left and need to be
		// counted once only.
		Terminals: 68 + 32 + 9 + 3,

		// All left hand sides of productions
		// Note that error was declared as a token and therefore does not show up in the list of nonterminals. In
		// addition the rule for primary_no_id is commented out and needs to be rmeoved.
		Nonterminals: 238,

		// All productions + alternatives
		// Note that some alternatives are commented out and need to be removed from the count.
		Productions: 238 + 633,
	},
	{
		Title:    "GCC 4.2.4 Java",
		FileName: "gcc-4.2.4-java.y",

		// All %token declarations + error token
		// Note that there are duplicate %token declarations to assign a tag after declaration. Searching for all
		// %token declarations would therefore result in duplicate tokens.
		Terminals: 109 + 1,

		// All left hand sides of productions
		// Note that searching for identifiers at the start of the line with a colon at the end will turn up results
		// in comments which need to be ignored.
		Nonterminals: 153,

		// All productions + alternatives
		// Note that one alternative is inside of a block comment starting with "Screws up thing". We need to remove
		// that from the result.
		Productions: 153 + 352,
	},
	{
		Title:    "Go 1.5.4",
		FileName: "go-1.5.4.y",

		// All %token declarations + error token + %left + char literals
		// Note that some %left declarations are identical to %token and should not be counted twice.
		Terminals: 46 + 1 + 3 + 24,

		// All left hand sides of productions
		Nonterminals: 127,

		// All productions + alternatives
		Productions: 127 + 210,
	},
	{
		Title:    "PHP 8.6.7",
		FileName: "php-8.6.7.y",

		// All %token declarations + error token + char literals + %precedence
		Terminals: 154 + 1 + 27 + 2,

		// All left hand sides of productions
		Nonterminals: 177,

		// All productions + alternatives
		Productions: 177 + 446,
	},
	{
		Title:    "PostgreSQL 18.4",
		FileName: "postgres-18.4.y",

		// All %token declarations + error token + %left + char literals
		// Note that some %left declarations are identical to %token and should not be counted twice.
		Terminals: 521 + 1 + 1 + 17,

		// All left hand sides of productions
		Nonterminals: 733,

		// All productions + alternatives
		// Note that some a_expr and json_name_and_value alternatives are commented out
		Productions: 733 + 2704 - 3,
	},
}

func init() {
	// We read the content of the grammar file and fill the Content attributes of all well-known grammars for ease of
	// access.
	for i := range WellKnownGrammars {
		data, err := grammarFS.ReadFile(WellKnownGrammars[i].FileName)
		if err != nil {
			panic(err)
		}
		WellKnownGrammars[i].Content = data
	}
}
