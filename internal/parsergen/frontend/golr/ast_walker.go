//nolint:exhaustive // The ASTWalker will only descend selected nonterminals and does not need to be exhaustive.
package golr

import (
	"errors"
	"fmt"
	"math"

	parsergenfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr/regex"
	scannergenfrontend "github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/pkg/scannergen/frontend/dsl"
)

// ASTWalker is a helper struct which walks the abstract syntax tree of a parsed GoLR grammar and extracts all
// information required to describe the context free grammar therein.
type ASTWalker struct {
	rules   []scannergenfrontend.Rule
	grammar parsergenfrontend.Grammar

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
	currentAssociativity parsergenfrontend.Associativity

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
func (w *ASTWalker) BuildGrammar(node *parser.Node) ([]scannergenfrontend.Rule, parsergenfrontend.Grammar, error) {
	if err := w.visitFile(node); err != nil {
		return nil, parsergenfrontend.Grammar{}, err
	}
	if w.startNonterminalName != "" {
		idx, ok := w.nonterminalIdxByName[w.startNonterminalName]
		if !ok {
			return nil, parsergenfrontend.Grammar{}, fmt.Errorf("unknown start nonterminal %q", w.startNonterminalName)
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
				fmt.Errorf("nonterminal %q is referenced but never defined", nonterminal.Name)
		}
	}

	if len(w.grammar.Productions) < 1 {
		return nil, parsergenfrontend.Grammar{}, errors.New("grammar requires at least one production")
	}
	return w.rules, w.grammar, nil
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
			if err := w.visitScannerPattern(child); err != nil {
				return err
			}
		case parser.NonterminalScannerAnnotationList:
			if err := w.visitScannerAnnotationList(child); err != nil {
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

	terminal, ok := node.Children[0].Symbol.Terminal()
	if !ok {
		return nil
	}

	switch terminal {
	case parser.TokenRegex:
		regexNode, err := regex.Parse(node.Children[0].Lexeme)
		if err != nil {
			return fmt.Errorf("invalid regex for terminal: %q: %w", w.grammar.Terminals[len(w.grammar.Terminals)-1].Name, err)
		}
		w.rules[len(w.rules)-1].Regex = *regexNode
	case parser.TokenString:
		alias := string(node.Children[0].Lexeme)
		idx := len(w.grammar.Terminals) - 1
		w.grammar.Terminals[idx].Alias = alias

		if _, ok := w.terminalIdxByAlias[alias]; ok {
			return fmt.Errorf("alias %s has already been declared", alias)
		}
		w.terminalIdxByAlias[alias] = idx

		// We need to strip the quotes from the alias when we create the literal.
		w.rules[len(w.rules)-1].Regex = *dsl.Literal(alias[1 : len(alias)-1])
	default:
		return nil
	}
	return nil
}

func (w *ASTWalker) visitScannerAnnotationList(node *parser.Node) error {
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
			if err := w.visitScannerAnnotationList(child); err != nil {
				return err
			}
		case parser.NonterminalScannerAnnotation:
			if err := w.visitScannerAnnotation(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitScannerAnnotation(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalScannerAnnotation {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		terminal, ok := child.Symbol.Terminal()
		if !ok {
			continue
		}
		if terminal == parser.TokenSkip {
			w.rules[len(w.rules)-1].Skip = true
		}
	}
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
			if err := w.visitRuleDeclList(child); err != nil {
				return err
			}
		case parser.NonterminalProductionDecl:
			if err := w.visitProductionDecl(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitProductionDecl(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalProductionDecl {
		panic("unexpected nonterminal")
	}

	name, err := w.getNameLexeme(node)
	if err != nil {
		return err
	}

	if _, ok := w.terminalIdxByName[name]; ok {
		return fmt.Errorf("left hand side of production %q is already declared as terminal", name)
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, parsergenfrontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
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
			if err := w.visitAlternativeList(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitAlternativeList(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlternativeList {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 3 {
		// alternative_list "|" alternative — recurse left, then create a new production for the right alternative.
		if err := w.visitAlternativeList(node.Children[0]); err != nil {
			return err
		}
		w.grammar.Productions = append(w.grammar.Productions, parsergenfrontend.Production{
			NonterminalIdx: w.grammar.Productions[len(w.grammar.Productions)-1].NonterminalIdx,
		})
		if err := w.visitAlternative(node.Children[2]); err != nil {
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
			if err := w.visitAlternative(child); err != nil {
				return err
			}
		}
	}
	return nil
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
			if err := w.visitSymbolList(child); err != nil {
				return err
			}
		case parser.NonterminalAlternativeAnnotationList:
			if err := w.visitAlternativeAnnotationList(child); err != nil {
				return err
			}
		}
	}
	return nil
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
			if err := w.visitAlternativeAnnotationList(child); err != nil {
				return err
			}
		case parser.NonterminalAlternativeAnnotation:
			if err := w.visitAlternativeAnnotation(child); err != nil {
				return err
			}
		}
	}
	return nil
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
			if err := w.visitSymbol(child); err != nil {
				return err
			}
		}
	}

	w.inAlternativeAnnotation = inAlternativeAnnotationBackup
	return nil
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
			if err := w.visitSymbolList(child); err != nil {
				return err
			}
		case parser.NonterminalSymbol:
			if err := w.visitSymbol(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *ASTWalker) visitSymbol(node *parser.Node) error {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbol {
		panic("unexpected nonterminal")
	}

	name, err := w.getSymbolName(node)
	if err != nil {
		return err
	}

	if w.inAlternativeAnnotation {
		return w.visitSymbolInAlternativeAnnotation(name)
	}

	if w.inPrecedenceDecl {
		return w.visitSymbolInPrecedenceDecl(name)
	}

	return w.visitSymbolInAlternative(name)
}

func (w *ASTWalker) visitSymbolInAlternativeAnnotation(name string) error {
	// We are inside @precedence(...) — set the precedence override terminal for the current production.
	if idx, ok := w.terminalIdxByName[name]; ok {
		w.grammar.Productions[len(w.grammar.Productions)-1].PrecedenceTerminalIdx = &idx
		return nil
	}
	if idx, ok := w.terminalIdxByAlias[name]; ok {
		w.grammar.Productions[len(w.grammar.Productions)-1].PrecedenceTerminalIdx = &idx
		return nil
	}
	return fmt.Errorf("undeclared terminal %s", name)
}

func (w *ASTWalker) visitSymbolInPrecedenceDecl(name string) error {
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
	return fmt.Errorf("undeclared terminal %s", name)
}

func (w *ASTWalker) visitSymbolInAlternative(name string) error {
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
		return fmt.Errorf("undeclared terminal %s", name)
	}

	if _, ok := w.nonterminalIdxByName[name]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, parsergenfrontend.Symbol{Name: name})
		w.nonterminalIdxByName[name] = len(w.grammar.Nonterminals) - 1
	}

	nonterminalIdx := w.nonterminalIdxByName[name]
	production := w.grammar.Productions[len(w.grammar.Productions)-1]
	production.SymbolRefs = append(production.SymbolRefs, parsergenfrontend.NewNonterminalRef(nonterminalIdx))
	w.grammar.Productions[len(w.grammar.Productions)-1] = production
	return nil
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
