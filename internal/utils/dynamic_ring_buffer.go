package utils

// DynamicRingBuffer is a ring buffer which is growing in size when it is filled up. This still provides the benefit
// of not allocating new memory when elements are added and removed in the same rate, but allows for a dynamic
// growth when necessary.
// DynamicRingBuffer is not thread safe and must only be used in one go routine.
type DynamicRingBuffer[T any] struct {
	buffer []T

	readIndex  int
	writeIndex int
	count      int
}

// NewDynamicRingBuffer allocates a new DynamicRingBuffer with the default capacity.
func NewDynamicRingBuffer[T any]() DynamicRingBuffer[T] {
	return NewDynamicRingBufferWithCapacity[T](1024)
}

// NewDynamicRingBufferWithCapacity allocates a new DynamicRingBuffer with the given elements count as capacity.
func NewDynamicRingBufferWithCapacity[T any](capacity int) DynamicRingBuffer[T] {
	return DynamicRingBuffer[T]{
		buffer: make([]T, capacity),
	}
}

// Length returns the number of elements currently available.
func (b *DynamicRingBuffer[T]) Length() int {
	return b.count
}

// IsEmpty returns true if there are no elements available.
func (b *DynamicRingBuffer[T]) IsEmpty() bool {
	return b.count == 0
}

// Capacity returns the capacity as the number of elements which can be stored without growing.
func (b *DynamicRingBuffer[T]) Capacity() int {
	return len(b.buffer)
}

// Reset removes all elements from the DynamicRingBuffer.
func (b *DynamicRingBuffer[T]) Reset() {
	b.readIndex = 0
	b.writeIndex = 0
	b.count = 0
}

// Add adds the given item to the end of the DynamicRingBuffer. It automatically grows the capacity when there is
// no more space available.
func (b *DynamicRingBuffer[T]) Add(item T) {
	if b.count == len(b.buffer) {
		b.Resize(len(b.buffer) * 15 / 10) // grow by 50%
	}

	b.buffer[b.writeIndex] = item
	b.writeIndex++
	b.writeIndex %= len(b.buffer)
	b.count++
}

// Remove removes the next item from the start of the DynamicRingBuffer and returns it.
// Remove will panic if the DynamicRingBuffer is empty.
func (b *DynamicRingBuffer[T]) Remove() T {
	if b.count == 0 {
		panic("Cannot remove item from empty DynamicRingBuffer[T].")
	}

	item := b.buffer[b.readIndex]
	b.readIndex++
	b.readIndex %= len(b.buffer)
	b.count--
	return item
}

// RemoveN removes the next n items from the start of the DynamicRingBuffer without returning them.
// RemoveN will panic if more items are removed than are available.
func (b *DynamicRingBuffer[T]) RemoveN(n int) {
	if b.count < n {
		panic("Cannot remove item from empty DynamicRingBuffer[T].")
	}
	b.readIndex += n
	b.readIndex %= len(b.buffer)
	b.count -= n
}

// Get returns the item at the given index without removing it from the DynamicRingBuffer. Valid values for the index
// are 0 to Length()-1. The index 0 is always the same item which Remove would return.
// Get will panic if an index is accessed which does not belong to an item currently stored.
func (b *DynamicRingBuffer[T]) Get(index int) T {
	if index < 0 || b.count <= index {
		panic("Index out of range for DynamicRingBuffer[T].")
	}
	return b.buffer[(b.readIndex+index)%len(b.buffer)]
}

// Resize allocates a new buffer with the desired capacity and copies all elements over. If the new capacity is less
// than the current number of items stored inside the DynamicRingBuffer, the current number of items will be used
// as the new capacity instead. No items can be removed by resizing to a smaller capacity.
func (b *DynamicRingBuffer[T]) Resize(newCapacity int) {
	newCapacity = max(b.count, newCapacity)
	newBuffer := make([]T, newCapacity)

	if b.readIndex < b.writeIndex {
		// Data is contiguous in the current buffer. We can copy everything over in one go.
		copy(newBuffer, b.buffer[b.readIndex:b.writeIndex])
	} else {
		// Data wraps around the end of the current buffer. We need to copy in two steps.
		n := copy(newBuffer, b.buffer[b.readIndex:len(b.buffer)])
		copy(newBuffer[n:], b.buffer[:b.writeIndex])
	}

	b.buffer = newBuffer
	b.readIndex = 0
	b.writeIndex = b.count
}
