package bison

import (
	"errors"
	"fmt"
	"golr/internal/parsergen/frontend"
	"golr/internal/parsergen/frontend/bison/parser"
)

type ASTWalker struct {
	grammar frontend.Grammar

	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int

	activePercentStart   bool
	startNonterminalName string
}

func NewASTWalker() *ASTWalker {
	return &ASTWalker{
		terminalIdxByName:    make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
	}
}

func (w *ASTWalker) BuildGrammar(node *parser.Node) (frontend.Grammar, error) {
	w.visitInput(node)
	if w.startNonterminalName != "" {
		idx, ok := w.nonterminalIdxByName[w.startNonterminalName]
		if !ok {
			return frontend.Grammar{}, fmt.Errorf("unknown start nonterminal %q", w.startNonterminalName)
		}
		w.grammar.StartNonterminalIdx = idx
	}
	return w.grammar, nil
}

func (w *ASTWalker) visitInput(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalInput {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalPrologueDeclarations:
			w.visitPrologueDeclarations(child)
		case parser.NonterminalGrammar:
			w.visitGrammar(child)
		}
	}
}

func (w *ASTWalker) visitPrologueDeclarations(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalPrologueDeclarations {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalPrologueDeclarations:
			w.visitPrologueDeclarations(child)
		case parser.NonterminalPrologueDeclaration:
			w.visitPrologueDeclaration(child)
		}
	}
}

func (w *ASTWalker) visitPrologueDeclaration(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalPrologueDeclaration {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalGrammarDeclaration:
			w.visitGrammarDeclaration(child)
		}
	}
}

func (w *ASTWalker) visitGrammarDeclaration(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalGrammarDeclaration {
		panic("unexpected nonterminal")
	}

	activePercentStartBackup := w.activePercentStart
	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if ok {
			switch nonterminal {
			case parser.NonterminalSymbolDeclaration:
				w.visitSymbolDeclaration(child)
			case parser.NonterminalSymbols_1:
				w.visitSymbols_1(child)
			}
			continue
		}

		terminal, ok := child.Symbol.Terminal()
		if ok {
			switch terminal {
			case parser.TokenPercentStart:
				w.activePercentStart = true
			}
			continue
		}
	}
	w.activePercentStart = activePercentStartBackup
}

func (w *ASTWalker) visitSymbols_1(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbols_1 {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalSymbol:
			w.visitSymbol(child)
		case parser.NonterminalSymbols_1:
			w.visitSymbols_1(child)
		}
	}
}

func (w *ASTWalker) visitSymbolDeclaration(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalSymbolDeclaration {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalTokenDecls:
			w.visitTokenDecls(child)
		case parser.NonterminalTokenDeclsForPrec:
			w.visitTokenDeclsForPrec(child)
		}
	}
}

func (w *ASTWalker) visitTokenDecls(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDecls {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalTokenDecl_1:
			w.visitTokenDecl_1(child)
		}
	}
}

func (w *ASTWalker) visitTokenDecl_1(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDecl_1 {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalTokenDecl_1:
			w.visitTokenDecl_1(child)
		case parser.NonterminalTokenDecl:
			w.visitTokenDecl(child)
		}
	}
}

func (w *ASTWalker) visitTokenDecl(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDecl {
		panic("unexpected nonterminal")
	}

	// id int.opt[num] alias
	id, err := w.getID(node)
	if err != nil {
		return
	}

	if _, ok := w.terminalIdxByName[id]; ok {
		// We have a duplicate. Ignore it.
		return
	}

	w.grammar.Terminals = append(w.grammar.Terminals, frontend.Symbol{
		Name: id,
	})
	w.terminalIdxByName[id] = len(w.grammar.Terminals) - 1

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalAlias:
			w.visitAlias(child)
		}
	}
}

func (w *ASTWalker) visitAlias(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalAlias {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 1 {
		if terminal, ok := node.Children[0].Symbol.Terminal(); ok && terminal == parser.TokenTstring {
			w.grammar.Terminals[len(w.grammar.Terminals)-1].Alias = string(node.Children[0].Lexeme)
			w.terminalIdxByName[string(node.Children[0].Lexeme)] = len(w.grammar.Terminals) - 1
		}
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalStringAsId:
			w.visitStringAsId(child)
		}
	}
}

func (w *ASTWalker) visitStringAsId(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalStringAsId {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 1 {
		if terminal, ok := node.Children[0].Symbol.Terminal(); ok && terminal == parser.TokenString {
			w.grammar.Terminals[len(w.grammar.Terminals)-1].Alias = string(node.Children[0].Lexeme)
			w.terminalIdxByName[string(node.Children[0].Lexeme)] = len(w.grammar.Terminals) - 1
		}
	}
}

func (w *ASTWalker) visitTokenDeclsForPrec(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDeclsForPrec {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalTokenDeclsForPrec:
			w.visitTokenDeclsForPrec(child)
		case parser.NonterminalTokenDeclForPrec_1:
			w.visitTokenDeclForPrec_1(child)
		}
	}
}

