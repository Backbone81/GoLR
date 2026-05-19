//nolint:exhaustive // The ASTWalker will only descend selected nonterminals and does not need to be exhaustive.
package golr

import (
	"errors"
	"fmt"
	"math"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// ASTWalker is a helper struct which walks the abstract syntax tree of a parsed GoLR grammar and extracts all
// information required to describe the context free grammar therein.
type ASTWalker struct {
	grammar frontend.Grammar

	terminalIdxByName    map[string]int
	terminalIdxByAlias   map[string]int
	nonterminalIdxByName map[string]int

	// startNonterminalName keeps track of the start symbol declared with @start.
	startNonterminalName string

	// currentPrecedence is the current precedence level. The first precedence declared has the highest priority. Every
	// following precedence is decremented by one for every @left, @right and @none
	// declaration.
	currentPrecedence int

	// currentAssociativity keeps track of the associativity being declared in the current precedence_decl.
	currentAssociativity frontend.Associativity

	// inPrecedenceDecl keeps track if we are currently inside a precedence_decl. Symbols visited via symbol_list
	// inside a precedence_decl are terminals being assigned the current precedence level and associativity.
	inPrecedenceDecl bool

	// inAlternativeAnnotation keeps track if we are currently inside an alternative_annotation. The symbol visited
	// inside an alternative_annotation is the terminal used to override the precedence for the current production.
	inAlternativeAnnotation bool
}

// NewASTWalker creates a new ASTWalker.
func NewASTWalker() *ASTWalker {
	return &ASTWalker{
		terminalIdxByName:    make(map[string]int),
		terminalIdxByAlias:   make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
		currentPrecedence:    math.MaxInt,
	}
}

// BuildGrammar takes the root node of the abstract syntax tree, traverses the tree to build the context free grammar
// and returns the finished grammar afterward.
func (w *ASTWalker) BuildGrammar(node *parser.Node) (frontend.Grammar, error) {
	if err := w.visitFile(node); err != nil {
		return frontend.Grammar{}, err
	}
	if w.startNonterminalName != "" {
		idx, ok := w.nonterminalIdxByName[w.startNonterminalName]
		if !ok {
			return frontend.Grammar{}, fmt.Errorf("unknown start nonterminal %q", w.startNonterminalName)
		}
		w.grammar.StartNonterminalIdx = idx
	}
	return w.grammar, nil
}

func (w *ASTWalker) visitFile(node *parser.Node) error {
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
			if err := w.visitScannerSection(child); err != nil {
				return err
			}
		case parser.NonterminalParserSection:
			if err := w.visitParserSection(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerSection(node *parser.Node) error {
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
			if err := w.visitScannerDeclList(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerDeclList(node *parser.Node) error {
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
			if err := w.visitScannerDeclList(child); err != nil {
				return err
			}
		case parser.NonterminalScannerDecl:
			if err := w.visitScannerDecl(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerDecl {
		panic("unexpected nonterminal")
	}

	name, err := w.getNameLexeme(node)
	if err != nil {
		return err
	}

	if _, ok := w.terminalIdxByName[name]; ok {
		return fmt.Errorf("terminal %q is declared multiple times", name)
	}

	w.grammar.Terminals = append(w.grammar.Terminals, frontend.Symbol{Name: name})
	w.terminalIdxByName[name] = len(w.grammar.Terminals) - 1

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalScannerDeclRhs:
			if err := w.visitScannerDeclRhs(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerDeclRhs(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerDeclRhs {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalScannerPattern:
			if err := w.visitScannerPattern(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerPattern(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerPattern {
		panic("unexpected nonterminal")
	}

	if len(node.Children) != 1 {
		return nil
	}

	// A STRING pattern doubles as the alias for the terminal, allowing productions to reference it by string.
	if terminal, ok := node.Children[0].Symbol.Terminal(); !ok || terminal != parser.TokenString {
		return nil
	}

	alias := string(node.Children[0].Lexeme)
	idx := len(w.grammar.Terminals) - 1
	w.grammar.Terminals[idx].Alias = alias

	if _, ok := w.terminalIdxByAlias[alias]; ok {
		return fmt.Errorf("alias %s has already been declared", alias)
	}
	w.terminalIdxByAlias[alias] = idx
	return nil
}

func (w *ASTWalker) visitParserSection(node *parser.Node) error {
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
			w.visitStartDecl(child)
		case parser.NonterminalPrecedenceSection:
			if err := w.visitPrecedenceSection(child); err != nil {
				return err
			}
		case parser.NonterminalRuleDeclList:
			if err := w.visitRuleDeclList(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitStartDecl(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalStartDecl {
		panic("unexpected nonterminal")
	}

	// start_decl : %empty | "@start" ":" NAME ";"
	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		if terminal == parser.TokenName && w.startNonterminalName == "" {
			w.startNonterminalName = string(child.Lexeme)
		}
	}
}

func (w *ASTWalker) visitPrecedenceSection(node *parser.Node) error {
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
			if err := w.visitPrecedenceDeclList(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitPrecedenceDeclList(node *parser.Node) error {
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
			if err := w.visitPrecedenceDeclList(child); err != nil {
				return err
			}
		case parser.NonterminalPrecedenceDecl:
			if err := w.visitPrecedenceDecl(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitPrecedenceDecl(node *parser.Node) error {
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
			w.visitAssociativity(child)
		case parser.NonterminalSymbolList:
			if err := w.visitSymbolList(child); err != nil {
				return err
			}
		}
	}

	w.inPrecedenceDecl = inPrecedenceDeclBackup
	return nil
}

func (w *ASTWalker) visitAssociativity(node *parser.Node) {
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
			w.currentAssociativity = frontend.AssociativityLeft
		case parser.TokenRight:
			w.currentAssociativity = frontend.AssociativityRight
		case parser.TokenNone:
			w.currentAssociativity = frontend.AssociativityNone
		}
	}
}

func (w *ASTWalker) visitRuleDeclList(node *parser.Node) error {
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
			w.visitRuleDeclList(child)
		case parser.NonterminalProductionDecl:
			w.visitProductionDecl(child)
		}
	}
}

func (w *ASTWalker) visitProductionDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalProductionDecl {
		panic("unexpected nonterminal")
	}

	name, err := w.getNameLexeme(node)
	if err != nil {
		return
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, frontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
	}

	w.grammar.Productions = append(w.grammar.Productions, frontend.Production{
		NonterminalIdx: w.nonterminalIdxByName[name],
	})

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalAlternativeList:
			w.visitAlternativeList(child)
		}
	}
}

func (w *ASTWalker) visitAlternativeList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeList {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 3 {
		// alternative_list "|" alternative — recurse left, then create a new production for the right alternative.
		w.visitAlternativeList(node.Children[0])
		w.grammar.Productions = append(w.grammar.Productions, frontend.Production{
			NonterminalIdx: w.grammar.Productions[len(w.grammar.Productions)-1].NonterminalIdx,
		})
		w.visitAlternative(node.Children[2])
		return
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalAlternative:
			w.visitAlternative(child)
		}
	}
}

func (w *ASTWalker) visitAlternative(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternative {
		panic("unexpected nonterminal")
	}

	// alternative : symbol_list alternative_annotation_list | "@empty"
	// The "@empty" alternative leaves the current production with an empty RHS — nothing to do in that case.
	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalSymbolList:
			w.visitSymbolList(child)
		case parser.NonterminalAlternativeAnnotationList:
			w.visitAlternativeAnnotationList(child)
		}
	}
}

func (w *ASTWalker) visitAlternativeAnnotationList(node *parser.Node) error {
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
			w.visitAlternativeAnnotationList(child)
		case parser.NonterminalAlternativeAnnotation:
			w.visitAlternativeAnnotation(child)
		}
	}
}

func (w *ASTWalker) visitAlternativeAnnotation(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeAnnotation {
		panic("unexpected nonterminal")
	}

	// alternative_annotation : "@precedence" "(" symbol ")"
	inAlternativeAnnotationBackup := w.inAlternativeAnnotation
	w.inAlternativeAnnotation = true

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal { //nolint:gocritic // We keep the switch for ease of extension and uniformity.
		case parser.NonterminalSymbol:
			w.visitSymbol(child)
		}
	}

	w.inAlternativeAnnotation = inAlternativeAnnotationBackup
}

func (w *ASTWalker) visitSymbolList(node *parser.Node) error {
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
			w.visitSymbolList(child)
		case parser.NonterminalSymbol:
			w.visitSymbol(child)
		}
	}
}

