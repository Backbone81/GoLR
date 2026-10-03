package utils

import (
	"fmt"
	"hash/fnv"
	"slices"
	"unsafe"
)

// Deque is a persistent sequence which grows at both ends and shrinks at its end. Every operation returns a new
// sequence and leaves the one it was called on unchanged, so many sequences derived from each other share most of their
// values instead of copying them. Pushing a value costs one allocation, dropping values costs none.
//
// The sequence is kept in two immutable singly linked lists, as in Okasaki, "Purely Functional Data Structures": front
// holds the values prepended, its top is the first value, and back holds the others, its top is the last value. The
// sequence is the values of front, followed by the values of back in reverse. A singly linked list grows and shrinks
// without touching a shared cell only at its top, which is why a single list cannot serve both ends. A doubly linked
// list cannot share cells at all, as every push would have to update a pointer of a shared cell.
//
// Dropping more values than back holds would reach the bottom of front, which a singly linked list does not give access
// to. The sequence is then rebuilt into a new back list first, which costs one allocation per value. A sequence which
// only drops values pushed at its end never pays for that.
//
// back is never empty while the sequence is not, so Last only reads the top of back. first points to the cell of the
// first value, which is the top of front or, while front is empty, the bottom of back, so First does not walk back. The
// zero value is an empty sequence.
type Deque[T any] struct {
	front    *dequeCell[T]
	back     *dequeCell[T]
	first    *dequeCell[T]
	frontLen int
	backLen  int
}

// dequeCell is a cell of the immutable singly linked lists of a Deque.
type dequeCell[T any] struct {
	value T
	next  *dequeCell[T]
}

// Len returns the number of values.
func (d Deque[T]) Len() int {
	return d.frontLen + d.backLen
}

// First returns the first value. The sequence must not be empty.
func (d Deque[T]) First() T { //nolint:ireturn // T is a concrete type.
	return d.first.value
}

// Last returns the last value. The sequence must not be empty.
func (d Deque[T]) Last() T { //nolint:ireturn // T is a concrete type.
	return d.back.value
}

// PushBack returns the sequence with the value appended.
func (d Deque[T]) PushBack(value T) Deque[T] {
	d.back = &dequeCell[T]{
		value: value,
		next:  d.back,
	}
	if d.Len() == 0 {
		d.first = d.back
	}
	d.backLen++
	return d
}

// PushFront returns the sequence with the value prepended. The first value of an empty sequence goes to back, which
// must not be empty.
func (d Deque[T]) PushFront(value T) Deque[T] {
	if d.backLen == 0 {
		return d.PushBack(value)
	}
	d.front = &dequeCell[T]{
		value: value,
		next:  d.front,
	}
	d.first = d.front
	d.frontLen++
	return d
}

// PopBack returns the sequence without its last count values, and appends these values to the buffer in order.
func (d Deque[T]) PopBack(count int, buffer []T) (Deque[T], []T) {
	d = d.withBackLen(count)
	from := len(buffer)
	for cell := d.back; len(buffer)-from < count; cell = cell.next {
		buffer = append(buffer, cell.value)
	}
	slices.Reverse(buffer[from:])
	return d.DropBack(count), buffer
}

// DropBack returns the sequence without its last count values.
func (d Deque[T]) DropBack(count int) Deque[T] {
	d = d.withBackLen(count)
	for range count {
		d.back = d.back.next
		d.backLen--
	}
	switch {
	case d.Len() == 0:
		d.first = nil
	case d.backLen == 0:
		d = d.withoutFront()
	}
	return d
}

// AppendAll appends the values to the buffer in order.
func (d Deque[T]) AppendAll(buffer []T) []T {
	for cell := d.front; cell != nil; cell = cell.next {
		buffer = append(buffer, cell.value)
	}
	from := len(buffer)
	for cell := d.back; cell != nil; cell = cell.next {
		buffer = append(buffer, cell.value)
	}
	slices.Reverse(buffer[from:])
	return buffer
}

// Hash calculates a hash over all values in order. Values are hashed by their memory representation. Equal sequences
// have equal hashes, no matter how their values are split between front and back. The values are collected in the
// buffer, which is returned for reuse.
func (d Deque[T]) Hash(buffer []T) (uint64, []T) {
	hash := fnv.New64a()
	values := d.AppendAll(buffer[:0])

	// We reinterpret the slice of values as a slice of bytes. We do this with unsafe pointer arithmetic to avoid
	// rewriting the values only for the hash. An empty sequence gives no bytes.
	valuesByteSize := len(values) * int(unsafe.Sizeof(values[0]))

	//nolint:gosec // unsafe is required for better performance
	valueBytes := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(values))), valuesByteSize)
	if _, err := hash.Write(valueBytes); err != nil {
		panic(err)
	}
	return hash.Sum64(), values
}

// withBackLen returns the same sequence with at least count values in back.
func (d Deque[T]) withBackLen(count int) Deque[T] {
	DebugAssert(func() error {
		if count > d.Len() {
			return fmt.Errorf("cannot remove %d of %d values", count, d.Len())
		}
		return nil
	})
	if count > d.backLen {
		return d.withoutFront()
	}
	return d
}

// withoutFront returns the same sequence with all its values in back.
func (d Deque[T]) withoutFront() Deque[T] {
	values := d.AppendAll(nil)
	var result Deque[T]
	for _, value := range values {
		result = result.PushBack(value)
	}
	return result
}

// DequeContains reports if the sequence holds the value.
func DequeContains[T comparable](d Deque[T], value T) bool {
	for _, cells := range [2]*dequeCell[T]{d.front, d.back} {
		for cell := cells; cell != nil; cell = cell.next {
			if cell.value == value {
				return true
			}
		}
	}
	return false
}
