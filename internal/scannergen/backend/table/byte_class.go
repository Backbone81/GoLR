package table

import (
	"github.com/backbone81/golr/internal/scannergen/backend"
)

// classTarget is the pair which decides the class a byte belongs to after a refinement step: the class the byte was in
// before, and the state it transitions to in the state doing the refining.
type classTarget struct {
	class  int
	target int
}

// ByteClasses partitions the possible input bytes into equivalence classes. Two bytes belong to the same class exactly
// when every state of the DFA transitions on both of them to the same target state, which makes them
// indistinguishable to the scanner.
//
// This is what makes a transition row small enough to be worth storing: a scanner which reads UTF-8 distinguishes far
// fewer byte values than it has bytes available, because whole ranges like the continuation bytes 0x80-0xBF are
// treated alike everywhere. Indexing a row by the class instead of by the byte turns a row of ByteValueCount entries
// into a row of Count entries.
//
// This follows the same idea as the column merging in section 3.3 of "Optimization of Parser Tables for Portable
// Compilers" by Dencker, Dürre and Heuft, where a vector maps every index to its class so that the table can be
// indexed by class instead. That scheme merges columns which are already identical, while we arrive at the classes by
// refining a partition.
//
// The zero value has every byte in class 0, which is the partition all bytes are indistinguishable in and the one the
// refinement starts from.
type ByteClasses struct {
	// ClassByByte maps an input byte to the index of the equivalence class it belongs to. Class indices are dense,
	// so they run from 0 to Count - 1.
	ClassByByte [ByteValueCount]int
}

// NewByteClasses computes the byte equivalence classes of the given DFA.
//
// The partition starts out with all bytes in a single class and is refined by every state in turn: a state splits a
// class whenever it does not transition to the same target state on all bytes of that class. Refinement only ever
// splits classes, so the order in which the states are visited does not change the resulting partition.
func NewByteClasses(dfa backend.DFA) ByteClasses {
	var result ByteClasses

	var targets [ByteValueCount]int
	for stateIdx := range dfa.States {
		fillTargets(&targets, &dfa.States[stateIdx])
		result.refine(&targets)
	}
	return result
}

// Count returns the number of distinct equivalence classes, and therefore the width of a transition row.
//
// The class indices are dense, so this is the highest class index in use plus one.
func (b *ByteClasses) Count() int {
	var maxClass int
	for _, class := range b.ClassByByte {
		maxClass = max(maxClass, class)
	}
	return maxClass + 1
}

// refine splits every class whose bytes the given target states do not agree on. Two bytes stay in the same class only
// if they were in the same class before and transition to the same target state.
func (b *ByteClasses) refine(targets *[ByteValueCount]int) {
	// Handing out the new class indices in the order the pairs are encountered keeps them dense.
	classByPair := make(map[classTarget]int)

	var count int
	for byteValue := range b.ClassByByte {
		pair := classTarget{
			class:  b.ClassByByte[byteValue],
			target: targets[byteValue],
		}
		class, ok := classByPair[pair]
		if !ok {
			class = count
			classByPair[pair] = class
			count++
		}
		b.ClassByByte[byteValue] = class
	}
}
