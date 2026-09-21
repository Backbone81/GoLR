//nolint:exhaustive // The TreeWalker will only descend selected nonterminals and does not need to be exhaustive.
package golr

import (
	"fmt"
	"math"

	parsergenfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/regex"
	scannergenfrontend "github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/utils"
	"github.com/backbone81/golr/pkg/scannergen/frontend/dsl"
)

// TreeWalker is a helper struct which walks the parse tree of a parsed GoLR grammar and extracts all
// information required to describe the context free grammar therein.
type TreeWalker struct {
	// scanner is the source the tree was parsed from. A node carries the span of the source it covers and not the
	// source itself, so resolving a node back into its bytes goes through the scanner.
	scanner parser.TokenSource

	rules   []scannergenfrontend.Rule
	grammar parsergenfrontend.Grammar

	terminalIdxByName    map[string]int
	terminalIdxByAlias   map[string]int
	nonterminalIdxByName map[string]int

	// startNonterminalName keeps track of the start symbol declared with @start.
	startNonterminalName string

	// startNonterminalByteOffset is where the start symbol declared with @start is in the source.
	startNonterminalByteOffset int

	// nonterminalByteOffsetByName holds where every nonterminal first appears in the source, so an error about a
	// nonterminal which is never defined can point at its first reference.
	nonterminalByteOffsetByName map[string]int

	// currentPrecedence is the current precedence level. The first precedence declared has the highest priority. Every
	// following precedence is decremented by one for every @left, @right and @none
	// declaration.
	currentPrecedence int

	// currentAssociativity keeps track of the associativity being declared in the current precedence_decl.
	currentAssociativity parsergenfrontend.Associativity

	// inPrecedenceDecl keeps track if we are currently inside a precedence_decl. Symbols visited via symbol_list
	// inside a precedence_decl are terminals being assigned the current precedence level and associativity.
	inPrecedenceDecl bool

	// inAlternativeAnnotation keeps track if we are currently inside an alternative_annotation. The symbol visited
	// inside an alternative_annotation is the terminal used to override the precedence for the current production.
	inAlternativeAnnotation bool

	// isFragment keeps track if we found a @fragment annotation. This means that we need to move the token declaration
	// to the fragment list.
	isFragment bool

	// lexemeByName holds all lexemes of token declarations. This is needed to make fragments work.
	lexemeByName map[string][]byte

	// lexemeByteOffsetByName holds where the lexeme of every token declaration is in the source.
	lexemeByteOffsetByName map[string]int

	// explicitProductionNames holds every name set with an @name annotation so far, so a second production asking for
	// the same name can be rejected.
	explicitProductionNames map[string]struct{}
}

// NewTreeWalker creates a new TreeWalker which resolves the spans of the tree against the given scanner.
func NewTreeWalker(scanner parser.TokenSource) *TreeWalker {
	return &TreeWalker{
		scanner:                     scanner,
		terminalIdxByName:           make(map[string]int),
		terminalIdxByAlias:          make(map[string]int),
		nonterminalIdxByName:        make(map[string]int),
		nonterminalByteOffsetByName: make(map[string]int),
		currentPrecedence:           math.MaxInt,
		lexemeByName:                make(map[string][]byte),
		lexemeByteOffsetByName:      make(map[string]int),
		explicitProductionNames:     make(map[string]struct{}),
	}
}

// text returns the bytes of the source the given node covers, as a view into the source rather than a copy of it.
func (w *TreeWalker) text(node *parser.Node) []byte {
	return w.scanner.Text(node.ByteOffset, node.ByteLength)
}

// errorAt returns an error with the given message, prefixed with the file path, line and column of the given byte
// offset of the source.
func (w *TreeWalker) errorAt(byteOffset int, format string, args ...any) error {
	position := w.scanner.Position(byteOffset)
	return fmt.Errorf("%s:%d:%d: %w", position.FilePath, position.Line, position.Column, fmt.Errorf(format, args...))
}

