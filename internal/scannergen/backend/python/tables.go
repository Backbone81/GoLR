package python

import (
	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// invalidTokenName is the name of the token the generated scanner uses when it has no token, which is what the accept
// table holds for a state which does not accept.
const invalidTokenName = "INVALID_TOKEN"

// syntheticTokenCount is how many terminals the generated scanner declares before the tokens of the rules. They are
// terminals which no input can produce, so they take the lowest values, which makes this also the value of the first
// token of a rule.
const syntheticTokenCount = 3

// Tables holds the lookup tables of a table driven scanner in the form the template writes them out. It is the same
// data the Go table driven backend emits, with the entry types left off, because Python has a single integer type and
// nothing to narrow.
type Tables struct {
	// ByteClassByByte maps an input byte to the byte class it belongs to, which is the column of the transition
	// table to look at.
	ByteClassByByte utils.IntArray

	// TransitionBase maps a state to the displacement of its row within TransitionNext.
	TransitionBase utils.IntArray

	// TransitionNext holds the transitions of all states packed into a single table. The entry a state has for a
	// byte class lives at TransitionBase[state] + class, if TransitionCheck confirms that the cell belongs to that
	// class.
	TransitionNext utils.IntArray

	// TransitionCheck holds the byte class every cell of TransitionNext belongs to, or NoByteClass for a cell which
	// no state occupies.
	TransitionCheck utils.IntArray

	// NoByteClass is the entry TransitionCheck uses for a cell which no state occupies. It is one past the highest
	// byte class in use, so no lookup can ever ask for it.
	NoByteClass int

	// AcceptTokenByState holds, for every state, the name of the token member the state accepts, or the name of the
	// invalid token for a state which does not accept.
	//
	// The names are written into the generated source instead of the numbers they stand for, which is what makes the
	// table a table of tokens a reader of the generated scanner can follow.
	AcceptTokenByState []string

	// SkippedTokens holds the name of the token member of every rule the grammar marked for skipping.
	//
	// It is a table rather than a test per rule, so that the code which reads it is the same for every grammar. A
	// grammar which skips nothing leaves it empty, which is an empty set and needs no case of its own.
	SkippedTokens []string
}

// NewTables compresses the given DFA into the lookup tables the generated scanner reads at runtime.
func NewTables(dfa backend.DFA) Tables {
	compressed := table.NewCompressedDFA(dfa)
	classCount := compressed.ByteClasses.Count()

	byteClassByByte := make([]int, table.ByteValueCount)
	for byteValue := range byteClassByByte {
		byteClassByByte[byteValue] = compressed.ByteClasses.ClassByByte[byteValue]
	}

	transitionNext := make([]int, len(compressed.Transitions.Next))
	transitionCheck := make([]int, len(compressed.Transitions.Check))
	for cellIdx := range transitionNext {
		// A cell which no state occupies is never read, because the check entry keeps the driver from using it.
		// Writing a zero into it keeps the entries free of a negative value.
		transitionNext[cellIdx] = max(compressed.Transitions.Next[cellIdx], 0)

		transitionCheck[cellIdx] = classCount
		if compressed.Transitions.Check[cellIdx] != utils.NoColumn {
			transitionCheck[cellIdx] = compressed.Transitions.Check[cellIdx]
		}
	}

	acceptTokenByState := make([]string, len(compressed.AcceptRuleIdxByStateIdx))
	for stateIdx, ruleIdx := range compressed.AcceptRuleIdxByStateIdx {
		if ruleIdx == table.NoRule {
			acceptTokenByState[stateIdx] = invalidTokenName
			continue
		}
		acceptTokenByState[stateIdx] = tokenName(ruleIdx, dfa.Rules[ruleIdx])
	}

	var skippedTokens []string
	for ruleIdx, rule := range dfa.Rules {
		if rule.Skip {
			skippedTokens = append(skippedTokens, tokenName(ruleIdx, rule))
		}
	}

	return Tables{
		ByteClassByByte: utils.NewPythonIntArray(byteClassByByte),
		TransitionBase:  utils.NewPythonIntArray(compressed.Transitions.Base),
		TransitionNext:  utils.NewPythonIntArray(transitionNext),
		TransitionCheck: utils.NewPythonIntArray(transitionCheck),

		NoByteClass:        classCount,
		AcceptTokenByState: acceptTokenByState,
		SkippedTokens:      skippedTokens,
	}
}
