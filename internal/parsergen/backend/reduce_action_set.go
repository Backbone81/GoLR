package backend

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

// ReduceActionSet is an ordered set of ReduceAction elements.
type ReduceActionSet struct {
	actions []ReduceAction
}

// NewReduceActionSet creates a new ordered reduce action set.
func NewReduceActionSet(reduceActions ...ReduceAction) ReduceActionSet {
	var result ReduceActionSet
	for _, action := range reduceActions {
		result.Add(action)
	}
	return result
}

// String returns a string representation of the ordered set with all its values.
func (s *ReduceActionSet) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for i := range s.actions {
		if i > 0 {
			builder.WriteString(", ")
		}
		fmt.Fprintf(&builder, "%s", &s.actions[i])
	}
	builder.WriteString("}")
	return builder.String()
}

// Add adds a new value to the ordered set. If the value is already present in the ordered set, the set is not changed.
// The return value reports if the value was added.
func (s *ReduceActionSet) Add(value ReduceAction) bool {
	index, found := slices.BinarySearchFunc(s.actions, value, CompareReduceAction)
	if found {
		return false
	}
	s.actions = slices.Insert(s.actions, index, value)
	return true
}

// Merge adds all values of the other ordered set.
func (s *ReduceActionSet) Merge(other *ReduceActionSet) {
	for _, value := range other.All() {
		s.Add(value)
	}
}

// Contains reports if the value is part of the ordered set or not.
func (s *ReduceActionSet) Contains(value ReduceAction) bool {
	_, found := slices.BinarySearchFunc(s.actions, value, CompareReduceAction)
	return found
}

// GetByIndex returns the value by its index of the ordered set.
func (s *ReduceActionSet) GetByIndex(index int) ReduceAction {
	return s.actions[index]
}

// Length returns the number of values of the ordered set.
func (s *ReduceActionSet) Length() int {
	return len(s.actions)
}

// IsEmpty reports if the ordered set is empty.
func (s *ReduceActionSet) IsEmpty() bool {
	return len(s.actions) == 0
}

// Equal reports if the ordered set has the same values as the other ordered set.
func (s *ReduceActionSet) Equal(other *ReduceActionSet) bool {
	return slices.EqualFunc(s.actions, other.actions, ReduceActionEqual)
}

// All returns an iterator over all values of the ordered set.
func (s *ReduceActionSet) All() iter.Seq2[int, ReduceAction] {
	return func(yield func(int, ReduceAction) bool) {
		for idx, value := range s.actions {
			if !yield(idx, value) {
				return
			}
		}
	}
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s ReduceActionSet) MarshalYAML() ([]byte, error) {
	if len(s.actions) == 0 {
		return yaml.Marshal(nil)
	}
	return yaml.Marshal(s.actions)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *ReduceActionSet) UnmarshalYAML(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	return yaml.Unmarshal(b, &s.actions)
}
