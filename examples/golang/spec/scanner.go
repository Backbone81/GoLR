package spec

import (
	"unicode"

	"golr/pkg/scannergen/frontend"
	. "golr/pkg/scannergen/frontend/dsl" //nolint:staticcheck // The DSL is intended to be used as dot import.
)

// GetScannerRules returns the rules for generating a scanner for parsing Go source code. The details can be found
// at https://go.dev/ref/spec#Source_code_representation and https://go.dev/ref/spec#Lexical_elements.
//
//nolint:funlen,maintidx // The long function is intended, as it fully describes the scanner rules.
func GetScannerRules() []frontend.Rule {
	var rules []frontend.Rule //nolint:prealloc // No need to preallocate. This is not time-critical.

	// Characters (https://go.dev/ref/spec#Characters)
	newline := Or(
		Literal("\n"),
		Literal("\r\n"),
	)
	horizontalWhitespace := Or(
		Literal(" "),
		Literal("\t"),
	)
	whitespaces := OneOrMore(
		Or(
			newline,
			horizontalWhitespace,
		))
	rules = append(rules, Rule("WS", whitespaces))

	unicodeChar := CharClass(
		CharRange(0, '\n'-1),
		// DEBUG: Temporarily switched to latin characters only for debugging purposes.
		// CharRange('\n'+1, unicode.MaxRune),
		CharRange('\n'+1, unicode.MaxASCII),
		CharRange('“', '“'),
		CharRange('”', '”'),
		CharRange('…', '…'),
		CharRange('’', '’'),
	)
	unicodeLetter := CharClass(
		// DEBUG: Temporarily switched to latin characters only for debugging purposes.
		// UnicodeCategory(unicode.Letter)...,
		CharRange('a', 'z'),
		CharRange('A', 'Z'),
	)
	unicodeDigit := CharClass(
		// DEBUG: Temporarily switched to latin characters only for debugging purposes.
		// UnicodeCategory(unicode.Number, unicode.Digit)...,
		CharRange('0', '9'),
	)

	// Letters and digits (https://go.dev/ref/spec#Letters_and_digits)
	letter := Or(
		unicodeLetter,
		Literal("_"),
	)
	decimalDigit := CharClass(
		CharRange('0', '9'),
	)
	binaryDigit := CharClass(
		CharRange('0', '1'),
	)
	octalDigit := CharClass(
		CharRange('0', '7'),
	)
	hexDigit := CharClass(
		CharRange('0', '9'),
		CharRange('A', 'F'),
		CharRange('a', 'f'),
	)

	// Comments (https://go.dev/ref/spec#Comments)
	lineComment := Concat(
		Literal("//"),
		ZeroOrMore(unicodeChar),
	)
	generalComment := Concat(
		Literal("/*"),
		ZeroOrMore(
			Or(
				NegCharClass(
					CharRange('*', '*'),
				),
				Concat(
					Literal("*"),
					NegCharClass(
						CharRange('/', '/'),
					),
				),
			),
		),
		Literal("*/"),
	)
	rules = append(rules, Rule("COMMENT", Or(
		lineComment,
		generalComment,
	),
	))

	// Keywords (https://go.dev/ref/spec#Keywords)
	rules = append(rules, Rule("BREAK", Literal("break")))
	rules = append(rules, Rule("CASE", Literal("case")))
	rules = append(rules, Rule("CHAN", Literal("chan")))
	rules = append(rules, Rule("CONST", Literal("const")))
	rules = append(rules, Rule("CONTINUE", Literal("continue")))

	rules = append(rules, Rule("DEFAULT", Literal("default")))
	rules = append(rules, Rule("DEFER", Literal("defer")))
	rules = append(rules, Rule("ELSE", Literal("else")))
	rules = append(rules, Rule("FALLTHROUGH", Literal("fallthrough")))
	rules = append(rules, Rule("FOR", Literal("for")))

	rules = append(rules, Rule("FUNC", Literal("func")))
	rules = append(rules, Rule("GO", Literal("go")))
	rules = append(rules, Rule("GOTO", Literal("goto")))
	rules = append(rules, Rule("IF", Literal("if")))
	rules = append(rules, Rule("IMPORT", Literal("import")))

	rules = append(rules, Rule("INTERFACE", Literal("interface")))
	rules = append(rules, Rule("MAP", Literal("map")))
	rules = append(rules, Rule("PACKAGE", Literal("package")))
	rules = append(rules, Rule("RANGE", Literal("range")))
	rules = append(rules, Rule("RETURN", Literal("return")))

	rules = append(rules, Rule("SELECT", Literal("select")))
	rules = append(rules, Rule("STRUCT", Literal("struct")))
	rules = append(rules, Rule("SWITCH", Literal("switch")))
	rules = append(rules, Rule("TYPE", Literal("type")))
	rules = append(rules, Rule("VAR", Literal("var")))

	// Identifiers (https://go.dev/ref/spec#Identifiers)
	identifier := Concat(
		letter,
		ZeroOrMore(
			Or(
				letter,
				unicodeDigit,
			),
		),
	)
	rules = append(rules, Rule("IDENT", identifier))

	// Operators and punctuation (https://go.dev/ref/spec#Operators_and_punctuation)
	rules = append(rules, Rule("ADD", Literal("+")))
	rules = append(rules, Rule("SUB", Literal("-")))
	rules = append(rules, Rule("MUL", Literal("*")))
	rules = append(rules, Rule("QUO", Literal("/")))
	rules = append(rules, Rule("REM", Literal("%")))

	rules = append(rules, Rule("AND", Literal("&")))
	rules = append(rules, Rule("OR", Literal("|")))
	rules = append(rules, Rule("XOR", Literal("^")))
	rules = append(rules, Rule("SHL", Literal("<<")))
	rules = append(rules, Rule("SHR", Literal(">>")))
	rules = append(rules, Rule("AND_NOT", Literal("&^")))

	rules = append(rules, Rule("ADD_ASSIGN", Literal("+=")))
	rules = append(rules, Rule("SUB_ASSIGN", Literal("-=")))
	rules = append(rules, Rule("MUL_ASSIGN", Literal("*=")))
	rules = append(rules, Rule("QUO_ASSIGN", Literal("/=")))
	rules = append(rules, Rule("REM_ASSIGN", Literal("%=")))

	rules = append(rules, Rule("AND_ASSIGN", Literal("&=")))
	rules = append(rules, Rule("OR_ASSIGN", Literal("|=")))
	rules = append(rules, Rule("XOR_ASSIGN", Literal("^=")))
	rules = append(rules, Rule("SHL_ASSIGN", Literal("<<=")))
	rules = append(rules, Rule("SHR_ASSIGN", Literal(">>=")))
	rules = append(rules, Rule("AND_NOT_ASSIGN", Literal("&^=")))

	rules = append(rules, Rule("LAND", Literal("&&")))
	rules = append(rules, Rule("LOR", Literal("||")))
	rules = append(rules, Rule("ARROW", Literal("<-")))
	rules = append(rules, Rule("INC", Literal("++")))
	rules = append(rules, Rule("DEC", Literal("--")))

	rules = append(rules, Rule("EQL", Literal("==")))
	rules = append(rules, Rule("LSS", Literal("<")))
	rules = append(rules, Rule("GTR", Literal(">")))
	rules = append(rules, Rule("ASSIGN", Literal("=")))
	rules = append(rules, Rule("NOT", Literal("!")))
	rules = append(rules, Rule("TILDE", Literal("~")))

	rules = append(rules, Rule("NEQ", Literal("!=")))
	rules = append(rules, Rule("LEQ", Literal("<=")))
	rules = append(rules, Rule("GEQ", Literal(">=")))
	rules = append(rules, Rule("DEFINE", Literal(":=")))
	rules = append(rules, Rule("ELLIPSIS", Literal("...")))

	rules = append(rules, Rule("LPAREN", Literal("(")))
	rules = append(rules, Rule("LBRACK", Literal("[")))
	rules = append(rules, Rule("LBRACE", Literal("{")))
	rules = append(rules, Rule("COMMA", Literal(",")))
	rules = append(rules, Rule("PERIOD", Literal(".")))

	rules = append(rules, Rule("RPAREN", Literal(")")))
	rules = append(rules, Rule("RBRACK", Literal("]")))
	rules = append(rules, Rule("RBRACE", Literal("}")))
	rules = append(rules, Rule("SEMICOLON", Literal(";")))
	rules = append(rules, Rule("COLON", Literal(":")))

	// Integer literals (https://go.dev/ref/spec#Integer_literals)
	decimalDigits := Concat(
		decimalDigit,
		ZeroOrMore(
			Concat(
				Optional(
					Literal("_"),
				),
				decimalDigit,
			),
		),
	)
	decimalLit := Or(
		Literal("0"),
		Concat(
			CharClass(
				CharRange('1', '9'),
			),
			Optional(
				Concat(
					Optional(
						Literal("_"),
					),
					decimalDigits,
				),
			),
		),
	)
	binaryDigits := Concat(
		binaryDigit,
		ZeroOrMore(
			Concat(
				Optional(
					Literal("_"),
				),
				binaryDigit,
			),
		),
	)
	binaryLit := Concat(
		Literal("0"),
		Or(
			Literal("b"),
			Literal("B"),
		),
		Optional(
			Literal("_"),
		),
		binaryDigits,
	)
	octalDigits := Concat(
		octalDigit,
		ZeroOrMore(
			Concat(
				Optional(
					Literal("_"),
				),
				octalDigit,
			),
		),
	)
	octalLit := Concat(
		Literal("0"),
		Or(
			Literal("o"),
			Literal("O"),
		),
		Optional(
			Literal("_"),
		),
		octalDigits,
	)
	hexDigits := Concat(
		hexDigit,
		ZeroOrMore(
			Concat(
				Optional(
					Literal("_"),
				),
				hexDigit,
			),
		),
	)
	hexLit := Concat(
		Literal("0"),
		Or(
			Literal("x"),
			Literal("X"),
		),
		Optional(
			Literal("_"),
		),
		hexDigits,
	)
	intLit := Or(
		decimalLit,
		binaryLit,
		octalLit,
		hexLit,
	)
	rules = append(rules, Rule("INT", intLit))

	// Floating-point literals (https://go.dev/ref/spec#Floating-point_literals)
	decimalExponent := Concat(
		Or(
			Literal("e"),
			Literal("E"),
		),
		Optional(
			Or(
				Literal("+"),
				Literal("-"),
			),
		),
		decimalDigits,
	)
	decimalFloatLit := Or(
		Concat(
			decimalDigits,
			Literal("."),
			Optional(
				decimalDigits,
			),
			Optional(
				decimalExponent,
			),
		),
		Concat(
			decimalDigits,
			decimalExponent,
		),
		Concat(
			Literal("."),
			decimalDigits,
			Optional(
				decimalExponent,
			),
		),
	)
	hexMantissa := Or(
		Concat(
			Optional(
				Literal("_"),
			),
			hexDigits,
			Literal("."),
			Optional(
				hexDigits,
			),
		),
		Concat(
			Optional(
				Literal("_"),
			),
			hexDigits,
		),
		Concat(
			Literal("."),
			hexDigits,
		),
	)
	hexExponent := Concat(
		Or(
			Literal("p"),
			Literal("P"),
		),
		Optional(
			Or(
				Literal("+"),
				Literal("-"),
			),
		),
		decimalDigits,
	)
	hexFloatLit := Concat(
		Literal("0"),
		Or(
			Literal("x"),
			Literal("X"),
		),
		hexMantissa,
		hexExponent,
	)
	floatLit := Or(
		decimalFloatLit,
		hexFloatLit,
	)
	rules = append(rules, Rule("FLOAT", floatLit))

	// Imaginary literals (https://go.dev/ref/spec#Imaginary_literals)
	imaginaryLit := Concat(
		Or(
			decimalDigits,
			intLit,
			floatLit,
		),
		Literal("i"),
	)
	rules = append(rules, Rule("IMAG", imaginaryLit))

	// Rune literals (https://go.dev/ref/spec#Rune_literals)
	octalByteValue := Concat(
		Literal(`\`),
		octalDigit,
		octalDigit,
		octalDigit,
	)
	hexByteValue := Concat(
		Literal(`\x`),
		hexDigit,
		hexDigit,
	)
	byteValue := Or(
		octalByteValue,
		hexByteValue,
	)
	littleUValue := Concat(
		Literal(`\u`),
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
	)
	bigUValue := Concat(
		Literal(`\U`),
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
		hexDigit,
	)
	escapedChar := Concat(
		Literal(`\`),
		Or(
			Literal("a"),
			Literal("b"),
			Literal("f"),
			Literal("n"),
			Literal("r"),
			Literal("t"),
			Literal("v"),
			Literal(`\`),
			Literal("'"),
			Literal(`"`),
		),
	)
	unicodeValue := Or(
		unicodeChar,
		littleUValue,
		bigUValue,
		escapedChar,
	)
	runeLit := Concat(
		Literal("'"),
		Or(
			unicodeValue,
			byteValue,
		),
		Literal("'"),
	)
	rules = append(rules, Rule("CHAR", runeLit))

	// String literals (https://go.dev/ref/spec#String_literals)
	rawStringLit := Concat(
		Literal("`"),
		ZeroOrMore(
			NegCharClass(
				CharRange('`', '`'),
			),
		),
		Literal("`"),
	)
	interpretedStringLit := Concat(
		Literal(`"`),
		ZeroOrMore(
			Or(
				NegCharClass(
					CharRange('\n', '\n'),
					CharRange('"', '"'),
				),
				littleUValue,
				bigUValue,
				escapedChar,
				byteValue,
			),
		),
		Literal(`"`),
	)
	stringLit := Or(
		rawStringLit,
		interpretedStringLit,
	)
	rules = append(rules, Rule("STRING", stringLit))
	return rules
}
