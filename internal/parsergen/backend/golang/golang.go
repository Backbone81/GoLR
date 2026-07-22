package golang

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"runtime/trace"
	"strings"
	"text/template"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

//go:embed parser.go.template
var parserTemplate string

var parsedTemplate = template.Must(template.New("parser.go.template").Funcs(template.FuncMap{
	"stateActions":         buildStateActions,
	"defaultReduce":        buildDefaultReduce,
	"gotoAfterNonterminal": buildGotoAfterNonterminal,
	"displayProduction":    displayProduction,
	"terminalName":         terminalName,
	"nonterminalName":      nonterminalName,
}).Parse(parserTemplate))

const (
	// acceptProductionIdx is the production index of `$accept -> Start $end` in the augmented grammar. Reducing by it
	// means the parse is done, which the template renders as the accept instead of as a reduce.
	acceptProductionIdx = 0

	// acceptNonterminalIdx is the nonterminal index of `$accept` in the augmented grammar. It never appears on the right
	// hand side of a production, so no state ever has a goto on it.
	acceptNonterminalIdx = 0
)

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
func ParserToFile(filePath string, parser backend.Parser, config Config) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

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
		if reduceAction.ProductionIdx == acceptProductionIdx {
			// The accept is not keyed on a lookahead. The end of input marker has already been shifted when the state is
			// reached, so there is no terminal left to switch on and the reduction lookahead set is empty by
			// construction. buildDefaultReduce renders it as the unconditional default action instead, which is also
			// what the Bison backed cores report it as (`$default accept` in the XML report).
			continue
		}
		if reduceAction.LookaheadSet.IsEmpty() {
			// A reduce which no terminal can trigger is a dead action. Emitting it would produce a `case` without any
			// terminal to switch on, which is not valid Go.
			continue
		}

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

// buildDefaultReduce returns the action for the `default` arm of a state, or nil when the state has none and an
// unexpected terminal is an error there.
//
// Besides the default reduce a core asked for, this also covers the accept. The cores encode the accept differently:
// the Bison backed ones report it as `$default accept` and it arrives as DefaultReduceProductionIdx, while the native
// GoLR ones keep it a reduce of the accept production with the empty reduction lookahead set the DeRemer-Pennello
// computation yields for it. Both mean the same unconditional action, so both render into the default arm.
func buildDefaultReduce(grammar frontend.Grammar, state backend.State) (*StateAction, error) {
	productionIdx, ok := defaultReduceProductionIdx(state)
	if !ok {
		return nil, nil //nolint:nilnil // The nil value is our sentinel value and therefore fine here.
	}

	var action StateAction
	action.IsReduce = true

	production := grammar.Productions[productionIdx]
	action.ProductionIdx = productionIdx
	action.PopSymbolCount = len(production.SymbolRefs)
	action.PushSymbol = production.NonterminalIdx
	return &action, nil
}

// defaultReduceProductionIdx returns the production which reduces on any lookahead in the state. The boolean reports if
// there is one at all.
func defaultReduceProductionIdx(state backend.State) (int, bool) {
	if state.DefaultReduceProductionIdx != nil {
		return *state.DefaultReduceProductionIdx, true
	}
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.ProductionIdx == acceptProductionIdx {
			return acceptProductionIdx, true
		}
	}
	return 0, false
}

type Goto struct {
	SourceStateIdx      int
	DestinationStateIdx int
}

func buildGotoAfterNonterminal(parser backend.Parser, nonterminalIdx int) []Goto {
	if nonterminalIdx == acceptNonterminalIdx {
		// `$accept` never appears on the right hand side of a production, so no state has a goto on it and nothing ever
		// reduces to it either - the accept production ends the parse instead of pushing its left hand side. Its goto
		// function would be empty and unreferenced.
		return nil
	}

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

func terminalName(symbol frontend.Symbol) string {
	if symbol.Name == "$end" {
		return "EndToken"
	}
	name := utils.GoIdentifier(symbol.Name)
	return "Token" + name
}

func nonterminalName(symbol frontend.Symbol) string {
	name := utils.GoIdentifier(symbol.Name)
	return "Nonterminal" + name
}
