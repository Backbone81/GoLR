package counterexample

import (
	"github.com/backbone81/golr/internal/utils"
)

// hashParser returns the hash of a SimulatedParser from the hash of its items and its depth.
func hashParser(itemsHash utils.Hash, depth int) uint64 {
	result := utils.NewHash()
	utils.WriteHash(&result, itemsHash)
	utils.WriteHash(&result, depth)
	return result.Sum64()
}

// hashConfiguration returns the hash of a ProductConfiguration from the hashes of its parsers and if the conflict
// terminal was shifted.
func hashConfiguration(parserHashes [2]uint64, terminalShifted bool) uint64 {
	var shifted uint64
	if terminalShifted {
		shifted = 1
	}
	result := utils.NewHash()
	utils.WriteHashSlice(&result, parserHashes[:])
	utils.WriteHash(&result, shifted)
	return result.Sum64()
}
