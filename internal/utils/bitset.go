package utils

import (
	"fmt"
	"hash/fnv"
	"iter"
	"math"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unsafe"

	"github.com/goccy/go-yaml"
)

// Bitset provides efficient storage for interacting with a set of bits. Each bit is referenced by its index. The bitset
// will automatically grow when bits are set on an index which is beyond the current capacity. Reading bits outside
// the current capacity will not result in a resize and will always return that the bit is not set.
type Bitset struct {
	// chunks holds all bitset chunks making up the bitset.
	chunks []bitsetChunk
}

// NewBitset creates a new bitset with the given bits set.
func NewBitset(idxs ...int) Bitset {
	var result Bitset
	for _, idx := range idxs {
		result.Add(idx)
	}
	return result
}

// Length returns the number of bits set.
func (b *Bitset) Length() int {
	var result int
	for _, chunk := range b.chunks {
		result += chunk.Length()
	}
	return result
}

// Contains reports true when the bit at idx is currently set.
func (b *Bitset) Contains(idx int) bool {
	AssertValidIndex(idx, math.MaxInt)
	chunkIdx := idx / bitsetChunkMaxBits
	bitIdx := idx % bitsetChunkMaxBits
	if chunkIdx < len(b.chunks) {
		return b.chunks[chunkIdx].Contains(bitIdx)
	}
	return false
}

// IsEmpty reports if no bit of the Bitset is currently set.
func (b *Bitset) IsEmpty() bool {
	for _, chunk := range b.chunks {
		if !chunk.IsEmpty() {
			return false
		}
	}
	return true
}

// Add marks the bit at idx as being set. If idx is beyond the current storage capacity, the storage will be resized.
func (b *Bitset) Add(idx int) {
	AssertValidIndex(idx, math.MaxInt)
	chunkIdx := idx / bitsetChunkMaxBits
	bitIdx := idx % bitsetChunkMaxBits
	if len(b.chunks) <= chunkIdx {
		newChunks := make([]bitsetChunk, chunkIdx+1)
		copy(newChunks, b.chunks)
		b.chunks = newChunks
	}
	b.chunks[chunkIdx].Add(bitIdx)
}

// Remove marks the bit at idx as not being set. If idx is beyond the current storage capacity, no resize will happen.
func (b *Bitset) Remove(idx int) {
	AssertValidIndex(idx, math.MaxInt)
	chunkIdx := idx / bitsetChunkMaxBits
	bitIdx := idx % bitsetChunkMaxBits
	if chunkIdx < len(b.chunks) {
		b.chunks[chunkIdx].Remove(bitIdx)
	}
}

// All returns an iterator over all set bits.
func (b *Bitset) All() iter.Seq[int] {
	return func(yield func(int) bool) {
		for chunkIdx, chunk := range b.chunks {
			for idx := range chunk.All() {
				if !yield(chunkIdx*bitsetChunkMaxBits + idx) {
					return
				}
			}
		}
	}
}

// Merge adds all the bits set in the other bitset to the current one.
func (b *Bitset) Merge(other *Bitset) {
	commonChunks := min(len(b.chunks), len(other.chunks))
	for i := 0; i < commonChunks; i++ {
		b.chunks[i] |= other.chunks[i]
	}
	b.chunks = append(b.chunks, other.chunks[commonChunks:]...)
}

// Equal reports if this bitset is equal to the other bitset. The bitsets can be of different size and still be equal
// as long as the set bits are located at locations which both bitsets share.
func (b *Bitset) Equal(other Bitset) bool {
	// make sure the same chunks are equal
	for i := 0; i < min(len(b.chunks), len(other.chunks)); i++ {
		if b.chunks[i] != other.chunks[i] {
			return false
		}
	}

	// make sure excessive chunks are empty
	for i := len(b.chunks); i < len(other.chunks); i++ {
		if other.chunks[i] != 0 {
			return false
		}
	}
	for i := len(other.chunks); i < len(b.chunks); i++ {
		if b.chunks[i] != 0 {
			return false
		}
	}
	return true
}

// MarshalYAML implements the yaml.Marshaler interface.
func (b Bitset) MarshalYAML() ([]byte, error) {
	idxs := make([]int, 0, b.Length())
	for idx := range b.All() {
		idxs = append(idxs, idx)
	}
	if len(idxs) == 0 {
		return yaml.Marshal(nil)
	}
	return yaml.Marshal(idxs)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (b *Bitset) UnmarshalYAML(data []byte) error {
	if slices.Equal(data, []byte("null")) {
		return nil
	}
	var idxs []int
	err := yaml.Unmarshal(data, &idxs)
	if err != nil {
		return err
	}
	for _, idx := range idxs {
		b.Add(idx)
	}
	return nil
}

// Bitset implements Stringer.
var _ fmt.Stringer = (*Bitset)(nil)

// String returns a string representation,
func (b *Bitset) String() string {
	var builder strings.Builder
	builder.WriteString("{")
	firstEntry := true
	for idx := range b.All() {
		if !firstEntry {
			builder.WriteString(", ")
		}
		builder.WriteString(strconv.Itoa(idx))
		if firstEntry {
			firstEntry = false
		}
	}
	builder.WriteString("}")
	return builder.String()
}

// Hash returns a hash value over all chunks. All chunks except for trailing empty chunks contribute to the hash.
func (b *Bitset) Hash() uint64 {
	chunkCount := len(b.chunks)
	for ; chunkCount > 0 && b.chunks[chunkCount-1].IsEmpty(); chunkCount-- {
		// We remove empty chunks at the end, as we do not want them to contribute to the hash. Otherwise, two bitsets
		// with the same bits set but with different sizes would lead to different hashes.
	}

	hash := fnv.New64a()
	if chunkCount > 0 {
		chunksByteSize := chunkCount * int(unsafe.Sizeof(b.chunks[0]))
		chunksBytes := unsafe.Slice((*byte)(unsafe.Pointer(&b.chunks[0])), chunksByteSize)
		if _, err := hash.Write(chunksBytes); err != nil {
			panic(err)
		}
	}
	return hash.Sum64()
}

// bitsetChunk is a single chunk of the Bitset.
type bitsetChunk uint64

const (
	// bitsetChunkMaxBits is the number of bits which can be stored in bitsetChunk.
	bitsetChunkMaxBits = int(unsafe.Sizeof(bitsetChunk(0)) * 8)
)

func (c *bitsetChunk) Length() int {
	var result int
	localChunk := uint64(*c)
	for localChunk != 0 {
		result++
		// clear the least significant bit
		localChunk &= localChunk - 1
	}
	return result
}

func (c *bitsetChunk) Contains(idx int) bool {
	AssertValidIndex(idx, bitsetChunkMaxBits)
	return (*c & (1 << idx)) != 0
}

func (c *bitsetChunk) IsEmpty() bool {
	return *c == 0
}

func (c *bitsetChunk) Add(idx int) {
	AssertValidIndex(idx, bitsetChunkMaxBits)
	*c |= 1 << idx
}

func (c *bitsetChunk) Remove(idx int) {
	AssertValidIndex(idx, bitsetChunkMaxBits)
	*c &= ^(1 << idx)
}

func (c *bitsetChunk) All() iter.Seq[int] {
	localChunk := uint64(*c)
	return func(yield func(int) bool) {
		for localChunk != 0 {
			// get the position of the least significant bit by counting the number of trailing zeroes
			idx := bits.TrailingZeros64(localChunk)

			// clear the least significant bit
			localChunk &= localChunk - 1
			if !yield(idx) {
				return
			}
		}
	}
}
