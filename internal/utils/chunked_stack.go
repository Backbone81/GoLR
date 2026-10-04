package utils

// chunkedStackChunkCapacity is the number of values a chunk of a ChunkedStack holds.
const chunkedStackChunkCapacity = 1024

// ChunkedStack provides a first in last out container like Stack, which stores its values in chunks of fixed capacity
// instead of one slice. A chunk is never copied, so the stack grows without copying its values, and a large stack is a
// few large objects for the garbage collector instead of a growing slice whose old arrays become garbage.
//
// An emptied chunk is kept for reuse for the lifetime of the stack.
type ChunkedStack[T any] struct {
	// chunks holds the values in the chunks up to activeChunk. Every chunk before activeChunk is full, the active chunk
	// is only empty when it is the first one, and the chunks behind it are empty.
	chunks      [][]T
	activeChunk int
}

// Push adds the value to the top of the stack.
func (s *ChunkedStack[T]) Push(value T) {
	switch {
	case len(s.chunks) == 0:
		s.chunks = append(s.chunks, make([]T, 0, chunkedStackChunkCapacity))
	case len(s.chunks[s.activeChunk]) == chunkedStackChunkCapacity:
		s.activeChunk++
		if s.activeChunk == len(s.chunks) {
			s.chunks = append(s.chunks, make([]T, 0, chunkedStackChunkCapacity))
		}
	}
	s.chunks[s.activeChunk] = append(s.chunks[s.activeChunk], value)
}

// Pop removes the value at the top of the stack.
func (s *ChunkedStack[T]) Pop() {
	chunk := s.chunks[s.activeChunk]
	s.chunks[s.activeChunk] = chunk[:len(chunk)-1]
	if len(chunk) == 1 && s.activeChunk > 0 {
		s.activeChunk--
	}
}

// Top returns the value at the top of the stack.
//
//nolint:ireturn // This is a false positive. We are in fact returning the value stored inside.
func (s *ChunkedStack[T]) Top() T {
	chunk := s.chunks[s.activeChunk]
	return chunk[len(chunk)-1]
}

// Size returns the number of elements on the stack.
func (s *ChunkedStack[T]) Size() int {
	if len(s.chunks) == 0 {
		return 0
	}
	return s.activeChunk*chunkedStackChunkCapacity + len(s.chunks[s.activeChunk])
}

// IsEmpty reports if the stack is empty.
func (s *ChunkedStack[T]) IsEmpty() bool {
	return s.Size() == 0
}
