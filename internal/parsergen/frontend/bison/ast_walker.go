package bison

import (
	"errors"
	"golr/internal/parsergen/frontend"
	"golr/internal/parsergen/frontend/bison/parser"
)

type ASTWalker struct {
	grammar frontend.Grammar

	terminalIdxByName    map[string]int
	nonterminalIdxByName map[string]int
}

func NewASTWalker() *ASTWalker {
	return &ASTWalker{
		terminalIdxByName:    make(map[string]int),
		nonterminalIdxByName: make(map[string]int),
	}
}

func (w *ASTWalker) BuildGrammar(node *parser.Node) frontend.Grammar {
	w.buildTerminals(node)
	w.buildProductions(node)
	return w.grammar
}

func (w *ASTWalker) buildTerminals(node *parser.Node) {
	nonterminal, ok := node.Symbol.Nonterminal()
	if !ok {
		return
	}

	switch nonterminal {
	case
		parser.NonterminalInput,

		parser.NonterminalPrologueDeclarations,
		parser.NonterminalPrologueDeclaration,

		parser.NonterminalGrammar,
		parser.NonterminalRulesOrGrammarDeclaration,

		parser.NonterminalGrammarDeclaration,
		parser.NonterminalSymbolDeclaration,
		parser.NonterminalTokenDecls,
		parser.NonterminalTokenDecl_1:
		for _, child := range node.Children {
			w.buildTerminals(child)
		}
	case parser.NonterminalTokenDecl:
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
	}
}

func (w *ASTWalker) buildProductions(node *parser.Node) {
	nonterminal, ok := node.Symbol.Nonterminal()
	if !ok {
		return
	}

	switch nonterminal {
	case
		parser.NonterminalInput,
		parser.NonterminalGrammar,
		parser.NonterminalRulesOrGrammarDeclaration:
		for _, child := range node.Children {
			w.buildProductions(child)
		}
	case parser.NonterminalRules:
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
			w.buildProductions(child)
		}
	case parser.NonterminalRhses_1:
		if len(node.Children) == 3 {
			if terminal, ok := node.Children[1].Symbol.Terminal(); ok && terminal == parser.TokenPipe {
				// We create a new production with the same nonterminal on the left hand side.
				w.buildProductions(node.Children[0])
				w.grammar.Productions = append(w.grammar.Productions, frontend.Production{
					NonterminalIdx: w.grammar.Productions[len(w.grammar.Productions)-1].NonterminalIdx,
				})
				w.buildProductions(node.Children[2])
				return
			}
		}
		for _, child := range node.Children {
			w.buildProductions(child)
		}
	case parser.NonterminalRhs:
		for _, child := range node.Children {
			w.buildProductions(child)
		}
	case parser.NonterminalSymbol:
		id, err := w.getID(node)
		if err != nil {
			return
		}
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
