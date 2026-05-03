package utils

import (
	"cmp"
	"fmt"
	"hash/fnv"
	"iter"
	"slices"
	"strings"
	"unsafe"

	"github.com/goccy/go-yaml"
)

// OrderedSet provides an implementation for a set of values. Values are guaranteed to be only present once within the
// set. Values within the set are ordered.
type OrderedSet[T cmp.Ordered] struct {
	// data holds the values for the ordered set. The values are always sorted in ascending order.
	data []T
}

// NewOrderedSet creates a new ordered set with the values added to it.
func NewOrderedSet[T cmp.Ordered](values ...T) OrderedSet[T] {
	var result OrderedSet[T]
	for _, value := range values {
		result.Add(value)
	}
	return result
}

// String returns a string representation of the ordered set with all its values.
func (s *OrderedSet[T]) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	for i := range s.data {
		if i > 0 {
			builder.WriteString(", ")
		}
		fmt.Fprintf(&builder, "%v", &s.data[i])
	}
	builder.WriteString("}")
	return builder.String()
}

// Add adds a new value to the ordered set. If the value is already present in the ordered set, the set is not changed.
// The return value reports if the value was added.
func (s *OrderedSet[T]) Add(value T) bool {
	index, found := slices.BinarySearch(s.data, value)
	if found {
		return false
	}
	s.data = slices.Insert(s.data, index, value)
	return true
}

// Remove removes a value from the ordered set. If the value is not present in the ordered set, the set is not changed.
// The return value reports if the value was removed.
func (s *OrderedSet[T]) Remove(value T) bool {
	index, found := slices.BinarySearch(s.data, value)
	if !found {
		return false
	}
	s.data = slices.Delete(s.data, index, index+1)
	return true
}

// Merge adds all values of the other ordered set.
func (s *OrderedSet[T]) Merge(other *OrderedSet[T]) {
	for _, value := range other.All() {
		s.Add(value)
	}
}

// Contains reports if the value is part of the ordered set or not.
func (s *OrderedSet[T]) Contains(value T) bool {
	_, found := slices.BinarySearch(s.data, value)
	return found
}

// GetByIndex returns the value by its index of the ordered set.
func (s *OrderedSet[T]) GetByIndex(index int) T {
	return s.data[index]
}

// Length returns the number of values of the ordered set.
func (s *OrderedSet[T]) Length() int {
	return len(s.data)
}

// IsEmpty reports if the ordered set is empty.
func (s *OrderedSet[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Hash calculates a hash over all values of the ordered set.
func (s *OrderedSet[T]) Hash() uint64 {
	hash := fnv.New64a()
	if len(s.data) > 0 {
		// We are converting the slice of values into a slice of bytes for calculating the hash. We do this with unsafe
		// pointer arithmetic to avoid rewriting data only for the hash, which we already have at hand.
		dataByteSize := len(s.data) * int(unsafe.Sizeof(s.data[0]))
		dataBytes := unsafe.Slice((*byte)(unsafe.Pointer(&s.data[0])), dataByteSize)
		if _, err := hash.Write(dataBytes); err != nil {
			panic(err)
		}
	}
	return hash.Sum64()
}

// Equal reports if the ordered set has the same values as the other ordered set.
func (s *OrderedSet[T]) Equal(other *OrderedSet[T]) bool {
	return slices.Equal(s.data, other.data)
}

// All returns an iterator over all values of the ordered set.
func (s *OrderedSet[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for idx, value := range s.data {
			if !yield(idx, value) {
				return
			}
		}
	}
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s OrderedSet[T]) MarshalYAML() ([]byte, error) {
	if len(s.data) == 0 {
		return yaml.Marshal(nil)
	}
	return yaml.Marshal(s.data)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *OrderedSet[T]) UnmarshalYAML(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	return yaml.Unmarshal(b, &s.data)
}

// CompareOrderedSet reports how two ordered sets should be ordered.
func CompareOrderedSet[T cmp.Ordered](lhs OrderedSet[T], rhs OrderedSet[T]) int {
	return slices.Compare(lhs.data, rhs.data)
}