func (w *ASTWalker) visitSymbol(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbol {
		panic("unexpected nonterminal")
	}

	name, err := w.getSymbolName(node)
	if err != nil {
		return
	}

	if w.inAlternativeAnnotation {
		// We are inside @precedence(...) — set the precedence override terminal for the current production.
		if idx, ok := w.terminalIdxByName[name]; ok {
			production := w.grammar.Productions[len(w.grammar.Productions)-1]
			production.PrecedenceTerminalIdx = &idx
			w.grammar.Productions[len(w.grammar.Productions)-1] = production
		}
		return
	}

	if w.inPrecedenceDecl {
		// We are inside a precedence_decl symbol_list — assign precedence and associativity to this terminal.
		if idx, ok := w.terminalIdxByName[name]; ok {
			w.grammar.Terminals[idx].Associativity = w.currentAssociativity
			w.grammar.Terminals[idx].Precedence = w.currentPrecedence
		}
		return
	}

	// We are in a production alternative — add the symbol to the current production's RHS.
	if terminalIdx, ok := w.terminalIdxByName[name]; ok {
		production := w.grammar.Productions[len(w.grammar.Productions)-1]
		production.SymbolRefs = append(production.SymbolRefs, frontend.NewTerminalRef(terminalIdx))
		w.grammar.Productions[len(w.grammar.Productions)-1] = production
		return
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, frontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
	}

	nonterminalIdx := w.nonterminalIdxByName[name]
	production := w.grammar.Productions[len(w.grammar.Productions)-1]
	production.SymbolRefs = append(production.SymbolRefs, frontend.NewNonterminalRef(nonterminalIdx))
	w.grammar.Productions[len(w.grammar.Productions)-1] = production
}

func (w *ASTWalker) getNameLexeme(node *parser.Node) (string, error) {
	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		if terminal == parser.TokenName {
			return string(child.Lexeme), nil
		}
	}
	return "", errors.New("no name token found")
}

func (w *ASTWalker) getSymbolName(node *parser.Node) (string, error) {
	if len(node.Children) != 1 {
		return "", errors.New("unexpected symbol node structure")
	}
	child := node.Children[0]
	terminal, ok := child.Symbol.Terminal()
	if !ok {
		return "", errors.New("expected terminal in symbol node")
	}
	switch terminal {
	case parser.TokenName:
		return string(child.Lexeme), nil
	case parser.TokenString:
		// Strings in symbol position reference a terminal by its alias.
		alias := string(child.Lexeme)
		if idx, ok := w.terminalIdxByAlias[alias]; ok {
			return w.grammar.Terminals[idx].Name, nil
		}
		return alias, nil
	}
	return "", fmt.Errorf("unexpected token %v in symbol node", terminal)
}
