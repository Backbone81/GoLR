package backend

import (
	"github.com/backbone81/golr/internal/utils"
)

// LookaheadSet is a set of terminal indexes.
type LookaheadSet = utils.Bitset

// NewLookaheadSet creates a new lookahead set.
var NewLookaheadSet = utils.NewBitset
