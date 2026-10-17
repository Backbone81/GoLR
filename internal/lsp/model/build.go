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
func Build(text []byte) *Model {
	b := builder{
		text:   text,
		tokens: parser.NewTokenSkipper(parser.NewScanner(text, "")),
		names:  map[string]string{},
		model:  &Model{},
	}
	// Fill the window with the first two tokens.
	b.advance()
	b.advance()
	b.build()
	b.model.index()
	return b.model
}

// builder holds the state of the pass.
type builder struct {
	text   []byte
	tokens *parser.TokenSkipper

	// token and tokenSpan are the current token, next and nextSpan the one after it.
	token     parser.Token
	tokenSpan Span
	next      parser.Token
	nextSpan  Span

	// names interns the names, so all occurrences of a name share one string.
	names map[string]string

	model *Model
}

// build reads the sections up to the end of the text.
func (b *builder) build() {
	for b.token != parser.EndToken {
		//nolint:exhaustive // Only the section keywords start something.
		switch b.token {
		case parser.TokenScanner:
			b.section(b.scannerStatement)
		case parser.TokenParser:
			b.section(b.parserStatement)
		default:
			b.advance()
		}
	}
}

// section reads `@scanner { ... }` or `@parser { ... }`, reading each statement with the given function.
func (b *builder) section(statement func()) {
	b.advance()
	b.skip(parser.TokenLbrace)
	for !b.atSectionEnd() {
		statement()
	}
	b.skip(parser.TokenRbrace)
}

// scannerStatement reads a statement of the scanner section, or skips a token which starts none.
func (b *builder) scannerStatement() {
	if b.token == parser.TokenIdentifier {
		b.scannerRule()
		return
	}
	b.advance()
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
		b.advance()
	}
}

// scannerRule reads `NAME : pattern annotations ;`. The name defines a terminal, a fragment if the annotations have
// @fragment, and a string pattern is the alias of a terminal. The rule has no references.
func (b *builder) scannerRule() {
	definitionIdx := b.definition(KindTerminal)

	var alias Span
	hasAlias := false
	if b.token == parser.TokenString {
		alias = b.tokenSpan
		hasAlias = true
		b.advance()
	}
	isFragment := false
	for !b.atRuleEnd() && b.token != parser.TokenIdentifier {
		if b.token == parser.TokenFragment {
			isFragment = true
		}
		b.advance()
	}

	definition := &b.model.definitions[definitionIdx]
	if isFragment {
		definition.Kind = KindFragment
	} else if hasAlias {
		b.addAlias(alias, definition.Name)
	}
	b.skip(parser.TokenSemi)
}

// parserRule reads `NAME : body ;`. The name defines a nonterminal, and every identifier and string of the body is a
// reference, also inside @precedence(...). The name inside @name(...) is the name of the production, not a symbol.
func (b *builder) parserRule() {
	b.definition(KindNonterminal)
	for !b.atRuleEnd() && !b.atDefinition() {
		switch {
		case b.isSymbol():
			b.addReference()
			b.advance()
		case b.token == parser.TokenPrecedence:
			b.advance()
			b.skip(parser.TokenLparen)
			if b.isSymbol() {
				b.addReference()
				b.advance()
			}
			b.skip(parser.TokenRparen)
		case b.token == parser.TokenName:
			b.advance()
			b.skip(parser.TokenLparen)
			b.skip(parser.TokenIdentifier)
			b.skip(parser.TokenRparen)
		default:
			// "|", @empty, @error and broken tokens.
			b.advance()
		}
	}
	b.skip(parser.TokenSemi)
}

// startDeclaration reads `@start : NAME ;`. The name is a reference.
func (b *builder) startDeclaration() {
	b.advance()
	b.skip(parser.TokenColon)
	if b.token == parser.TokenIdentifier {
		b.addReference()
		b.advance()
	}
	b.skip(parser.TokenSemi)
}

// precedenceBlock reads `@precedence { lines }`.
func (b *builder) precedenceBlock() {
	b.advance()
	b.skip(parser.TokenLbrace)
	for !b.atSectionEnd() {
		if b.isAssociativity() {
			b.precedenceLine()
		} else {
			b.advance()
		}
	}
	b.skip(parser.TokenRbrace)
}

// precedenceLine reads `@left : symbols ;`, also with @right, @none and @precedence. Every identifier and string is a
// reference.
func (b *builder) precedenceLine() {
	b.advance()
	b.skip(parser.TokenColon)
	for !b.atRuleEnd() {
		if b.isSymbol() {
			b.addReference()
		}
		b.advance()
	}
	b.skip(parser.TokenSemi)
}

// definition records the name at the current token as a definition of the kind, consumes `NAME :` and returns the
// index of the definition.
func (b *builder) definition(kind Kind) int {
	definition := Definition{
		Name: b.name(),
		Kind: kind,
		Span: b.tokenSpan,
	}
	b.model.definitions = append(b.model.definitions, definition)
	b.addOccurrence(definition.Name, RoleDefinition, definition.Span)
	b.advance()
	b.skip(parser.TokenColon)
	return len(b.model.definitions) - 1
}

// addAlias records the string at the span as the alias of the terminal.
func (b *builder) addAlias(span Span, terminal string) {
	alias := Alias{
		Name:     b.intern(b.text[span.Offset:span.End()]),
		Terminal: terminal,
		Span:     span,
	}
	b.model.aliases = append(b.model.aliases, alias)
	b.addOccurrence(alias.Name, RoleAlias, alias.Span)
}

// addReference records the current token as a reference.
func (b *builder) addReference() {
	reference := Reference{
		Name: b.name(),
		Span: b.tokenSpan,
	}
	b.model.references = append(b.model.references, reference)
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

// name returns the interned text of the current token.
func (b *builder) name() string {
	return b.intern(b.text[b.tokenSpan.Offset:b.tokenSpan.End()])
}

// intern returns the string of the bytes, the same string for the same bytes.
func (b *builder) intern(text []byte) string {
	if name, ok := b.names[string(text)]; ok {
		return name
	}
	name := string(text)
	b.names[name] = name
	return name
}

// advance moves the window to the next token.
func (b *builder) advance() {
	b.token = b.next
	b.tokenSpan = b.nextSpan
	b.tokens.Next()
	b.next = b.tokens.Token()
	b.nextSpan = Span{
		Offset: b.tokens.ByteOffset(),
		Length: b.tokens.ByteLength(),
	}
}

// skip consumes the current token if it is of the given type.
func (b *builder) skip(token parser.Token) {
	if b.token == token {
		b.advance()
	}
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
