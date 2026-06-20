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
		one.Merge(&two)
		Expect(one).To(Equal(utils.NewBitset(3, 2, 64+32, 7, 4, 8, 64+40, 130)))
	})

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