// BuildGrammar takes the root node of the parse tree, traverses the tree to build the context free grammar
// and returns the finished grammar afterward.
func (w *TreeWalker) BuildGrammar(node parser.Node) ([]scannergenfrontend.Rule, parsergenfrontend.Grammar, error) {
	if err := w.visitFile(&node); err != nil {
		return nil, parsergenfrontend.Grammar{}, err
	}
	if w.startNonterminalName != "" {
		idx, ok := w.nonterminalIdxByName[w.startNonterminalName]
		if !ok {
			return nil, parsergenfrontend.Grammar{}, w.errorAt(
				w.startNonterminalByteOffset,
				"unknown start nonterminal %q",
				w.startNonterminalName,
			)
		}
		w.grammar.StartNonterminalIdx = idx
	}

	// Validate that every nonterminal referenced on any production right hand side is also defined on a left hand side.
	definedNonterminals := make(map[int]struct{}, len(w.grammar.Nonterminals))
	for _, production := range w.grammar.Productions {
		definedNonterminals[production.NonterminalIdx] = struct{}{}
	}
	for idx, nonterminal := range w.grammar.Nonterminals {
		if _, ok := definedNonterminals[idx]; !ok {
			return nil,
				parsergenfrontend.Grammar{},
				w.errorAt(
					w.nonterminalByteOffsetByName[nonterminal.Name],
					"nonterminal %q is referenced but never defined",
					nonterminal.Name,
				)
		}
	}

	if len(w.grammar.Productions) < 1 {
		// There is no single place a missing production belongs to, so the error points at the end of the grammar.
		return nil, parsergenfrontend.Grammar{}, w.errorAt(
			node.ByteOffset+node.ByteLength,
			"grammar requires at least one production",
		)
	}

	// Nonterminals are interned in order of first appearance, which counts right hand side references (see
	// visitProductionDecl and visitSymbolInAlternative for the interning sites). That makes a nonterminal referenced
	// before it is defined get a lower index than nonterminals declared earlier. Renumber them into declaration order so
	// the indices match what the Bison-backed core assigns, which keeps generated parsers diff-friendly across cores.
	parsergenfrontend.RenumberNonterminalsInDeclarationOrder(&w.grammar)

	return w.rules, w.grammar, nil
}

// internErrorTerminal adds the error symbol to the grammar the first time a production references it and returns its
// name, so that the symbol lookups which follow resolve it through the same terminal lookup as any other symbol. The
// name carries a leading dollar sign, which the IDENTIFIER pattern of the GoLR grammar does not allow, so no terminal a
// scanner section declares can collide with it.
//
// The symbol is added on first use rather than seeded up front, so that a grammar which does not ask for error recovery
// does not carry it. It gets no scanner rule: no input can produce it, and the generated scanner declares the constant
// which names it among its reserved tokens rather than deriving it from a rule. Adding a terminal without a rule is
// safe here because the two lists only have to line up while the scanner section is read, which has finished by the
// time a production can reference the symbol.
func (w *TreeWalker) internErrorTerminal() string {
	name := parsergenfrontend.SymbolError.Name
	if _, ok := w.terminalIdxByName[name]; ok {
		return name
	}

	w.grammar.Terminals = append(w.grammar.Terminals, parsergenfrontend.SymbolError)
	w.terminalIdxByName[name] = len(w.grammar.Terminals) - 1
	return name
}

