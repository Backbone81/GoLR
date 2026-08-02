package table

import (
	"fmt"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/utils"
)

// TransitionTable is the DFA with its transitions laid out as a dense table indexed by state and byte class, which is
// the form the row displacement compression consumes.
type TransitionTable struct {
	// ByteClasses maps an input byte to the column of Rows which holds the transition for it.
	ByteClasses ByteClasses

	// Rows holds one row per DFA state, each of ByteClasses.Count entries. An entry is the index of the state to
	// transition to, or NoTransition when the state has no transition on the bytes of that class.
	Rows [][]int
}

// NewTransitionTable computes the byte equivalence classes of the given DFA and lays its transitions out as rows
// indexed by those classes.
func NewTransitionTable(dfa backend.DFA) TransitionTable {
	byteClasses := NewByteClasses(dfa)
	classCount := byteClasses.Count()

	rows := make([][]int, len(dfa.States))
	var targets [ByteValueCount]int
	for stateIdx := range dfa.States {
		fillTargets(&targets, &dfa.States[stateIdx])

		row := make([]int, classCount)
		for classIdx := range row {
			row[classIdx] = NoTransition
		}
		for byteValue := range targets {
			classIdx := byteClasses.ClassByByte[byteValue]
			utils.DebugAssert(func() error {
				// Every byte of a class transitions to the same state, otherwise the class would have been
				// split. Writing the entry more than once must therefore always write the same value.
				if row[classIdx] != NoTransition && row[classIdx] != targets[byteValue] {
					return fmt.Errorf(
						"state %d transitions on byte class %d to state %d and to state %d",
						stateIdx, classIdx, row[classIdx], targets[byteValue],
					)
				}
				return nil
			})
			row[classIdx] = targets[byteValue]
		}
		rows[stateIdx] = row
	}

	return TransitionTable{
		ByteClasses: byteClasses,
		Rows:        rows,
	}
}
