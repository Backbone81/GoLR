package spec

import (
	"github.com/backbone81/golr/pkg/scannergen/frontend"
	//nolint:staticcheck // The DSL is intended to be used as dot import.
	. "github.com/backbone81/golr/pkg/scannergen/frontend/dsl"
)

// GetScannerRules returns the rules for generating a scanner for parsing GoLR grammar files.
//
//nolint:funlen // The long function is intended, as it fully describes the scanner rules.
func GetScannerRules() []frontend.Rule {
	var rules []frontend.Rule //nolint:prealloc // No need to preallocate. This is not time-critical.

	newline := Or(
		Literal("\n"),
		Literal("\r\n"),
	)
	horizontalWhitespace := Or(
		Literal(" "),
		Literal("\t"),
		Literal("\f"),
		Literal("\v"),
	)
	whitespaces := OneOrMore(
		Or(
			newline,
			horizontalWhitespace,
		))
	rules = append(rules, Rule("WHITESPACE", whitespaces))

	multilineComment := Concat(
		Literal("/*"),
		ZeroOrMore(
			Or(
				NegCharClass(CharRange('*', '*')),
				Concat(
					Literal("*"),
					NegCharClass(CharRange('/', '/')),
				),
			),
		),
		Literal("*/"),
	)
	lineComment := Concat(
		Literal("//"),
		ZeroOrMore(NegCharClass(CharRange('\n', '\n'))),
	)
	comment := Or(multilineComment, lineComment)
	rules = append(rules, Rule("COMMENT", comment))

	rules = append(rules, Rule("SCANNER", Literal("@scanner")))
	rules = append(rules, Rule("PARSER", Literal("@parser")))
	rules = append(rules, Rule("PRECEDENCE", Literal("@precedence")))
	rules = append(rules, Rule("START", Literal("@start")))
	rules = append(rules, Rule("LEFT", Literal("@left")))
	rules = append(rules, Rule("RIGHT", Literal("@right")))
	rules = append(rules, Rule("NONE", Literal("@none")))
	rules = append(rules, Rule("SKIP", Literal("@skip")))
	rules = append(rules, Rule("EMPTY", Literal("@empty")))
	rules = append(rules, Rule("LBRACE", Literal("{")))
	rules = append(rules, Rule("RBRACE", Literal("}")))
	rules = append(rules, Rule("LPAREN", Literal("(")))
	rules = append(rules, Rule("RPAREN", Literal(")")))
	rules = append(rules, Rule("COLON", Literal(":")))
	rules = append(rules, Rule("SEMI", Literal(";")))
	rules = append(rules, Rule("PIPE", Literal("|")))
	rules = append(rules, Rule("COMMA", Literal(",")))

	name := Concat(
		CharClass(
			CharRange('a', 'z'),
			CharRange('A', 'Z'),
			CharRange('_', '_'),
		),
		ZeroOrMore(
			CharClass(
				CharRange('a', 'z'),
				CharRange('A', 'Z'),
				CharRange('0', '9'),
				CharRange('_', '_'),
			),
		),
	)
	rules = append(rules, Rule("NAME", name))

	regularExpression := Concat(
		Literal(`/`),
		ZeroOrMore(Or(
			NegCharClass(CharRange('/', '/')),
			Literal(`\/`),
		)),
		Literal(`/`),
	)
	rules = append(rules, Rule("REGEX", regularExpression))

	standardString := Concat(
		Literal(`"`),
		ZeroOrMore(Or(
			NegCharClass(CharRange('"', '"')),
			Literal(`\"`),
		)),
		Literal(`"`),
	)
	rules = append(rules, Rule("STRING", standardString))

	return rules
}
