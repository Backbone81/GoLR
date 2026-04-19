package golang

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"golr/internal/parsergen/backend"
	"golr/internal/parsergen/frontend"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"
)

//go:embed parser.go.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.go.template").Funcs(template.FuncMap{
	"stateActions":         buildStateActions,
	"gotoAfterNonterminal": buildGotoAfterNonterminal,
	"displayProduction":    displayProduction,
}).Parse(parserTemplate))

type Config struct {
	PackageName string
}

type TemplateContext struct {
	Config Config
	Parser backend.Parser
}

// FromParser writes the parser as Go source code to the given writer. Returns an error if the Go source code can not be
// encoded successfully.
func FromParser(writer io.Writer, parser backend.Parser, config Config) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Backends: Golang: FromParser").End()

	var buffer bytes.Buffer
	if err := parsedTemplate.Execute(&buffer, TemplateContext{
		Config: config,
		Parser: parser,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	source := buffer.Bytes()

	var joinedErr error
	formatted, err := format.Source(source)
	if err != nil {
		joinedErr = errors.Join(joinedErr, err)
	} else {
		source = formatted
	}

	if _, err := writer.Write(source); err != nil {
		joinedErr = errors.Join(joinedErr, err)
	}
	return joinedErr
}

// ParserToFile writes the parser as Go source code to the given file path. Returns an error if the file can not be
// written or the Go source code can not be encoded successfully.
func ParserToFile(filePath string, parser backend.Parser, config Config) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer file.Close()

	return FromParser(file, parser, config)
}

// ParserToString returns the parser as Go source code. Returns an error if the Go source code can not be encoded
// successfully.
func ParserToString(parser backend.Parser, config Config) (string, error) {
	var builder strings.Builder
	if err := FromParser(&builder, parser, config); err != nil {
		return "", err
	}
	return builder.String(), nil
}

type StateAction struct {
	LookaheadTerminalIdxs []int
	LookaheadSymbols      []frontend.Symbol
	IsReduce              bool

	// Reduce action: number of symbols to pop from stack.
	PopSymbolCount int

	// Reduce action: symbol to push onto stack.
	PushSymbol int

	// Shift action: state to push onto stack.
	PushState int

	// Reduce action: the production which is reduced
	ProductionIdx int
}

func buildStateActions(grammar frontend.Grammar, state backend.State) ([]StateAction, error) {
	var result []StateAction
	for _, reduceAction := range state.ReduceActions.All() {
		var action StateAction
		for symbol := range reduceAction.LookaheadSet.All() {
			action.LookaheadTerminalIdxs = append(action.LookaheadTerminalIdxs, symbol)
			action.LookaheadSymbols = append(action.LookaheadSymbols, grammar.Terminals[symbol])
		}
		action.IsReduce = true

		production := grammar.Productions[reduceAction.ProductionIdx]
		action.ProductionIdx = reduceAction.ProductionIdx
		action.PopSymbolCount = len(production.SymbolRefs)
		action.PushSymbol = production.NonterminalIdx
		result = append(result, action)
	}
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef().IsNonterminal() {
			continue
		}
		var action StateAction
		action.LookaheadTerminalIdxs = append(action.LookaheadTerminalIdxs, transitionAction.SymbolRef().Idx())
		action.LookaheadSymbols = append(action.LookaheadSymbols, grammar.Terminals[transitionAction.SymbolRef().Idx()])
		action.PushState = transitionAction.StateIdx()
		result = append(result, action)
	}
	return result, nil
}

type Goto struct {
	SourceStateIdx      int
	DestinationStateIdx int
}

func buildGotoAfterNonterminal(parser backend.Parser, nonterminalIdx int) []Goto {
	var result []Goto
	nonterminalRef := frontend.NewNonterminalRef(nonterminalIdx)
	for stateIdx, state := range parser.States {
		for _, transitionAction := range state.TransitionActions.All() {
			if transitionAction.SymbolRef() == nonterminalRef {
				result = append(result, Goto{
					SourceStateIdx:      stateIdx,
					DestinationStateIdx: transitionAction.StateIdx(),
				})
				break
			}
		}
	}
	return result
}

func displayProduction(grammar frontend.Grammar, productionIdx int) string {
	var builder strings.Builder
	production := grammar.Productions[productionIdx]

	lhs := grammar.Nonterminals[production.NonterminalIdx]
	builder.WriteString(lhs.String())
	builder.WriteString(" -> ")
	for idx, symbolRef := range production.SymbolRefs {
		if idx > 0 {
			builder.WriteString(" ")
		}
		if symbolRef.IsNonterminal() {
			builder.WriteString(grammar.Nonterminals[symbolRef.Idx()].String())
		} else {
			builder.WriteString(grammar.Terminals[symbolRef.Idx()].String())
		}
	}
	if len(production.SymbolRefs) == 0 {
		builder.WriteString("%empty")
	}
	return builder.String()
}
