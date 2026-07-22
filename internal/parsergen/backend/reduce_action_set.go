package backend

import (
	"encoding/json"
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

// Remove removes a value from the ordered set. If the value is not present in the ordered set, the set is not changed.
// The return value reports if the value was removed.
func (s *ReduceActionSet) Remove(value ReduceAction) bool {
	index, found := slices.BinarySearchFunc(s.actions, value, CompareReduceAction)
	if !found {
		return false
	}
	s.actions = slices.Delete(s.actions, index, index+1)
	return true
}

// Clear removes all reduce actions from the set while keeping the already allocated backing storage. Refilling the set
// afterward reuses that storage instead of allocating a new one, so this is what you want when a set is emptied and
// rebuilt with a similar number of reduce actions. Note that this keeps a reference to the previous reduce actions
// until they are overwritten, so it is not suitable for letting their lookahead sets be garbage collected.
func (s *ReduceActionSet) Clear() {
	s.actions = s.actions[:0]
}

// Clone returns a copy of the ordered set which shares no storage with the original. A plain copy of a reduce action
// set keeps referencing the actions of the original, so removing an action from the copy would remove it from the
// original as well. Clone is what you want when the original must stay untouched.
//
// The lookahead sets of the reduce actions are cloned as well, because a reduce action holds its lookahead set as a
// bitset, which shares its storage on a plain copy just the same.
func (s *ReduceActionSet) Clone() ReduceActionSet {
	result := ReduceActionSet{
		actions: slices.Clone(s.actions),
	}
	for i := range result.actions {
		result.actions[i].LookaheadSet = result.actions[i].LookaheadSet.Clone()
	}
	return result
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

// MarshalJSON implements the json.Marshaler interface.
func (s ReduceActionSet) MarshalJSON() ([]byte, error) {
	if len(s.actions) == 0 {
		return json.Marshal(nil)
	}
	return json.Marshal(s.actions)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *ReduceActionSet) UnmarshalJSON(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	return json.Unmarshal(b, &s.actions)
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
