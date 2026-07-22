package golr

import (
	"hash/fnv"
	"iter"
	"slices"
	"unsafe"

	"github.com/backbone81/golr/internal/parsergen/backend"
)

// ItemSet is an ordered set of LR(1) items. Every core occurs at most once within the set. Adding an item for a core
// which is already present merges the lookahead sets instead of storing a second item.
//
// Canonical LR(1) distinguishes states by their item set including the lookahead sets. Two item sets which agree on
// their cores but disagree on any lookahead set are different states.
type ItemSet struct {
	// items holds the items of the set, ordered ascending by their core.
	items []Item
}

// Add adds an item for the core with the given lookahead set. If the core is already present, the lookahead set is
// merged into the existing item. The return value reports if the item set changed.
//
// The lookahead set is copied, so that the caller can keep using its own lookahead set without affecting the item set.
func (s *ItemSet) Add(core backend.Core, lookaheadSet *backend.LookaheadSet) bool {
	idx, found := slices.BinarySearchFunc(s.items, core, compareItemToCore)
	if !found {
		var localLookaheadSet backend.LookaheadSet
		localLookaheadSet.Merge(lookaheadSet)
		s.items = slices.Insert(s.items, idx, Item{
			Core:         core,
			LookaheadSet: localLookaheadSet,
		})
		return true
	}

	return s.items[idx].LookaheadSet.Merge(lookaheadSet)
}

// LookaheadSetForCore returns the lookahead set for the given core. The returned pointer is only valid until the next
// call to Add, as adding an item can move the items around in memory.
func (s *ItemSet) LookaheadSetForCore(core backend.Core) *backend.LookaheadSet {
	idx, found := slices.BinarySearchFunc(s.items, core, compareItemToCore)
	if !found {
		return nil
	}
	return &s.items[idx].LookaheadSet
}

// All returns an iterator over all items of the item set, ordered ascending by their core.
func (s *ItemSet) All() iter.Seq2[int, Item] {
	return func(yield func(int, Item) bool) {
		for idx, item := range s.items {
			if !yield(idx, item) {
				return
			}
		}
	}
}

// CoreSet returns the cores of all items of the item set, dropping the lookahead sets.
func (s *ItemSet) CoreSet() backend.CoreSet {
	var result backend.CoreSet
	for _, item := range s.items {
		result.Add(item.Core)
	}
	return result
}

// Equal reports if this item set holds the same cores with the same lookahead sets as the other item set.
func (s *ItemSet) Equal(other *ItemSet) bool {
	if len(s.items) != len(other.items) {
		return false
	}
	for idx := range s.items {
		if s.items[idx].Core != other.items[idx].Core {
			return false
		}
		if !s.items[idx].LookaheadSet.Equal(other.items[idx].LookaheadSet) {
			return false
		}
	}
	return true
}

// Hash calculates a hash over all cores and their lookahead sets.
func (s *ItemSet) Hash() uint64 {
	hash := fnv.New64a()
	for idx := range s.items {
		// We are converting the core into a slice of bytes for calculating the hash. We do this with unsafe pointer
		// arithmetic to avoid rewriting data only for the hash, which we already have at hand.
		//nolint:gosec // unsafe is required for better performance
		coreBytes := unsafe.Slice((*byte)(unsafe.Pointer(&s.items[idx].Core)), unsafe.Sizeof(s.items[idx].Core))

		if _, err := hash.Write(coreBytes); err != nil {
			panic(err)
		}
		if _, err := hash.Write(s.items[idx].LookaheadSet.Bytes()); err != nil {
			panic(err)
		}
	}
	return hash.Sum64()
}

// compareItemToCore reports how an item is ordered relative to a core. Item sets are ordered by their core alone, as
// every core occurs at most once.
func compareItemToCore(item Item, core backend.Core) int {
	switch {
	case item.Core < core:
		return -1
	case item.Core > core:
		return 1
	default:
		return 0
	}
}
