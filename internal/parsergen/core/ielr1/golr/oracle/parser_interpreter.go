package oracle

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// AugmentGrammar (see frontend.AugmentGrammar) always inserts the EOF terminal as the first terminal and the
// `$accept -> Start EOF` production as the first production. Both indexes are therefore fixed at 0 for any augmented
// grammar, which is what the interpreter builds its accept and end-of-input handling on.
const (
	// eofTerminalIdx is the terminal index of the EOF symbol.
	eofTerminalIdx = 0

	// acceptProductionIdx is the production index of `$accept -> Start EOF`. Reducing by it means the parse is done.
	acceptProductionIdx = 0
)

// ParserActionKind classifies the single step an LR parser takes.
type ParserActionKind int

const (
	// ParserActionShift consumes the current input terminal and pushes a new state.
	ParserActionShift ParserActionKind = iota

	// ParserActionReduce reduces by a production, popping its right-hand side and pushing the goto state.
	ParserActionReduce

	// ParserActionAccept reports a successful parse (a reduce by the augmented start production).
	ParserActionAccept

	// ParserActionReject reports that no action applies for the current state and input terminal.
	ParserActionReject
)

// ParserAction is the outcome of a single ParserInterpreter step. It is deliberately a small comparable value type
// so that two interpreters can be compared in lockstep with a plain `==`: an identical sequence of actions from two
// parsers over the same input means they produced the same parse.
//
// Only the field relevant to Kind carries a meaningful value; the other is set to -1. This keeps two actions of the
// same kind but different payloads unequal, and two actions of different kinds unequal, without any special casing.
//
// ParserActionReject carries no payload on purpose. The input position at which a reject happens is fully determined
// by the preceding actions (the cursor only advances on shifts), so in a lockstep comparison two rejects that are
// compared are always at the same position — including it could never distinguish them and would only bake
// history-derived state into the value's identity. The position is diagnostic, not part of the action, and is
// exposed through ParserInterpreter.Offset.
type ParserAction struct {
	Kind ParserActionKind

	// TerminalIdx is the shifted terminal index for ParserActionShift, otherwise -1.
	TerminalIdx int

	// ProductionIdx is the reduced production index for ParserActionReduce, otherwise -1.
	ProductionIdx int
}

// ParserAction implements fmt.Stringer.
var _ fmt.Stringer = (*ParserAction)(nil)

// String returns a string representation, used for readable failure diagnostics in the differential test.
func (a ParserAction) String() string {
	switch a.Kind {
	case ParserActionShift:
		return fmt.Sprintf("shift(terminal %d)", a.TerminalIdx)
	case ParserActionReduce:
		return fmt.Sprintf("reduce(production %d)", a.ProductionIdx)
	case ParserActionAccept:
		return "accept"
	case ParserActionReject:
		return "reject"
	default:
		return "unknown"
	}
}

// ParserInterpreter is a stepwise LR interpreter over a resolved parser table. It owns its own state stack and input
// cursor, so two interpreters built from two different parser tables can be driven in lockstep and any divergence in
// their behavior — a different action, or one advancing the input while the other does not — surfaces immediately.
//
// The parser table is expected to be conflict-free (already resolved through conflict.Resolve). The interpreter does
// not resolve conflicts: for each state it takes the first applicable action it finds, which is only well defined when
// at most one action applies. Applying it to an unresolved table is a programming error, not an interpreter concern.
type ParserInterpreter struct {
	parser backend.Parser

	// input is the sequence of terminal indexes to parse, ending in the EOF terminal (index 0).
	input  []int
	offset int

	// stateStack mirrors the generated parser's state stack. It starts with the start state 0.
	stateStack []int

	// stepCount and maxSteps bound the number of steps to guard against a reduce loop from a malformed table. A
	// well-formed table over a finite input always terminates well below the bound.
	stepCount int
	maxSteps  int

	// parserAction provides the last action taken.
	parserAction ParserAction

	// trace, when non-nil, receives one readable line per step describing the action the interpreter took. It is the
	// interpreter's own debugging channel, so a divergence between two tables can be inspected by tracing each of them
	// rather than reconstructing what happened by hand. It is nil and costs nothing unless WithTrace sets it.
	trace io.Writer
}

// ParserInterpreterOption customizes a ParserInterpreter at construction time.
type ParserInterpreterOption func(*ParserInterpreter)

// WithMaxSteps overrides the runaway step bound the interpreter guards against. It exists so two interpreters compared
// in lockstep can share one bound: the default bound scales with the table's state count, so an IELR(1) table (fewer
// states) and a canonical LR(1) table for the same grammar would otherwise cut a non-terminating parse — as a cyclic
// grammar produces — off at different steps and read as a spurious divergence instead of the agreement it is.
func WithMaxSteps(maxSteps int) ParserInterpreterOption {
	return func(i *ParserInterpreter) {
		i.maxSteps = maxSteps
	}
}

