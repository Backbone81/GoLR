package utils

// Stack provides a first in last out container.
type Stack[T any] struct {
	data []T
}

// Push adds the value to the top of the stack.
func (s *Stack[T]) Push(value T) {
	s.data = append(s.data, value)
}

// Pop removed the value at the top of the stack.
func (s *Stack[T]) Pop() {
	s.data = s.data[:len(s.data)-1]
}

// Top returns the value at the top of the stack.
//
//nolint:ireturn // This is a false positive. We are in fact returning the value stored inside.
func (s *Stack[T]) Top() T {
	return s.data[len(s.data)-1]
}

// Size returns the number of elements on the stack.
func (s *Stack[T]) Size() int {
	return len(s.data)
}

// IsEmpty reports if the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return s.Size() == 0
}
