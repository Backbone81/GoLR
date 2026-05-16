package backend

import (
	"github.com/backbone81/golr/internal/utils"
)

// CoreSet is an ordered set of cores.
type CoreSet struct {
	utils.OrderedSet[Core]
}

// NewCoreSet creates a new core set.
func NewCoreSet(values ...Core) CoreSet {
	return CoreSet{
		OrderedSet: utils.NewOrderedSet[Core](values...),
	}
}

func (s CoreSet) Equal(other *CoreSet) bool {
	return s.OrderedSet.Equal(&other.OrderedSet)
}
