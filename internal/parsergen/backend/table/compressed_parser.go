package table

import (
	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// CompressedParser is a parser in the form a table driven parser reads it at runtime. It holds the decisions of the
// parser as lookup tables instead of as control flow, which is what allows a backend for a new language to be a small
// driver instead of another encoding of the state machine.
//
// The tables are compressed in two steps. A default first takes out of the rows what most of their entries agree on,
// which leaves the rows sparse, and the row displacement then packs the entries which are left into a single array. The
// action table defaults per state, to the reduction which the state performs on most of its lookaheads, which is
// backend.ApplyDefaultReductions. The goto table defaults per nonterminal, to the state which a goto on it leads to in
// most of the states which have one, see NewDefaultGotos. Both steps keep every decision the state model describes, so
// a lookup returns exactly what reading the states returns.
type CompressedParser struct {
	// Actions holds the actions of all states packed into a single array, indexed by state and terminal. A terminal
	// without an entry is one the state takes its default action for.
	Actions utils.RowDisplacement

	// DefaultActionByStateIdx holds, for every state, the action it takes for a terminal Actions has no entry for, or
	// NoAction when such a terminal is a syntax error there.
	DefaultActionByStateIdx []Action

	// Gotos holds the gotos of all states packed into a single array, indexed by state and nonterminal. An entry is
	// the index of the state the goto leads to. A nonterminal without an entry is one the state takes the default
	// goto of that nonterminal for.
	Gotos utils.RowDisplacement

	// DefaultGotoByNonterminalIdx holds, for every nonterminal, the state a goto on it leads to when Gotos has no
	// entry for it, or NoGoto when no state of the parser has a goto on it at all.
	DefaultGotoByNonterminalIdx []int

	// ErrorTerminalIdx is the terminal index of the error symbol, or NoTerminal when the grammar does not have it.
	// This is the column of Actions which the error recovery reads, see ErrorShiftStateIdx.
	ErrorTerminalIdx int
}

// NewCompressedParser compresses the given parser into the tables a table driven parser uses.
func NewCompressedParser(parser backend.Parser) CompressedParser {
	defaultGotoByNonterminalIdx := NewDefaultGotos(parser)
	gotoRows := NewGotoTable(parser)
	ApplyDefaultGotos(gotoRows, defaultGotoByNonterminalIdx)

	return CompressedParser{
		Actions:                     utils.NewRowDisplacement(NewActionTable(parser), int(NoAction)),
		DefaultActionByStateIdx:     NewDefaultActions(parser),
		Gotos:                       utils.NewRowDisplacement(gotoRows, NoGoto),
		DefaultGotoByNonterminalIdx: defaultGotoByNonterminalIdx,
		ErrorTerminalIdx:            errorTerminalIdx(parser.Grammar),
	}
}

// StateCount returns the number of states of the parser.
func (c *CompressedParser) StateCount() int {
	return len(c.DefaultActionByStateIdx)
}

// Action returns what the parser does in the given state when the scanner delivers the given terminal, or NoAction when
// the terminal is a syntax error there. This is the access code a generated table driver performs.
//
// The lookup is the whole decision, the default action of the state included: an entry which the state has of its own
// wins, and only a terminal without one falls through to the default action. That order is what keeps a terminal the
// state rejects on purpose an error even though the state reduces on everything else, see ActionKindError.
func (c *CompressedParser) Action(stateIdx int, terminalIdx int) Action {
	action := Action(c.Actions.Lookup(stateIdx, TerminalColumn(terminalIdx)))
	if action == NoAction {
		return c.DefaultActionByStateIdx[stateIdx]
	}
	return action
}

// Goto returns the state the parser continues in when it has reduced to the given nonterminal and uncovered the given
// state. This is the access code a generated table driver performs.
//
// The lookup is the whole decision, the default goto of the nonterminal included: an entry which the state has of its
// own wins, and only a nonterminal without one falls through to the default goto.
//
// A state which has no goto on the nonterminal at all is not distinguishable from a state whose goto the default
// covers, so this returns the default for it rather than NoGoto, see ApplyDefaultGotos. Only a nonterminal which no
// state of the parser has a goto on yields NoGoto.
func (c *CompressedParser) Goto(stateIdx int, nonterminalIdx int) int {
	if targetStateIdx := c.Gotos.Lookup(stateIdx, nonterminalIdx); targetStateIdx != NoGoto {
		return targetStateIdx
	}
	return c.DefaultGotoByNonterminalIdx[nonterminalIdx]
}

// ErrorShiftStateIdx returns the state the parser continues in when it shifts the error symbol in the given state, and
// reports whether the state can shift the error symbol at all. The states which can are the places the grammar marked
// to resume at after a syntax error, which is what the error recovery pops the stack down to.
//
// The error symbol is a terminal like any other, so its shift needs no table of its own: it is the entry of the action
// table in the column of the error terminal. Nothing else can read that column, because no scanner ever delivers the
// error symbol - the parser shifts it itself while recovering.
//
// The default action of the state is deliberately not consulted here. Only an entry the state has of its own is a place
// to resume at, and reducing instead of shifting would take the recovery to a state the grammar never marked.
func (c *CompressedParser) ErrorShiftStateIdx(stateIdx int) (int, bool) {
	if c.ErrorTerminalIdx == NoTerminal {
		return 0, false
	}
	action := Action(c.Actions.Lookup(stateIdx, TerminalColumn(c.ErrorTerminalIdx)))
	if action == NoAction || action.Kind() != ActionKindShift {
		return 0, false
	}
	return action.StateIdx(), true
}

// HasErrorRecovery reports whether any state of the parser can shift the error symbol, which is the case for a grammar
// which marks places to resume at after a syntax error. A generated parser leaves out the parts of the recovery which
// cost something on the hot path when no state can.
func (c *CompressedParser) HasErrorRecovery() bool {
	for stateIdx := range c.StateCount() {
		if _, ok := c.ErrorShiftStateIdx(stateIdx); ok {
			return true
		}
	}
	return false
}
