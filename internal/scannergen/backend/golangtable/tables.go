package golangtable

import (
	"fmt"
	"strings"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/utils"
)

// valuesPerLine is the number of table entries written per line of source code. Tables can hold thousands of entries,
// so they are wrapped to keep the generated file readable.
const valuesPerLine = 16

// IntArray is a table of non-negative integers together with the narrowest Go integer type which can hold all of them.
// Picking the type per scanner instead of using a single wide type is what keeps the tables small, because the entries
// of a scanner with few states fit into a single byte each.
type IntArray struct {
	// Type is the Go type of an entry.
	Type string

	// Values are the entries of the table.
	Values []int
}

// NewIntArray returns the given values as a table typed by the largest value it has to hold.
func NewIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return IntArray{
		Type:   goUintType(maxValue),
		Values: values,
	}
}

// Literal returns the entries of the table as the body of a Go composite literal, wrapped over several lines.
func (a IntArray) Literal() string {
	var builder strings.Builder
	for i, value := range a.Values {
		if i%valuesPerLine == 0 {
			builder.WriteString("\n")
		}
		fmt.Fprintf(&builder, "%d, ", value)
	}
	builder.WriteString("\n")
	return builder.String()
}

// Tables holds the lookup tables of a table driven scanner in the form the template writes them out.
type Tables struct {
	// ByteClassByByte maps an input byte to the byte class it belongs to, which is the column of the transition
	// table to look at.
	ByteClassByByte IntArray

	// TransitionBase maps a state to the displacement of its row within TransitionNext.
	TransitionBase IntArray

	// TransitionNext holds the transitions of all states packed into a single table. The entry a state has for a
	// byte class lives at TransitionBase[state] + class, if TransitionCheck confirms that the cell belongs to that
	// class.
	TransitionNext IntArray

	// TransitionCheck holds the byte class every cell of TransitionNext belongs to, or NoByteClass for a cell which
	// no state occupies. It has the same entry type as ByteClassByByte, so that the driver can compare the two
	// without a conversion.
	TransitionCheck IntArray

	// NoByteClass is the entry TransitionCheck uses for a cell which no state occupies. It is one past the highest
	// byte class in use, so no lookup can ever ask for it.
	NoByteClass int

	// AcceptTokenByState holds, for every state, the name of the token constant the state accepts, or the name of
	// the invalid token for a state which does not accept.
	AcceptTokenByState []string
}

// NewTables compresses the given DFA into the lookup tables the generated scanner reads at runtime.
func NewTables(dfa backend.DFA) Tables {
	compressed := table.NewCompressedDFA(dfa)
	classCount := compressed.ByteClasses.Count()

	// The check entries hold a byte class or the one value past them which means the cell is unused, so both
	// tables are typed by that value and can be compared against each other directly.
	classType := goUintType(classCount)

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
		ByteClassByByte: IntArray{Type: classType, Values: byteClassByByte},
		TransitionBase:  NewIntArray(compressed.Transitions.Base),
		TransitionNext:  NewIntArray(transitionNext),
		TransitionCheck: IntArray{Type: classType, Values: transitionCheck},

		NoByteClass:        classCount,
		AcceptTokenByState: acceptTokenByState,
	}
}

// goUintType returns the narrowest unsigned Go integer type which can hold every value up to the given one.
func goUintType(maxValue int) string {
	switch {
	case maxValue <= 0xFF:
		return "uint8"
	case maxValue <= 0xFFFF:
		return "uint16"
	default:
		return "uint32"
	}
}
