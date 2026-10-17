package model

// CompletionKind is the place in the grammar a completion context stands for.
type CompletionKind int

const (
	// CompletionNone offers nothing: the name of a rule, before its ":", inside @name(...), and inside a comment,
	// string or regular expression.
	CompletionNone CompletionKind = iota

	// CompletionTopLevel is outside of the sections.
	CompletionTopLevel

	// CompletionScannerRuleBody is after the ":" of a scanner rule.
	CompletionScannerRuleBody

	// CompletionParserStatementStart is where a statement of the parser section starts.
	CompletionParserStatementStart

	// CompletionStartBody is after `@start :`.
	CompletionStartBody

	// CompletionPrecedenceLineStart is where a line of the precedence block starts.
	CompletionPrecedenceLineStart

	// CompletionPrecedenceLineBody is after the ":" of a precedence line.
	CompletionPrecedenceLineBody

	// CompletionProductionBody is after the ":" of a parser rule.
	CompletionProductionBody

	// CompletionPrecedenceArgument is inside @precedence(...) of a production.
	CompletionPrecedenceArgument
)

// Keywords returns the Keywords valid at the place.
func (k CompletionKind) Keywords() []string {
	switch k {
	case CompletionNone, CompletionStartBody, CompletionPrecedenceArgument:
		return nil
	case CompletionTopLevel:
		return []string{"@scanner", "@parser"}
	case CompletionScannerRuleBody:
		return []string{"@empty", "@skip", "@fragment"}
	case CompletionParserStatementStart:
		return []string{"@start", "@precedence"}
	case CompletionPrecedenceLineStart:
		return []string{"@left", "@right", "@none", "@precedence"}
	case CompletionPrecedenceLineBody:
		return []string{"@error"}
	case CompletionProductionBody:
		return []string{"@empty", "@error", "@precedence", "@name"}
	default:
		return nil
	}
}

// Symbols reports whether the names of the Symbols are valid at the place.
func (k CompletionKind) Symbols() bool {
	switch k {
	case CompletionStartBody, CompletionPrecedenceLineBody, CompletionProductionBody, CompletionPrecedenceArgument:
		return true
	case CompletionNone, CompletionTopLevel, CompletionScannerRuleBody, CompletionParserStatementStart,
		CompletionPrecedenceLineStart:
		return false
	default:
		return false
	}
}