// WithTrace makes the interpreter write one readable line per step to the writer, describing the state the action was
// decided in, the lookahead, and the action taken (a reduce expanded into its named production). It turns the
// interpreter into its own tracer, so debugging a divergence between two tables is a matter of tracing each of them to a
// buffer and reading the two side by side. Passing nil disables tracing, which is also the default.
func WithTrace(writer io.Writer) ParserInterpreterOption {
	return func(i *ParserInterpreter) {
		i.trace = writer
	}
}

// DefaultMaxSteps returns the runaway step bound for an input of the given length (EOF included) parsed over a table
// with the given number of states. It is a loose but safe guard: every shift consumes input and every reduce without a
// shift is bounded by the automaton size, so this comfortably exceeds the steps any well-formed parse needs while still
// capping a reduce loop. It is exported so a lockstep comparison can size one shared bound off the larger of two tables.
func DefaultMaxSteps(inputLen int, stateCount int) int {
	return (inputLen+1)*(stateCount+1)*4 + 1024
}

// NewParserInterpreter creates an interpreter for the given resolved parser table and input. The input is a sequence of
// terminal indexes. The EOF terminal with index 0 is appended automatically, matching the shape produced by
// frontend.AugmentGrammar. The interpreter does not modify the parser table or the input.
func NewParserInterpreter(parser backend.Parser, input []int, options ...ParserInterpreterOption) *ParserInterpreter {
	input = append(input, eofTerminalIdx)
	interpreter := &ParserInterpreter{
		parser:     parser,
		input:      input,
		stateStack: []int{0},
		maxSteps:   DefaultMaxSteps(len(input), len(parser.States)),
	}
	for _, option := range options {
		option(interpreter)
	}
	return interpreter
}

// shiftAction builds an ParserActionShift for the given terminal index.
func (i *ParserInterpreter) shiftAction(terminalIdx int) ParserAction {
	return ParserAction{
		Kind:          ParserActionShift,
		TerminalIdx:   terminalIdx,
		ProductionIdx: -1,
	}
}

// reduceAction builds an ParserActionReduce for the given production index.
func (i *ParserInterpreter) reduceAction(productionIdx int) ParserAction {
	return ParserAction{
		Kind:          ParserActionReduce,
		TerminalIdx:   -1,
		ProductionIdx: productionIdx,
	}
}

// acceptAction builds an ParserActionAccept.
func (i *ParserInterpreter) acceptAction() ParserAction {
	return ParserAction{
		Kind:          ParserActionAccept,
		TerminalIdx:   -1,
		ProductionIdx: -1,
	}
}

// rejectAction builds an ParserActionReject.
func (i *ParserInterpreter) rejectAction() ParserAction {
	return ParserAction{
		Kind:          ParserActionReject,
		TerminalIdx:   -1,
		ProductionIdx: -1,
	}
}

// Offset returns the current input position. After a reject it is the position at which the parse got stuck. It is
// useful for diagnostics but deliberately not part of ParserAction's identity (see the ParserAction doc comment).
func (i *ParserInterpreter) Offset() int {
	return i.offset
}

// Next advances the parse by exactly one LR action and returns it. The action mutates the interpreter's own state stack
// and input cursor. It returns true as long as progress can be made.
//
// When a trace writer is set (see WithTrace) it writes one readable line per step, describing the state the action was
// decided in, the lookahead, and the action taken. The state stack and lookahead are captured before the step, because
// a shift advances the input cursor and a reduce rewrites the stack, so reading them afterwards would describe the next
// decision rather than this one.
func (i *ParserInterpreter) Next() bool {
	if i.parserAction.Kind == ParserActionAccept || i.parserAction.Kind == ParserActionReject {
		return false
	}

	var stackBeforeStep []int
	var lookaheadBeforeStep int
	if i.trace != nil {
		stackBeforeStep = slices.Clone(i.stateStack)
		lookaheadBeforeStep = i.input[i.offset]
	}

	i.advance()

	if i.trace != nil {
		i.writeTraceLine(stackBeforeStep, lookaheadBeforeStep)
	}
	return true
}

// advance takes the single LR action for the current state and lookahead, mutating the state stack and input cursor and
// recording it in parserAction. It always leaves parserAction set, so Next can report progress unconditionally. It is
// split out from Next so the trace can bracket exactly one step without the tracing having to be threaded through every
// branch.
func (i *ParserInterpreter) advance() {
	i.stepCount++
	if i.stepCount > i.maxSteps {
		i.parserAction = i.rejectAction()
		return
	}

	state := &i.parser.States[i.stateStack[len(i.stateStack)-1]]
	terminal := i.input[i.offset]

	// An explicit reduce action takes priority over a shift, mirroring the generated parser which lists reduce and
	// shift actions in one switch keyed by terminal. The accept production's reduce carries an empty lookahead set by
	// construction — nothing can follow the augmented start symbol — so it is matched on its production index and
	// fires as the accept at the end of input rather than through the lookahead set.
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.ProductionIdx == acceptProductionIdx || reduceAction.LookaheadSet.Contains(terminal) {
			i.reduce(reduceAction.ProductionIdx)
			return
		}
	}

	// Otherwise a shift: a terminal transition for the current terminal.
	for _, transitionAction := range state.TransitionActions.All() {
		symbolRef := transitionAction.SymbolRef()
		if symbolRef.IsTerminal() && symbolRef.Idx() == terminal {
			i.stateStack = append(i.stateStack, transitionAction.StateIdx())
			// Clamp the offset to the last input terminal, which is always EOF. Once EOF has been shifted the offset
			// stays put on it, so reading the current terminal never needs a bounds check and keeps returning EOF.
			i.offset = min(i.offset+1, len(i.input)-1)
			i.parserAction = i.shiftAction(terminal)
			return
		}
	}

	// Otherwise a default reduce on any lookahead, matching the `default:` case of a generated state function.
	if state.DefaultReduceProductionIdx != nil {
		i.reduce(*state.DefaultReduceProductionIdx)
		return
	}

	// No action applies for this state and terminal.
	i.parserAction = i.rejectAction()
}