func (w *ASTWalker) visitTokenDeclForPrec_1(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDeclForPrec_1 {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalTokenDeclForPrec_1:
			w.visitTokenDeclForPrec_1(child)
		case parser.NonterminalTokenDeclForPrec:
			w.visitTokenDeclForPrec(child)
		}
	}
}

func (w *ASTWalker) visitTokenDeclForPrec(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalTokenDeclForPrec {
		panic("unexpected nonterminal")
	}

	id, err := w.getID(node)
	if err != nil {
		id, err = w.getStringAsID(node)
		if err != nil {
			id, err = w.getCharLiteralAsID(node)
			if err != nil {
				return
			}
		}
	}

	if _, ok := w.terminalIdxByName[id]; !ok {
		if id[0] == '\'' {
			w.grammar.Terminals = append(w.grammar.Terminals, frontend.Symbol{
				Name:  fmt.Sprintf("CHAR_%d", int(id[1])),
				Alias: id,
			})
		} else {
			w.grammar.Terminals = append(w.grammar.Terminals, frontend.Symbol{
				Name: id,
			})
		}
		w.terminalIdxByName[id] = len(w.grammar.Terminals) - 1
	}
}

func (w *ASTWalker) visitGrammar(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalGrammar {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalGrammar:
			w.visitGrammar(child)
		case parser.NonterminalRulesOrGrammarDeclaration:
			w.visitRulesOrGrammarDeclaration(child)
		}
	}
}

func (w *ASTWalker) visitRulesOrGrammarDeclaration(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalRulesOrGrammarDeclaration {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalRules:
			w.visitRules(child)
		case parser.NonterminalGrammarDeclaration:
			w.visitGrammarDeclaration(child)
		}
	}
}

func (w *ASTWalker) visitRules(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalRules {
		panic("unexpected nonterminal")
	}

	// id_colon named_ref.opt COLON rhses.1
	idColon, err := w.getIDColon(node)
	if err != nil {
		return
	}

	if _, ok := w.nonterminalIdxByName[idColon]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, frontend.Symbol{
			Name: idColon,
		})
		w.nonterminalIdxByName[idColon] = len(w.grammar.Nonterminals) - 1
	}
	w.grammar.Productions = append(w.grammar.Productions, frontend.Production{
		NonterminalIdx: w.nonterminalIdxByName[idColon],
	})

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalRhses_1:
			w.visitRhses_1(child)
		}
	}
}

func (w *ASTWalker) visitRhses_1(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalRhses_1 {
		panic("unexpected nonterminal")
	}

	if len(node.Children) == 3 {
		if terminal, ok := node.Children[1].Symbol.Terminal(); ok && terminal == parser.TokenPipe {
			// We create a new production with the same nonterminal on the left hand side.
			w.visitRhses_1(node.Children[0])
			w.grammar.Productions = append(w.grammar.Productions, frontend.Production{
				NonterminalIdx: w.grammar.Productions[len(w.grammar.Productions)-1].NonterminalIdx,
			})
			w.visitRhs(node.Children[2])
			return
		}
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			continue
		}
		switch nonterminal {
		case parser.NonterminalRhs:
			w.visitRhs(child)
		case parser.NonterminalRhses_1:
			w.visitRhses_1(child)
		}
	}
}

func (w *ASTWalker) visitRhs(node *parser.Node) {
	if nonterminal, ok := node.Symbol.Nonterminal(); !ok || nonterminal != parser.NonterminalRhs {
		panic("unexpected nonterminal")
	}

	for _, child := range node.Children {
		nonterminal, ok := child.Symbol.Nonterminal()
		if !ok {
			// When we encounter a terminal in the context of a Rhs, we stop because that is usually a situation
			// where "rhs PERCENT_PREC symbol" is encountered and the symbol is nothing we want to extract as
			// nonterminal.
			return
		}
		switch nonterminal {
		case parser.NonterminalSymbol:
			w.visitSymbol(child)
		case parser.NonterminalRhs:
			w.visitRhs(child)
		}
	}
}

