package utils

import (
	"encoding/binary"
	"hash"
	"unsafe"
)

// Hash is an implementation of the 64 bit FNV-1a hash.
//
// We cannot use the hash/fnv implementation of the go standard library, because we need to be able to copy the current
// state of the hash when we calculate counterexamples.
//
// The zero value is not a valid hash, always use NewHash() to create a new one.
type Hash uint64

// Hash implements hash.Hash64.
var _ hash.Hash64 = (*Hash)(nil)

// The parameters of the 64 bit FNV-1a hash.
const (
	fnvOffsetBasis Hash = 14695981039346656037
	fnvPrime       Hash = 1099511628211
)

// NewHash returns the hash of no data.
func NewHash() Hash {
	return fnvOffsetBasis
}

// Write adds the data to the hash. It never returns an error.
func (h *Hash) Write(data []byte) (int, error) {
	result := *h
	for _, b := range data {
		result ^= Hash(b)
		result *= fnvPrime
	}
	*h = result
	return len(data), nil
}

// Sum appends the hash in big endian byte order to b.
func (h *Hash) Sum(b []byte) []byte {
	return binary.BigEndian.AppendUint64(b, uint64(*h))
}

// Reset returns the hash to the hash of no data.
func (h *Hash) Reset() {
	*h = fnvOffsetBasis
}

// Size returns the number of bytes Sum appends.
func (h *Hash) Size() int {
	return 8
}

// BlockSize returns the block size of the hash. FNV-1a takes the data byte by byte.
func (h *Hash) BlockSize() int {
	return 1
}

// Sum64 returns the hash.
func (h *Hash) Sum64() uint64 {
	return uint64(*h)
}

// HashInteger is the types WriteHash adds to a Hash.
type HashInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// WriteHash adds the bytes of the value in memory to the hash, in the byte order of the machine.
func WriteHash[T HashInteger](h *Hash, value T) {
	//nolint:gosec // unsafe is required for better performance
	_, _ = h.Write(unsafe.Slice((*byte)(unsafe.Pointer(&value)), unsafe.Sizeof(value)))
}

// WriteHashSlice adds the bytes of the values in memory to the hash, in the byte order of the machine. The hash is the
// same as with WriteHash for each value in order.
func WriteHashSlice[T HashInteger](h *Hash, values []T) {
	//nolint:gosec // unsafe is required for better performance
	valueBytes := unsafe.Slice(
		(*byte)(unsafe.Pointer(unsafe.SliceData(values))), len(values)*int(unsafe.Sizeof(values[0])),
	)
	_, _ = h.Write(valueBytes)
}