func (w *TreeWalker) visitFile(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalFile {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalScannerSection:
			if err := w.visitScannerSection(&child); err != nil {
				return err
			}
		case parser.NonterminalParserSection:
			if err := w.visitParserSection(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitScannerSection(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerSection {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalScannerDeclList:
			if err := w.visitScannerDeclList(&child); err != nil {
				return err
			}
		}
	}
	return w.resolvePatterns()
}

func (w *TreeWalker) resolvePatterns() error {
	// The loop below reaches a scanner rule and its terminal with the same index, which makes this the one place where
	// the two lists have to line up. They do while the scanner section is read, because every declaration appends to
	// both. Afterwards they are allowed to drift apart: the error symbol is added as a terminal when a production first
	// references it and gets no scanner rule, because no input can produce it.
	utils.DebugAssert(func() error {
		if len(w.rules) != len(w.grammar.Terminals) {
			return fmt.Errorf(
				"scanner rules and terminals are out of sync: %d rules, %d terminals",
				len(w.rules),
				len(w.grammar.Terminals),
			)
		}
		return nil
	})

	// Fragments are lexemes which are not terminals
	fragments := make(map[string][]byte)
	for name, lexeme := range w.lexemeByName {
		if _, ok := w.terminalIdxByName[name]; !ok {
			fragments[name] = lexeme
		}
	}

	for idx, rule := range w.rules {
		lexeme, ok := w.lexemeByName[rule.Name]
		if !ok || len(lexeme) == 0 {
			continue
		}
		if lexeme[0] == '"' {
			alias := string(lexeme)
			if _, exists := w.terminalIdxByAlias[alias]; exists {
				return w.errorAt(w.lexemeByteOffsetByName[rule.Name], "alias %s has already been declared", alias)
			}
			w.grammar.Terminals[idx].Alias = alias
			w.terminalIdxByAlias[alias] = idx
			w.rules[idx].Regex = *dsl.Literal(alias[1 : len(alias)-1])
		} else {
			regexNode, err := regex.Parse(lexeme, fragments)
			if err != nil {
				return w.errorAt(
					w.lexemeByteOffsetByName[rule.Name],
					"invalid regex for terminal %q: %w",
					rule.Name,
					err,
				)
			}
			w.rules[idx].Regex = *regexNode
		}
	}
	return nil
}

func (w *TreeWalker) visitScannerDeclList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerDeclList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalScannerDeclList:
			if err := w.visitScannerDeclList(&child); err != nil {
				return err
			}
		case parser.NonterminalScannerDecl:
			if err := w.visitScannerDecl(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitScannerDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerDecl {
		panic("unexpected nonterminal")
	}

	// We reset the fragment for each declaration.
	w.isFragment = false

	nameNode, err := w.getNameNode(node)
	if err != nil {
		return err
	}
	name := string(w.text(nameNode))

	if _, ok := w.lexemeByName[name]; ok {
		return w.errorAt(nameNode.ByteOffset, "terminal %q is declared multiple times", name)
	}
	w.lexemeByName[name] = nil // prime the lexeme now, so we cannot forget empty declarations

	w.grammar.Terminals = append(w.grammar.Terminals, parsergenfrontend.Symbol{Name: name})
	w.terminalIdxByName[name] = len(w.grammar.Terminals) - 1
	w.rules = append(w.rules, scannergenfrontend.Rule{Name: name})

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalScannerDeclRhs:
			if err := w.visitScannerDeclRhs(&child); err != nil {
				return err
			}
		}
	}

	// In case @fragment was seen among the annotations, we drop the token from the rules list.
	if w.isFragment {
		w.rules = w.rules[:len(w.rules)-1]
		w.grammar.Terminals = w.grammar.Terminals[:len(w.grammar.Terminals)-1]
		delete(w.terminalIdxByName, name)
	}
	return nil
}

