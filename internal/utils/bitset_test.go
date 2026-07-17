package utils_test

import (
	"slices"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Bitset", func() {
	It("should correctly set bits", func() {
		var bitset utils.Bitset
		Expect(bitset.Contains(0)).To(BeFalse())
		Expect(bitset.Contains(1)).To(BeFalse())
		Expect(bitset.Contains(64 + 32)).To(BeFalse())
		Expect(bitset.Length()).To(Equal(0))
		Expect(bitset.IsEmpty()).To(BeTrue())

		bitset.Add(0)

		Expect(bitset.Contains(0)).To(BeTrue())
		Expect(bitset.Contains(1)).To(BeFalse())
		Expect(bitset.Contains(64 + 32)).To(BeFalse())
		Expect(bitset.Length()).To(Equal(1))
		Expect(bitset.IsEmpty()).To(BeFalse())

		bitset.Add(1)

		Expect(bitset.Contains(0)).To(BeTrue())
		Expect(bitset.Contains(1)).To(BeTrue())
		Expect(bitset.Contains(64 + 32)).To(BeFalse())
		Expect(bitset.Length()).To(Equal(2))
		Expect(bitset.IsEmpty()).To(BeFalse())

		bitset.Add(64 + 32)

		Expect(bitset.Contains(0)).To(BeTrue())
		Expect(bitset.Contains(1)).To(BeTrue())
		Expect(bitset.Contains(64 + 32)).To(BeTrue())
		Expect(bitset.Length()).To(Equal(3))
		Expect(bitset.IsEmpty()).To(BeFalse())
	})

	It("should report if adding a bit changed the bitset", func() {
		var bitset utils.Bitset

		Expect(bitset.Add(0)).To(BeTrue())
		Expect(bitset.Add(0)).To(BeFalse())

		// The bit lives in a chunk which does not exist yet, so the bitset has to grow for it.
		Expect(bitset.Add(64 + 32)).To(BeTrue())
		Expect(bitset.Add(64 + 32)).To(BeFalse())

		// A bit which was removed before is set again.
		bitset.Remove(0)
		Expect(bitset.Add(0)).To(BeTrue())
	})

	It("should correctly remove bits", func() {
		var bitset utils.Bitset
		bitset.Add(0)
		bitset.Add(1)
		bitset.Add(64 + 32)

		Expect(bitset.Contains(0)).To(BeTrue())
		Expect(bitset.Contains(1)).To(BeTrue())
		Expect(bitset.Contains(64 + 32)).To(BeTrue())
		Expect(bitset.Length()).To(Equal(3))
		Expect(bitset.IsEmpty()).To(BeFalse())

		bitset.Remove(0)

		Expect(bitset.Contains(0)).To(BeFalse())
		Expect(bitset.Contains(1)).To(BeTrue())
		Expect(bitset.Contains(64 + 32)).To(BeTrue())
		Expect(bitset.Length()).To(Equal(2))
		Expect(bitset.IsEmpty()).To(BeFalse())

		bitset.Remove(1)

		Expect(bitset.Contains(0)).To(BeFalse())
		Expect(bitset.Contains(1)).To(BeFalse())
		Expect(bitset.Contains(64 + 32)).To(BeTrue())
		Expect(bitset.Length()).To(Equal(1))
		Expect(bitset.IsEmpty()).To(BeFalse())

		bitset.Remove(64 + 32)

		Expect(bitset.Contains(0)).To(BeFalse())
		Expect(bitset.Contains(1)).To(BeFalse())
		Expect(bitset.Contains(64 + 32)).To(BeFalse())
		Expect(bitset.Length()).To(Equal(0))
		Expect(bitset.IsEmpty()).To(BeTrue())
	})

	It("should correctly iterate", func() {
		var bitset utils.Bitset
		bits := []int{3, 2, 64 + 32, 7, 4}
		for _, bit := range bits {
			bitset.Add(bit)
		}
		slices.Sort(bits)

		var iteration []int
		for bit := range bitset.All() {
			iteration = append(iteration, bit)
		}

		Expect(iteration).To(Equal(bits))
	})

	It("should correctly merge", func() {
		one := utils.NewBitset(3, 2, 64+32, 7, 4)
		two := utils.NewBitset(8, 7, 64+40, 130)
		Expect(one.Merge(&two)).To(BeTrue())
		Expect(one).To(Equal(utils.NewBitset(3, 2, 64+32, 7, 4, 8, 64+40, 130)))
	})

	DescribeTable("should correctly report if a merge changed the bitset",
		func(lhsBits []int, rhsBits []int, wantChanged bool) {
			lhsBitset := utils.NewBitset(lhsBits...)
			rhsBitset := utils.NewBitset(rhsBits...)
			wantLhsBitset := utils.NewBitset(append(slices.Clone(lhsBits), rhsBits...)...)

			Expect(lhsBitset.Merge(&rhsBitset)).To(Equal(wantChanged))
			Expect(lhsBitset.Equal(wantLhsBitset)).To(BeTrue())
		},
		// A merge only changes the bitset when the right-hand side holds a bit which is not set yet. Sharing bits with
		// the right-hand side is not a change.
		Entry("both empty", nil, nil, false),
		Entry("right-hand side empty", []int{1}, nil, false),
		Entry("left-hand side empty", nil, []int{1}, true),
		Entry("identical bits", []int{1, 2}, []int{1, 2}, false),
		Entry("right-hand side is a subset", []int{1, 2}, []int{2}, false),
		Entry("disjoint bits in the same chunk", []int{1}, []int{2}, true),
		Entry("overlapping bits with one new bit", []int{1, 2}, []int{2, 3}, true),
		Entry("right-hand side has a bit in a new chunk", []int{1}, []int{64 + 1}, true),
		Entry("right-hand side repeats a bit in a later chunk", []int{1, 64 + 1}, []int{64 + 1}, false),
		Entry("right-hand side skips over an empty chunk", []int{1}, []int{128 + 1}, true),
		Entry("left-hand side reaches further than the right-hand side", []int{1, 128 + 1}, []int{1}, false),
	)

	It("should correctly intersect", func() {
		one := utils.NewBitset(3, 2, 64+32, 7, 4)
		two := utils.NewBitset(8, 7, 64+40, 130, 4)
		Expect(one.Intersect(&two)).To(BeTrue())
		Expect(one.Equal(utils.NewBitset(7, 4))).To(BeTrue())
	})

	DescribeTable("should correctly report if an intersect changed the bitset",
		func(lhsBits []int, rhsBits []int, wantChanged bool, wantBits []int) {
			lhsBitset := utils.NewBitset(lhsBits...)
			rhsBitset := utils.NewBitset(rhsBits...)
			wantLhsBitset := utils.NewBitset(wantBits...)

			Expect(lhsBitset.Intersect(&rhsBitset)).To(Equal(wantChanged))
			Expect(lhsBitset.Equal(wantLhsBitset)).To(BeTrue())
		},
		// An intersect only changes the bitset when this side holds a bit which the other side does not hold.
		Entry("both empty", nil, nil, false, nil),
		Entry("right-hand side empty", []int{1}, nil, true, nil),
		Entry("left-hand side empty", nil, []int{1}, false, nil),
		Entry("identical bits", []int{1, 2}, []int{1, 2}, false, []int{1, 2}),
		Entry("right-hand side is a subset", []int{1, 2}, []int{2}, true, []int{2}),
		Entry("left-hand side is a subset", []int{2}, []int{1, 2}, false, []int{2}),
		Entry("disjoint bits in the same chunk", []int{1}, []int{2}, true, nil),
		Entry("overlapping bits with one new bit", []int{1, 2}, []int{2, 3}, true, []int{2}),
		Entry("left-hand side has a bit in a later chunk", []int{1, 64 + 1}, []int{1}, true, []int{1}),
		Entry("right-hand side reaches further than the left-hand side", []int{1}, []int{1, 64 + 1}, false, []int{1}),
	)

	It("should correctly report equality", func() {
		one := utils.NewBitset(3, 2, 64+32, 7, 4)
		two := utils.NewBitset(7, 4, 64+32, 2, 3)
		Expect(one.Equal(two)).To(BeTrue())

		two.Add(5)
		Expect(one.Equal(two)).To(BeFalse())

		two.Remove(5)
		Expect(one.Equal(two)).To(BeTrue())

		two.Add(200)
		Expect(one.Equal(two)).To(BeFalse())

		two.Remove(200)
		Expect(one.Equal(two)).To(BeTrue())
	})

	It("should correctly return the raw bytes", func() {
		var bitset utils.Bitset
		Expect(bitset.Bytes()).To(BeNil())

		bitset.Add(0)
		Expect(bitset.Bytes()).To(Equal([]byte{1, 0, 0, 0, 0, 0, 0, 0}))

		// Growing the bitset into a second chunk and clearing that chunk again must not change the bytes, as the
		// trailing empty chunk does not hold any bits.
		bitset.Add(64)
		Expect(bitset.Bytes()).To(HaveLen(16))
		bitset.Remove(64)
		Expect(bitset.Bytes()).To(Equal([]byte{1, 0, 0, 0, 0, 0, 0, 0}))

		// Bitsets which are equal return the same bytes, even when one of them holds more storage than the other.
		one := utils.NewBitset(3, 200)
		two := utils.NewBitset(3, 200)
		two.Remove(200)
		one.Remove(200)
		Expect(one.Equal(two)).To(BeTrue())
		Expect(one.Bytes()).To(Equal(two.Bytes()))
	})

	It("should correctly calculate hashes", func() {
		one := utils.NewBitset(3, 2, 64+32, 7, 4)
		two := utils.NewBitset(7, 4, 64+32, 2, 3)
		Expect(one.Hash()).To(Equal(two.Hash()))

		two.Add(5)
		Expect(one.Hash()).ToNot(Equal(two.Hash()))

		two.Remove(5)
		Expect(one.Hash()).To(Equal(two.Hash()))

		two.Add(200)
		Expect(one.Hash()).ToNot(Equal(two.Hash()))

		two.Remove(200)
		Expect(one.Hash()).To(Equal(two.Hash()))
	})

	It("should correctly compare", func() {
		one := utils.NewBitset(3, 2, 64+32, 7, 4)
		two := utils.NewBitset(7, 4, 64+32, 2, 3)
		Expect(one.Compare(two)).To(Equal(0))
		Expect(two.Compare(one)).To(Equal(0))

		Expect(one.Compare(one)).To(Equal(0)) //nolint:gocritic // Comparing a set to itself is the point of this assertion.

		two.Add(5)
		Expect(one.Compare(two)).ToNot(Equal(0))
		Expect(one.Compare(two)).To(Equal(-two.Compare(one)))

		two.Remove(5)
		two.Add(200)
		two.Remove(200)
		Expect(one.Equal(two)).To(BeTrue())
		Expect(one.Compare(two)).To(Equal(0))

		three := utils.NewBitset(3, 2, 64+32, 7, 4, 200)
		Expect(one.Equal(three)).To(BeFalse())
		Expect(one.Compare(three)).ToNot(Equal(0))
		Expect(one.Compare(three)).To(Equal(-three.Compare(one)))
	})
})

func BenchmarkBitset(b *testing.B) {
	b.Run("Add()", func(b *testing.B) {
		for range b.N {
			var bitset utils.Bitset
			for bit := range 256 {
				bitset.Add(bit)
			}
		}
	})

	b.Run("Add() and Remove()", func(b *testing.B) {
		for range b.N {
			var bitset utils.Bitset
			for bit := range 256 {
				bitset.Add(bit)
			}
			for bit := range 256 {
				bitset.Remove(bit)
			}
		}
	})
}
