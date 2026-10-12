package protocol

import "encoding/json"

// Optional is an optional property which is either absent or holds a value. The zero value is absent.
type Optional[T any] struct {
	value T
	set   bool
}

// NewOptional returns an Optional holding the value.
func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{value: value, set: true}
}

// Get returns the value and whether it is present.
//
//nolint:ireturn // T is the type of the value, not an interface.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.set
}

// IsZero reports whether the value is absent, which leaves out a field tagged omitzero.
func (o Optional[T]) IsZero() bool {
	return !o.set
}

// MarshalJSON writes the value. An absent value is written as the zero value of T.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.value)
}
