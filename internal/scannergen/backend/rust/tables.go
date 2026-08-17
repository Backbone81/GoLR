package rust

import (
	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// invalidTokenName is the name of the token the generated scanner uses when it has no token, which is what the accept
// table holds for a state which does not accept.
const invalidTokenName = "InvalidToken"

// Tables holds the lookup tables of a table driven scanner in the form the template writes them out. It is the same
// data the Go table driven backend emits, differing only in the type a table is given, which is a Rust one here.
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
	// no state occupies. It has the same entry type as ByteClassByByte, so that the driver can compare the two
	// without a conversion.
	TransitionCheck utils.IntArray

	// NoByteClass is the entry TransitionCheck uses for a cell which no state occupies. It is one past the highest
	// byte class in use, so no lookup can ever ask for it.
	NoByteClass int

	// AcceptTokenByState holds, for every state, the name of the token variant the state accepts, or the name of the
	// invalid token for a state which does not accept.
	//
	// The names are written into the generated source instead of the numbers they stand for, which is what makes the
	// table a [Token] a reader of the generated scanner can follow.
	AcceptTokenByState []string
}

// NewTables compresses the given DFA into the lookup tables the generated scanner reads at runtime.
func NewTables(dfa backend.DFA) Tables {
	compressed := table.NewCompressedDFA(dfa)
	classCount := compressed.ByteClasses.Count()

	// The check entries hold a byte class or the one value past them which means the cell is unused, so both tables
	// are typed by that value and can be compared against each other directly.
	classType := utils.RustUintType(classCount)

	byteClassByByte := make([]int, table.ByteValueCount)
	for byteValue := range byteClassByByte {
		byteClassByByte[byteValue] = compressed.ByteClasses.ClassByByte[byteValue]
	}

	transitionNext := make([]int, len(compressed.Transitions.Next))
	transitionCheck := make([]int, len(compressed.Transitions.Check))
	for cellIdx := range transitionNext {
		// A cell which no state occupies is never read, because the check entry keeps the driver from using it.
		// Writing a zero into it keeps the entry type free of a negative value.
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

	return Tables{
		ByteClassByByte: utils.NewTypedIntArray(classType, byteClassByByte),
		TransitionBase:  utils.NewRustIntArray(compressed.Transitions.Base),
		TransitionNext:  utils.NewRustIntArray(transitionNext),
		TransitionCheck: utils.NewTypedIntArray(classType, transitionCheck),

		NoByteClass:        classCount,
		AcceptTokenByState: acceptTokenByState,
	}
}