func (w *ASTWalker) visitSymbol(node *parser.Node) {
	nonterminal, ok := node.Symbol.Nonterminal()
	if !ok || nonterminal != parser.NonterminalSymbol {
		panic("unexpected nonterminal")
	}

	var id string
	var err error
	id, err = w.getID(node)
	if err != nil {
		id, err = w.getStringAsID(node)
		if err != nil {
			id, err = w.getCharLiteralAsID(node)
			if err != nil {
				return
			}

			// char literals are always terminals but need not be pre-declared with %token
			if _, ok := w.terminalIdxByName[id]; !ok {
				w.grammar.Terminals = append(w.grammar.Terminals, frontend.Symbol{
					Name:  fmt.Sprintf("CHAR_%d", int(id[1])),
					Alias: id,
				})
				w.terminalIdxByName[id] = len(w.grammar.Terminals) - 1
			}
		}
	}

	if w.activePercentStart {
		// We arrived here through the %start declaration. We set the start nonterminal only if not already set.
		if w.startNonterminalName == "" {
			w.startNonterminalName = id
		}
		return
	}

	// We are here because we are reading a rule and need to fill the right hand side of the production.
	if terminalIdx, ok := w.terminalIdxByName[id]; ok {
		production := w.grammar.Productions[len(w.grammar.Productions)-1]
		production.SymbolRefs = append(production.SymbolRefs, frontend.NewTerminalRef(terminalIdx))
		w.grammar.Productions[len(w.grammar.Productions)-1] = production
		return
	}
	if _, ok := w.nonterminalIdxByName[id]; !ok {
		w.grammar.Nonterminals = append(w.grammar.Nonterminals, frontend.Symbol{
			Name: id,
		})
		w.nonterminalIdxByName[id] = len(w.grammar.Nonterminals) - 1
	}

	nonterminalIdx := w.nonterminalIdxByName[id]
	production := w.grammar.Productions[len(w.grammar.Productions)-1]
	production.SymbolRefs = append(production.SymbolRefs, frontend.NewNonterminalRef(nonterminalIdx))
	w.grammar.Productions[len(w.grammar.Productions)-1] = production
}

func (w *ASTWalker) getID(node *parser.Node) (string, error) {
	firstChild := node.Children[0]
	idNonterminal, ok := firstChild.Symbol.Nonterminal()
	if !ok {
		return "", errors.New("no nonterminal")
	}
	if idNonterminal != parser.NonterminalId {
		return "", errors.New("expected nonterminal id")
	}

	firstChildChild := firstChild.Children[0]
	idTerminal, ok := firstChildChild.Symbol.Terminal()
	if !ok {
		return "", errors.New("no token")
	}
	if idTerminal != parser.TokenId {
		return "", errors.New("expected token id")
	}
	return string(firstChildChild.Lexeme), nil
}

func (w *ASTWalker) getStringAsID(node *parser.Node) (string, error) {
	firstChild := node.Children[0]
	idNonterminal, ok := firstChild.Symbol.Nonterminal()
	if !ok {
		return "", errors.New("no nonterminal")
	}
	if idNonterminal != parser.NonterminalStringAsId {
		return "", errors.New("expected nonterminal string as id")
	}

	firstChildChild := firstChild.Children[0]
	idTerminal, ok := firstChildChild.Symbol.Terminal()
	if !ok {
		return "", errors.New("no token")
	}
	if idTerminal != parser.TokenString {
		return "", errors.New("expected token string")
	}
	return string(firstChildChild.Lexeme), nil
}

func (w *ASTWalker) getCharLiteralAsID(node *parser.Node) (string, error) {
	firstChild := node.Children[0]
	idNonterminal, ok := firstChild.Symbol.Nonterminal()
	if !ok {
		return "", errors.New("no nonterminal")
	}
	if idNonterminal != parser.NonterminalId {
		return "", errors.New("expected nonterminal string as id")
	}

	firstChildChild := firstChild.Children[0]
	idTerminal, ok := firstChildChild.Symbol.Terminal()
	if !ok {
		return "", errors.New("no token")
	}
	if idTerminal != parser.TokenCharLiteral {
		return "", errors.New("expected token char literal")
	}
	return string(firstChildChild.Lexeme), nil
}

func (w *ASTWalker) getIDColon(node *parser.Node) (string, error) {
	firstChild := node.Children[0]
	idNonterminal, ok := firstChild.Symbol.Nonterminal()
	if !ok {
		return "", errors.New("no nonterminal")
	}
	if idNonterminal != parser.NonterminalIdColon {
		return "", errors.New("expected nonterminal id colon")
	}

	firstChildChild := firstChild.Children[0]
	idTerminal, ok := firstChildChild.Symbol.Terminal()
	if !ok {
		return "", errors.New("no token")
	}
	if idTerminal != parser.TokenIdColon {
		return "", errors.New("expected token id colon")
	}
	return string(firstChildChild.Lexeme), nil
}
