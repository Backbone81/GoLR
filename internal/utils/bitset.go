package utils

import (
	"encoding/json"
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
// The return value reports if the bit was set which was not set before.
func (b *Bitset) Add(idx int) bool {
	AssertValidIndex(idx, math.MaxInt)
	chunkIdx := idx / bitsetChunkMaxBits
	bitIdx := idx % bitsetChunkMaxBits
	if len(b.chunks) <= chunkIdx {
		newChunks := make([]bitsetChunk, chunkIdx+1)
		copy(newChunks, b.chunks)
		b.chunks = newChunks
	}
	return b.chunks[chunkIdx].Add(bitIdx)
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

// Clone returns a copy of the bitset which shares no storage with the original. A plain copy of a bitset keeps
// referencing the chunks of the original, so setting or removing a bit on the copy would change the original as well.
// Clone is what you want when the original must stay untouched.
func (b *Bitset) Clone() Bitset {
	return Bitset{
		chunks: slices.Clone(b.chunks),
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

// Merge adds all the bits set in the other bitset to the current one. The return value reports if a bit was set which
// was not set before.
func (b *Bitset) Merge(other *Bitset) bool {
	changed := false
	commonChunks := min(len(b.chunks), len(other.chunks))
	for i := range commonChunks {
		// The chunk changes when the other chunk holds a bit which this chunk does not hold yet.
		changed = changed || other.chunks[i]&^b.chunks[i] != 0
		b.chunks[i] |= other.chunks[i]
	}
	for i := range other.chunks[commonChunks:] {
		changed = changed || other.chunks[commonChunks+i] != 0
	}
	b.chunks = append(b.chunks, other.chunks[commonChunks:]...)
	return changed
}

// Intersect removes every bit from the bitset which is not also set in the other bitset, so the bitset is reduced to
// the intersection of both. Bits which are set in the other bitset but not in this one are never added. The return
// value reports if a bit was removed.
func (b *Bitset) Intersect(other *Bitset) bool {
	changed := false
	for i := range b.chunks {
		var otherChunk bitsetChunk
		if i < len(other.chunks) {
			otherChunk = other.chunks[i]
		}
		// A bit is removed when this chunk holds a bit which the other chunk does not hold.
		changed = changed || b.chunks[i]&^otherChunk != 0
		b.chunks[i] &= otherChunk
	}
	return changed
}

// Equal reports if this bitset is equal to the other bitset. The bitsets can be of different size and still be equal
// as long as the set bits are located at locations which both bitsets share.
func (b *Bitset) Equal(other Bitset) bool {
	// make sure the same chunks are equal
	for i := range min(len(b.chunks), len(other.chunks)) {
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

// Compare returns a negative number, zero, or a positive number reporting whether this bitset sorts before, equal to,
// or after the other bitset. The ordering is consistent with Equal: bitsets that differ only in trailing empty chunks
// compare as equal.
func (b *Bitset) Compare(other Bitset) int {
	for i := range max(len(b.chunks), len(other.chunks)) {
		var left, right bitsetChunk
		if i < len(b.chunks) {
			left = b.chunks[i]
		}
		if i < len(other.chunks) {
			right = other.chunks[i]
		}
		if left != right {
			if left < right {
				return -1
			}
			return 1
		}
	}
	return 0
}

// MarshalJSON implements the json.Marshaler interface.
func (b Bitset) MarshalJSON() ([]byte, error) {
	idxs := make([]int, 0, b.Length())
	for idx := range b.All() {
		idxs = append(idxs, idx)
	}
	if len(idxs) == 0 {
		return json.Marshal(nil)
	}
	return json.Marshal(idxs)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *Bitset) UnmarshalJSON(data []byte) error {
	if slices.Equal(data, []byte("null")) {
		return nil
	}
	var idxs []int
	err := json.Unmarshal(data, &idxs)
	if err != nil {
		return err
	}
	for _, idx := range idxs {
		b.Add(idx)
	}
	return nil
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

// String returns a string representation.
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

// Bytes returns the raw bytes of the chunks which hold the bits of the bitset. Trailing empty chunks are not part of
// the result, so two bitsets which are Equal always return the same bytes, no matter how much storage they hold. An
// empty bitset returns nil.
//
// The result aliases the storage of the bitset. It must not be modified and it is only valid until the bitset is
// modified. This is meant for hashing and serializing a bitset without copying its bits.
func (b *Bitset) Bytes() []byte {
	chunkCount := len(b.chunks)
	for ; chunkCount > 0 && b.chunks[chunkCount-1].IsEmpty(); chunkCount-- {
		// We remove empty chunks at the end, as we do not want them to contribute to the result. Otherwise, two bitsets
		// with the same bits set but with different sizes would lead to different bytes.
	}
	if chunkCount == 0 {
		return nil
	}
	chunksByteSize := chunkCount * int(unsafe.Sizeof(b.chunks[0]))

	//nolint:gosec // unsafe is required for better performance
	return unsafe.Slice((*byte)(unsafe.Pointer(&b.chunks[0])), chunksByteSize)
}

// Hash returns a hash value over all chunks. All chunks except for trailing empty chunks contribute to the hash.
func (b *Bitset) Hash() uint64 {
	hash := fnv.New64a()
	if _, err := hash.Write(b.Bytes()); err != nil {
		panic(err)
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

func (c *bitsetChunk) Add(idx int) bool {
	AssertValidIndex(idx, bitsetChunkMaxBits)
	mask := bitsetChunk(1) << idx
	changed := *c&mask == 0
	*c |= mask
	return changed
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
