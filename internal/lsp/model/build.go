package model

import (
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// Build returns the symbol model of the text.
//
// The pass is a lenient descent over the tokens, without whitespace and comments. A scanner rule, a parser rule and a
// precedence line end at ";" or "}", a section ends at "}", and all of them end at "@scanner" or "@parser", which
// cannot appear inside a section. A scanner rule also ends at an identifier and a parser rule at `IDENTIFIER :`, which
// start the next rule, so a missing ";" loses no rule. Every loop consumes a token or ends, so the pass ends on any
// text.
//
// Every token the pass consumes names the completion context which follows it, so completion knows the place of the
// caret without a pass of its own.
func Build(text []byte) *Model {
	b := builder{
		text:          text,
		source:        string(text),
		scanner:       parser.NewScanner(text, ""),
		definitionIDs: make([]nameID, 0, initialDefinitions),
		referenceIDs:  make([]nameID, 0, initialReferences),
		model: &Model{
			text:        text,
			names:       make([]string, 0, initialDefinitions),
			nameIDs:     make(map[string]nameID, initialDefinitions),
			definitions: make([]Definition, 0, initialDefinitions),
			references:  make([]Reference, 0, initialReferences),
			occurrences: make([]Occurrence, 0, initialOccurrences),
			completions: make([]CompletionContext, 0, initialCompletions),
		},
	}
	// The window starts on an empty token at offset zero, whose consumption records the first completion context.
	b.readNext()
	b.advance(CompletionTopLevel)
	b.build()
	b.model.index(b.definitionIDs, b.aliasIDs, b.referenceIDs)
	return b.model
}

// The initial capacities of the lists, about a quarter of what PostgreSQL 18.4 converted to golr needs. They save
// most of the allocations of growing the lists for a big grammar.
const (
	initialDefinitions = 320
	initialReferences  = 2100
	initialOccurrences = 2400
	initialCompletions = 1200
)

// builder holds the state of the pass.
type builder struct {
	text    []byte
	scanner *parser.Scanner

	// source is a copy of the text, from which the names are sliced without allocating each of them.
	source string

	// token and tokenSpan are the current token, next and nextSpan the one after it.
	token     parser.Token
	tokenSpan Span
	next      parser.Token
	nextSpan  Span

	// comments are the comments between the current and the next token. Their completion contexts are recorded
	// when the current token is consumed, which keeps the contexts in the order of the text.
	comments []Span

	// definitionIDs, aliasIDs and referenceIDs are the name IDs of the definitions, aliases and references, in the
	// same order.
	definitionIDs []nameID
	aliasIDs      []nameID
	referenceIDs  []nameID

	model *Model
}

// build reads the sections up to the end of the text.
func (b *builder) build() {
	for b.token != parser.EndToken {
		//nolint:exhaustive // Only the section keywords start something.
		switch b.token {
		case parser.TokenScanner:
			b.section(b.scannerStatement, CompletionNone)
		case parser.TokenParser:
			b.section(b.parserStatement, CompletionParserStatementStart)
		default:
			b.advance(CompletionTopLevel)
		}
	}
}

// section reads `@scanner { ... }` or `@parser { ... }`, reading each statement with the given function. The
// completion kind is the one where a statement starts.
func (b *builder) section(statement func(), statementStart CompletionKind) {
	b.advance(CompletionNone)
	b.skip(parser.TokenLbrace, statementStart)
	for !b.atSectionEnd() {
		statement()
	}
	b.skip(parser.TokenRbrace, CompletionTopLevel)
}

// scannerStatement reads a statement of the scanner section, or skips a token which starts none.
func (b *builder) scannerStatement() {
	if b.token == parser.TokenIdentifier {
		b.scannerRule()
		return
	}
	b.advance(CompletionNone)
}

// parserStatement reads a statement of the parser section, or skips a token which starts none.
func (b *builder) parserStatement() {
	//nolint:exhaustive // Every other token starts no statement.
	switch b.token {
	case parser.TokenStart:
		b.startDeclaration()
	case parser.TokenPrecedence:
		b.precedenceBlock()
	case parser.TokenIdentifier:
		b.parserRule()
	default:
		b.advance(CompletionParserStatementStart)
	}
}

// scannerRule reads `NAME : pattern annotations ;`. The name defines a terminal, a fragment if the annotations have
// @fragment, and a string pattern is the alias of a terminal. The rule has no references.
func (b *builder) scannerRule() {
	definitionIdx := b.definition(KindTerminal, CompletionScannerRuleBody)

	var alias Span
	hasAlias := false
	if b.token == parser.TokenString {
		alias = b.tokenSpan
		hasAlias = true
		b.advance(CompletionScannerRuleBody)
	}
	isFragment := false
	for !b.atRuleEnd() && b.token != parser.TokenIdentifier {
		if b.token == parser.TokenFragment {
			isFragment = true
		}
		b.advance(CompletionScannerRuleBody)
	}

	definition := &b.model.definitions[definitionIdx]
	if isFragment {
		definition.Kind = KindFragment
	} else if hasAlias {
		b.addAlias(alias, definition.Name)
	}
	b.skip(parser.TokenSemi, CompletionNone)
}

// parserRule reads `NAME : body ;`. The name defines a nonterminal, and every identifier and string of the body is a
// reference, also inside @precedence(...). The name inside @name(...) is the name of the production, not a symbol.
func (b *builder) parserRule() {
	b.definition(KindNonterminal, CompletionProductionBody)
	for !b.atRuleEnd() && !b.atDefinition() {
		switch {
		case b.isSymbol():
			b.addReference()
			b.advance(CompletionProductionBody)
		case b.token == parser.TokenPrecedence:
			b.precedenceAnnotation()
		case b.token == parser.TokenName:
			b.advance(CompletionNone)
			b.skip(parser.TokenLparen, CompletionNone)
			b.skip(parser.TokenIdentifier, CompletionNone)
			b.skip(parser.TokenRparen, CompletionProductionBody)
		default:
			// "|", @empty, @error and broken tokens.
			b.advance(CompletionProductionBody)
		}
	}
	b.skip(parser.TokenSemi, CompletionParserStatementStart)
}

// precedenceAnnotation reads `@precedence ( SYMBOL )` of a production. The symbol is a reference.
func (b *builder) precedenceAnnotation() {
	b.advance(CompletionNone)
	afterSymbol := CompletionProductionBody
	if b.skip(parser.TokenLparen, CompletionPrecedenceArgument) {
		afterSymbol = CompletionNone
	}
	if b.isSymbol() {
		b.addReference()
		b.advance(afterSymbol)
	}
	b.skip(parser.TokenRparen, CompletionProductionBody)
}

// startDeclaration reads `@start : NAME ;`. The name is a reference.
func (b *builder) startDeclaration() {
	b.advance(CompletionNone)
	b.skip(parser.TokenColon, CompletionStartBody)
	if b.token == parser.TokenIdentifier {
		b.addReference()
		b.advance(CompletionNone)
	}
	b.skip(parser.TokenSemi, CompletionParserStatementStart)
}

// precedenceBlock reads `@precedence { lines }`.
func (b *builder) precedenceBlock() {
	b.advance(CompletionNone)
	b.skip(parser.TokenLbrace, CompletionPrecedenceLineStart)
	for !b.atSectionEnd() {
		if b.isAssociativity() {
			b.precedenceLine()
		} else {
			b.advance(CompletionPrecedenceLineStart)
		}
	}
	b.skip(parser.TokenRbrace, CompletionParserStatementStart)
}

// precedenceLine reads `@left : symbols ;`, also with @right, @none and @precedence. Every identifier and string is a
// reference.
func (b *builder) precedenceLine() {
	b.advance(CompletionNone)
	b.skip(parser.TokenColon, CompletionPrecedenceLineBody)
	for !b.atRuleEnd() {
		if b.isSymbol() {
			b.addReference()
		}
		b.advance(CompletionPrecedenceLineBody)
	}
	b.skip(parser.TokenSemi, CompletionPrecedenceLineStart)
}

// definition records the name at the current token as a definition of the kind, consumes `NAME :` and returns the
// index of the definition. The completion kind is the one of the body after the ":".
func (b *builder) definition(kind Kind, body CompletionKind) int {
	name, id := b.name()
	definition := Definition{
		Name: name,
		Kind: kind,
		Span: b.tokenSpan,
	}
	b.model.definitions = append(b.model.definitions, definition)
	b.definitionIDs = append(b.definitionIDs, id)
	b.addOccurrence(definition.Name, RoleDefinition, definition.Span)
	b.advance(CompletionNone)
	b.skip(parser.TokenColon, body)
	return len(b.model.definitions) - 1
}

// addAlias records the string at the span as the alias of the terminal.
func (b *builder) addAlias(span Span, terminal string) {
	name, id := b.intern(span)
	alias := Alias{
		Name:     name,
		Terminal: terminal,
		Span:     span,
	}
	b.model.aliases = append(b.model.aliases, alias)
	b.aliasIDs = append(b.aliasIDs, id)
	b.addOccurrence(alias.Name, RoleAlias, alias.Span)
}

// addReference records the current token as a reference.
func (b *builder) addReference() {
	name, id := b.name()
	reference := Reference{
		Name: name,
		Span: b.tokenSpan,
	}
	b.model.references = append(b.model.references, reference)
	b.referenceIDs = append(b.referenceIDs, id)
	b.addOccurrence(reference.Name, RoleReference, reference.Span)
}

// addOccurrence records an occurrence. The pass records them in the order of the text: the alias of a scanner rule is
// recorded at the end of the rule, but nothing else is recorded inside a scanner rule.
func (b *builder) addOccurrence(name string, role Role, span Span) {
	b.model.occurrences = append(b.model.occurrences, Occurrence{
		Name: name,
		Role: role,
		Span: span,
	})
}

// addCompletion records the completion kind from the offset on, unless it is the kind already in place.
func (b *builder) addCompletion(offset int, kind CompletionKind) {
	if len(b.model.completions) > 0 && b.model.completions[len(b.model.completions)-1].Kind == kind {
		return
	}
	b.model.completions = append(b.model.completions, CompletionContext{
		Offset: offset,
		Kind:   kind,
	})
}

// name returns the interned text of the current token and its name ID.
func (b *builder) name() (string, nameID) {
	return b.intern(b.tokenSpan)
}

// intern returns the text of the span and its name ID, the same string and ID for the same text.
func (b *builder) intern(span Span) (string, nameID) {
	name := b.source[span.Offset:span.End()]
	if id, ok := b.model.nameIDs[name]; ok {
		return b.model.names[id], id
	}
	id := nameID(len(b.model.names)) //nolint:gosec
	b.model.names = append(b.model.names, name)
	b.model.nameIDs[name] = id
	return name, id
}

// advance consumes the current token, after which completion offers what the kind says, and moves the window to the
// next token.
//
// Nothing is completed inside a comment, string or regular expression. Such a region starts one byte into the token,
// as a caret in front of it is outside. A line comment and an unterminated string, regular expression or comment
// have no closing delimiter, so a caret at their end is still inside, and their region ends one byte later, after the
// newline which ends them.
func (b *builder) advance(kind CompletionKind) {
	b.addCompletion(b.regionEnd(b.token, b.tokenSpan), kind)
	for _, comment := range b.comments {
		b.addCompletion(comment.Offset+1, CompletionNone)
		b.addCompletion(b.regionEnd(parser.TokenComment, comment), kind)
	}
	b.comments = b.comments[:0]

	b.token = b.next
	b.tokenSpan = b.nextSpan
	if b.isText(b.token, b.tokenSpan) {
		b.addCompletion(b.tokenSpan.Offset+1, CompletionNone)
	}
	b.readNext()
}

// readNext reads the next token into the window, keeping the comments in front of it and dropping the whitespace.
func (b *builder) readNext() {
	for b.scanner.Next() {
		token := b.scanner.Token()
		span := Span{
			Offset: b.scanner.ByteOffset(),
			Length: b.scanner.ByteLength(),
		}
		if token == parser.TokenComment {
			b.comments = append(b.comments, span)
			continue
		}
		if token.IsSkipped() {
			continue
		}
		b.next = token
		b.nextSpan = span
		return
	}
	b.next = parser.EndToken
	b.nextSpan = Span{Offset: b.scanner.ByteOffset()}
}

// regionEnd returns the offset after the token from which on the completion context after it applies.
func (b *builder) regionEnd(token parser.Token, span Span) int {
	if b.isUnclosed(token, span) {
		return span.End() + 1
	}
	return span.End()
}

// isText reports whether the token is a comment, string or regular expression, closed or not.
func (b *builder) isText(token parser.Token, span Span) bool {
	return token == parser.TokenComment || token == parser.TokenString || token == parser.TokenRegex ||
		b.isUnclosed(token, span)
}

// isUnclosed reports whether the token is a line comment or an unterminated string, regular expression or comment.
// The scanner ends the invalid token of an unterminated one at the newline or the end of the text.
func (b *builder) isUnclosed(token parser.Token, span Span) bool {
	if span.Length == 0 {
		return false
	}
	first := b.text[span.Offset]
	//nolint:exhaustive // Only comments and invalid tokens can be unclosed.
	switch token {
	case parser.TokenComment:
		return span.Length > 1 && b.text[span.Offset+1] == '/'
	case parser.InvalidToken:
		return first == '"' || first == '/'
	default:
		return false
	}
}

// skip consumes the current token if it is of the given type, after which completion offers what the kind says.
// Reports whether it consumed the token.
func (b *builder) skip(token parser.Token, kind CompletionKind) bool {
	if b.token != token {
		return false
	}
	b.advance(kind)
	return true
}

// isSymbol reports whether the current token names a symbol: an identifier by its name, a string by its alias.
func (b *builder) isSymbol() bool {
	return b.token == parser.TokenIdentifier || b.token == parser.TokenString
}

// isAssociativity reports whether the current token starts a precedence line.
func (b *builder) isAssociativity() bool {
	return b.token == parser.TokenLeft || b.token == parser.TokenRight || b.token == parser.TokenNone ||
		b.token == parser.TokenPrecedence
}

// atDefinition reports whether the current tokens are `IDENTIFIER :`, which start a rule.
func (b *builder) atDefinition() bool {
	return b.token == parser.TokenIdentifier && b.next == parser.TokenColon
}

// atSectionEnd reports whether the current token ends a section or a precedence block.
func (b *builder) atSectionEnd() bool {
	return b.token == parser.TokenRbrace || b.token == parser.TokenScanner || b.token == parser.TokenParser ||
		b.token == parser.EndToken
}

// atRuleEnd reports whether the current token ends a rule or a precedence line.
func (b *builder) atRuleEnd() bool {
	return b.token == parser.TokenSemi || b.atSectionEnd()
}