func (w *TreeWalker) visitScannerDeclRhs(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerDeclRhs {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		if terminal, ok := child.Symbol.Terminal(); ok && terminal == parser.TokenEmpty {
			w.rules[len(w.rules)-1].Regex = *dsl.CharClass()
			continue
		}

		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalScannerPattern:
			if err := w.visitScannerPattern(&child); err != nil {
				return err
			}
		case parser.NonterminalScannerAnnotationList:
			if err := w.visitScannerAnnotationList(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitScannerPattern(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerPattern {
		panic("unexpected nonterminal")
	}

	if len(node.Children) != 1 {
		return nil
	}

	if _, ok := node.Children[0].Symbol.Terminal(); !ok {
		return nil
	}

	name := w.rules[len(w.rules)-1].Name
	w.lexemeByName[name] = w.text(&node.Children[0])
	w.lexemeByteOffsetByName[name] = node.Children[0].ByteOffset
	return nil
}

func (w *TreeWalker) visitScannerAnnotationList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerAnnotationList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalScannerAnnotationList:
			if err := w.visitScannerAnnotationList(&child); err != nil {
				return err
			}
		case parser.NonterminalScannerAnnotation:
			if err := w.visitScannerAnnotation(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitScannerAnnotation(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerAnnotation {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		switch terminal {
		case parser.TokenSkip:
			w.rules[len(w.rules)-1].Skip = true
		case parser.TokenFragment:
			w.isFragment = true
		}
	}
	return nil
}

func (w *TreeWalker) visitParserSection(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalParserSection {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalStartDecl:
			w.visitStartDecl(&child)
		case parser.NonterminalPrecedenceSection:
			if err := w.visitPrecedenceSection(&child); err != nil {
				return err
			}
		case parser.NonterminalRuleDeclList:
			if err := w.visitRuleDeclList(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitStartDecl(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalStartDecl {
		panic("unexpected nonterminal")
	}

	// start_decl : %empty | "@start" ":" IDENTIFIER ";"
	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		if terminal == parser.TokenIdentifier && w.startNonterminalName == "" {
			w.startNonterminalName = string(w.text(&child))
			w.startNonterminalByteOffset = child.ByteOffset
		}
	}
}

func (w *TreeWalker) visitPrecedenceSection(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalPrecedenceSection {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalPrecedenceDeclList:
			if err := w.visitPrecedenceDeclList(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitPrecedenceDeclList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalPrecedenceDeclList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalPrecedenceDeclList:
			if err := w.visitPrecedenceDeclList(&child); err != nil {
				return err
			}
		case parser.NonterminalPrecedenceDecl:
			if err := w.visitPrecedenceDecl(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitPrecedenceDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalPrecedenceDecl {
		panic("unexpected nonterminal")
	}

	w.currentPrecedence--

	inPrecedenceDeclBackup := w.inPrecedenceDecl
	w.inPrecedenceDecl = true

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalAssociativity:
			w.visitAssociativity(&child)
		case parser.NonterminalSymbolList:
			if err := w.visitSymbolList(&child); err != nil {
				return err
			}
		}
	}

	w.inPrecedenceDecl = inPrecedenceDeclBackup
	return nil
}

func (w *TreeWalker) visitAssociativity(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAssociativity {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		switch terminal {
		case parser.TokenLeft:
			w.currentAssociativity = parsergenfrontend.AssociativityLeft
		case parser.TokenRight:
			w.currentAssociativity = parsergenfrontend.AssociativityRight
		case parser.TokenNone:
			w.currentAssociativity = parsergenfrontend.AssociativityNone
		case parser.TokenPrecedence:
			w.currentAssociativity = parsergenfrontend.AssociativityUndeclared
		}
	}
}

func (w *TreeWalker) visitRuleDeclList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalRuleDeclList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalRuleDeclList:
			if err := w.visitRuleDeclList(&child); err != nil {
				return err
			}
		case parser.NonterminalProductionDecl:
			if err := w.visitProductionDecl(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitProductionDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalProductionDecl {
		panic("unexpected nonterminal")
	}

	nameNode, err := w.getNameNode(node)
	if err != nil {
		return err
	}
	name := string(w.text(nameNode))

	if _, ok := w.terminalIdxByName[name]; ok {
		return w.errorAt(nameNode.ByteOffset, "left hand side of production %q is already declared as terminal", name)
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, parsergenfrontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
		w.nonterminalByteOffsetByName[name] = nameNode.ByteOffset
	}

	w.grammar.Productions = append(w.grammar.Productions, parsergenfrontend.Production{
		NonterminalIdx: w.nonterminalIdxByName[name],
	})

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalAlternativeList:
			if err := w.visitAlternativeList(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitAlternativeList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeList {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 3 {
		// alternative_list "|" alternative — recurse left, then create a new production for the right alternative.
		if err := w.visitAlternativeList(&node.Children[0]); err != nil {
			return err
		}
		w.grammar.Productions = append(w.grammar.Productions, parsergenfrontend.Production{
			NonterminalIdx: w.grammar.Productions[len(w.grammar.Productions)-1].NonterminalIdx,
		})
		if err := w.visitAlternative(&node.Children[2]); err != nil {
			return err
		}
		return nil
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalAlternative:
			if err := w.visitAlternative(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitAlternative(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternative {
		panic("unexpected nonterminal")
	}

	// alternative : symbol_list alternative_annotation_list | "@empty" alternative_annotation_list
	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalSymbolList:
			if err := w.visitSymbolList(&child); err != nil {
				return err
			}
		case parser.NonterminalAlternativeAnnotationList:
			if err := w.visitAlternativeAnnotationList(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitAlternativeAnnotationList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeAnnotationList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalAlternativeAnnotationList:
			if err := w.visitAlternativeAnnotationList(&child); err != nil {
				return err
			}
		case parser.NonterminalAlternativeAnnotation:
			if err := w.visitAlternativeAnnotation(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitAlternativeAnnotation(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeAnnotation {
		panic("unexpected nonterminal")
	}

	// alternative_annotation : "@precedence" "(" symbol ")"
	//                        | "@name" "(" IDENTIFIER ")"
	if w.isNameAnnotation(node) {
		return w.visitNameAnnotation(node)
	}

	inAlternativeAnnotationBackup := w.inAlternativeAnnotation
	w.inAlternativeAnnotation = true

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalSymbol:
			if err := w.visitSymbol(&child); err != nil {
				return err
			}
		}
	}

	w.inAlternativeAnnotation = inAlternativeAnnotationBackup
	return nil
}

// isNameAnnotation reports whether the alternative_annotation node is an @name annotation rather than an @precedence
// one, by looking for the "@name" keyword among its children.
func (w *TreeWalker) isNameAnnotation(node *parser.Node) bool {
	for _, child := range node.Children {
		if terminal, ok := child.Symbol.Terminal(); ok && terminal == parser.TokenName {
			return true
		}
	}
	return false
}

// visitNameAnnotation sets the explicit name of the current production from an "@name" "(" IDENTIFIER ")" annotation.
// Two productions must not ask for the same name.
func (w *TreeWalker) visitNameAnnotation(node *parser.Node) error {
	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok || terminal != parser.TokenIdentifier {
			continue
		}

		name := string(w.text(&child))
		if _, exists := w.explicitProductionNames[name]; exists {
			return w.errorAt(child.ByteOffset, "duplicate production name %q", name)
		}
		w.explicitProductionNames[name] = struct{}{}
		w.grammar.Productions[len(w.grammar.Productions)-1].Name = &name
		return nil
	}
	return w.errorAt(node.ByteOffset, "no IDENTIFIER token found in @name annotation")
}

func (w *TreeWalker) visitSymbolList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbolList {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalSymbolList:
			if err := w.visitSymbolList(&child); err != nil {
				return err
			}
		case parser.NonterminalSymbol:
			if err := w.visitSymbol(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *TreeWalker) visitSymbol(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbol {
		panic("unexpected nonterminal")
	}

	name, err := w.getSymbolName(node)
	if err != nil {
		return err
	}

	if w.inAlternativeAnnotation {
		return w.visitSymbolInAlternativeAnnotation(node, name)
	}

	if w.inPrecedenceDecl {
		return w.visitSymbolInPrecedenceDecl(node, name)
	}

	return w.visitSymbolInAlternative(node, name)
}

func (w *TreeWalker) visitSymbolInAlternativeAnnotation(node *parser.Node, name string) error {
	// We are inside @precedence(...) — set the precedence override terminal for the current production.
	if idx, ok := w.terminalIdxByName[name]; ok {
		w.grammar.Productions[len(w.grammar.Productions)-1].PrecedenceTerminalIdx = &idx
		return nil
	}
	if idx, ok := w.terminalIdxByAlias[name]; ok {
		w.grammar.Productions[len(w.grammar.Productions)-1].PrecedenceTerminalIdx = &idx
		return nil
	}
	return w.errorAt(node.ByteOffset, "undeclared terminal %s", name)
}

func (w *TreeWalker) visitSymbolInPrecedenceDecl(node *parser.Node, name string) error {
	// We are inside a precedence_decl symbol_list — assign precedence and associativity to this terminal.
	if idx, ok := w.terminalIdxByName[name]; ok {
		w.grammar.Terminals[idx].Associativity = w.currentAssociativity
		w.grammar.Terminals[idx].Precedence = w.currentPrecedence
		return nil
	}
	if idx, ok := w.terminalIdxByAlias[name]; ok {
		w.grammar.Terminals[idx].Associativity = w.currentAssociativity
		w.grammar.Terminals[idx].Precedence = w.currentPrecedence
		return nil
	}
	return w.errorAt(node.ByteOffset, "undeclared terminal %s", name)
}

func (w *TreeWalker) visitSymbolInAlternative(node *parser.Node, name string) error {
	// We are in a production alternative — add the symbol to the current production's RHS.
	if terminalIdx, ok := w.terminalIdxByName[name]; ok {
		production := w.grammar.Productions[len(w.grammar.Productions)-1]
		production.SymbolRefs = append(production.SymbolRefs, parsergenfrontend.NewTerminalRef(terminalIdx))
		w.grammar.Productions[len(w.grammar.Productions)-1] = production
		return nil
	}

	if terminalIdx, ok := w.terminalIdxByAlias[name]; ok {
		production := w.grammar.Productions[len(w.grammar.Productions)-1]
		production.SymbolRefs = append(production.SymbolRefs, parsergenfrontend.NewTerminalRef(terminalIdx))
		w.grammar.Productions[len(w.grammar.Productions)-1] = production
		return nil
	}

	if len(name) > 0 && name[0] == '"' {
		return w.errorAt(node.ByteOffset, "undeclared terminal %s", name)
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, parsergenfrontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
		w.nonterminalByteOffsetByName[name] = node.ByteOffset
	}

	nonterminalIdx := w.nonterminalIdxByName[name]
	production := w.grammar.Productions[len(w.grammar.Productions)-1]
	production.SymbolRefs = append(production.SymbolRefs, parsergenfrontend.NewNonterminalRef(nonterminalIdx))
	w.grammar.Productions[len(w.grammar.Productions)-1] = production
	return nil
}

// getNameNode returns the IDENTIFIER child of the given node, which is the name the node declares.
func (w *TreeWalker) getNameNode(node *parser.Node) (*parser.Node, error) {
	for i := range node.Children {
		terminal, ok := node.Children[i].Symbol.Terminal()
		if !ok {
			continue
		}
		if terminal == parser.TokenIdentifier {
			return &node.Children[i], nil
		}
	}
	return nil, w.errorAt(node.ByteOffset, "no name token found")
}

func (w *TreeWalker) getSymbolName(node *parser.Node) (string, error) {
	if len(node.Children) != 1 {
		return "", w.errorAt(node.ByteOffset, "unexpected symbol node structure")
	}
	child := node.Children[0]
	terminal, ok := child.Symbol.Terminal()
	if !ok {
		return "", w.errorAt(node.ByteOffset, "expected terminal in symbol node")
	}
	switch terminal {
	case parser.TokenIdentifier:
		return string(w.text(&child)), nil
	case parser.TokenError:
		return w.internErrorTerminal(), nil
	case parser.TokenString:
		// Strings in symbol position reference a terminal by its alias.
		alias := string(w.text(&child))
		if idx, ok := w.terminalIdxByAlias[alias]; ok {
			return w.grammar.Terminals[idx].Name, nil
		}
		return alias, nil
	}
	return "", w.errorAt(node.ByteOffset, "unexpected token %v in symbol node", terminal)
}