// Value returns the parser action for the last Next() call.
func (i *ParserInterpreter) Value() ParserAction {
	return i.parserAction
}

// reduce performs a reduce by the given production . Reducing by the augmented start production
// accepts. A malformed table that would pop below the start state or lacks the required goto rejects rather than
// panicking, so a bug surfaces as a divergence instead of a crash.
func (i *ParserInterpreter) reduce(productionIdx int) {
	if productionIdx == acceptProductionIdx {
		i.parserAction = i.acceptAction()
		return
	}

	production := i.parser.Grammar.Productions[productionIdx]
	popCount := len(production.SymbolRefs)
	if popCount > len(i.stateStack)-1 {
		i.parserAction = i.rejectAction()
		return
	}
	i.stateStack = i.stateStack[:len(i.stateStack)-popCount]

	currentState := &i.parser.States[i.stateStack[len(i.stateStack)-1]]
	gotoState, ok := i.gotoStateForNonterminal(currentState, production.NonterminalIdx)
	if !ok {
		i.parserAction = i.rejectAction()
		return
	}
	i.stateStack = append(i.stateStack, gotoState)
	i.parserAction = i.reduceAction(productionIdx)
}

// writeTraceLine writes one step to the trace: the step number, the state the action was decided in and the full stack
// it sat on, the lookahead, and the action taken. The stack and lookahead are the ones captured before the step, so the
// line describes the decision rather than its after-effect. It is only called when a trace writer is set.
func (i *ParserInterpreter) writeTraceLine(stackBeforeStep []int, lookaheadBeforeStep int) {
	topStateIdx := stackBeforeStep[len(stackBeforeStep)-1]
	fmt.Fprintf(
		i.trace,
		"step %-5d state %-4d lookahead %-6s %-40s stack %v\n",
		i.stepCount, topStateIdx, i.formatTerminal(lookaheadBeforeStep), i.formatAction(), stackBeforeStep,
	)
}

// formatAction renders the last action for a trace line in the grammar's own vocabulary: a reduce expanded into its
// named production and a shift naming the terminal it consumes, so a reader does not have to translate indexes by hand.
func (i *ParserInterpreter) formatAction() string {
	switch i.parserAction.Kind {
	case ParserActionShift:
		return "shift " + i.formatTerminal(i.parserAction.TerminalIdx)
	case ParserActionReduce:
		return "reduce " + i.formatProduction(i.parserAction.ProductionIdx)
	default:
		return i.parserAction.String()
	}
}

// formatProduction renders a production as its named left and right hand sides, e.g. "N0 -> N1 N4 N1 N5", which is far
// easier to follow in a trace than a bare production index.
func (i *ParserInterpreter) formatProduction(productionIdx int) string {
	production := i.parser.Grammar.Productions[productionIdx]
	var builder strings.Builder
	builder.WriteString(i.parser.Grammar.Nonterminals[production.NonterminalIdx].Name)
	builder.WriteString(" ->")
	for _, symbolRef := range production.SymbolRefs {
		builder.WriteString(" ")
		builder.WriteString(i.formatSymbolRef(symbolRef))
	}
	return builder.String()
}

// formatSymbolRef renders a grammar symbol as its name, terminal or nonterminal alike.
func (i *ParserInterpreter) formatSymbolRef(symbolRef frontend.SymbolRef) string {
	if symbolRef.IsTerminal() {
		return i.formatTerminal(symbolRef.Idx())
	}
	return i.parser.Grammar.Nonterminals[symbolRef.Idx()].Name
}

// formatTerminal renders a terminal index as its name, or as "$" for the EOF terminal, so a trace reads in the grammar's
// own vocabulary rather than in indexes.
func (i *ParserInterpreter) formatTerminal(terminalIdx int) string {
	if terminalIdx == eofTerminalIdx {
		return "$"
	}
	return i.parser.Grammar.Terminals[terminalIdx].Name
}

// gotoStateForNonterminal finds the goto target state for the given nonterminal in the state's transition actions.
func (i *ParserInterpreter) gotoStateForNonterminal(state *backend.State, nonterminalIdx int) (int, bool) {
	for _, transitionAction := range state.TransitionActions.All() {
		symbolRef := transitionAction.SymbolRef()
		if symbolRef.IsNonterminal() && symbolRef.Idx() == nonterminalIdx {
			return transitionAction.StateIdx(), true
		}
	}
	return 0, false
}
