package dot

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

//go:embed parser.dot.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.dot.template").Funcs(template.FuncMap{
	"dotString":         dotString,
	"stateLabel":        stateLabel,
	"transitionActions": transitionActions,
	"transitionLabel":   transitionLabel,
}).Parse(parserTemplate))

type TemplateContext struct {
	Parser backend.Parser
}

// FromParser writes the parser as DOT document to the given writer. Returns an error if the DOT document can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: DOT: FromParser").End()

	if err := parsedTemplate.Execute(writer, TemplateContext{
		Parser: parser,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	return nil
}

// ParserToFile writes the parser as DOT document to the given file path. Returns an error if the file can not be
// written or the DOT document can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the DOT file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromParser(file, parser)
}

// ParserToString returns the parser as DOT document. Returns an error if the DOT document can not be encoded
// successfully.
func ParserToString(parser backend.Parser) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// dotString wraps text in double quotes and escapes any internal double quotes.
func dotString(text string) string {
	text = strings.ReplaceAll(text, `"`, `\"`)
	return `"` + text + `"`
}

// transitionActions returns the transition actions of a state as a slice so the template can range over them.
func transitionActions(state backend.State) []backend.TransitionAction {
	var result []backend.TransitionAction
	for _, action := range state.TransitionActions.All() {
		result = append(result, action)
	}
	return result
}

// transitionLabel returns the display label for a transition edge.
func transitionLabel(grammar frontend.Grammar, action backend.TransitionAction) string {
	if action.SymbolRef().IsTerminal() {
		return grammar.Terminals[action.SymbolRef().Idx()].String()
	}
	return grammar.Nonterminals[action.SymbolRef().Idx()].String()
}

// stateLabel returns the DOT label attribute value for a state node, including the enclosing quotes.
// The label shows the state index and all kernel items with a dot marker indicating the parse position.
func stateLabel(grammar frontend.Grammar, stateIdx int, state backend.State) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, `State %d\n\l`, stateIdx)
	lastLhsSymbol := -1
	for _, kernelItem := range state.KernelItems.All() {
		production := grammar.Productions[kernelItem.ProductionIdx()]
		nonterminal := grammar.Nonterminals[production.NonterminalIdx]

		if lastLhsSymbol != production.NonterminalIdx {
			fmt.Fprintf(&builder, "%d %s:", kernelItem.ProductionIdx(), nonterminal)
			lastLhsSymbol = production.NonterminalIdx
		} else {
			fmt.Fprintf(&builder, "%d ", kernelItem.ProductionIdx())
			for range len(nonterminal.String()) {
				builder.WriteString(" ")
			}
			builder.WriteString("|")
		}
		for index, symbolRef := range production.SymbolRefs {
			var symbol frontend.Symbol
			if symbolRef.IsNonterminal() {
				symbol = grammar.Nonterminals[symbolRef.Idx()]
			} else {
				symbol = grammar.Terminals[symbolRef.Idx()]
			}
			if kernelItem.Position() == index {
				builder.WriteString(" •")
			}
			fmt.Fprintf(&builder, " %s", symbol)
		}
		if kernelItem.Position() == len(production.SymbolRefs) {
			builder.WriteString(" •")
		}
		builder.WriteString(`\l`)
	}
	return dotString(builder.String())
}
