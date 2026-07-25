package utils_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("OrderedSet", func() {
	It("should return the index of a value", func() {
		orderedSet := utils.NewOrderedSet[int](5, 1, 3)

		index, found := orderedSet.IndexOf(1)
		Expect(found).To(BeTrue())
		Expect(index).To(Equal(0))

		index, found = orderedSet.IndexOf(3)
		Expect(found).To(BeTrue())
		Expect(index).To(Equal(1))

		index, found = orderedSet.IndexOf(5)
		Expect(found).To(BeTrue())
		Expect(index).To(Equal(2))

		_, found = orderedSet.IndexOf(4)
		Expect(found).To(BeFalse())
	})

	It("should return the index of a value matching the iteration order", func() {
		orderedSet := utils.NewOrderedSet[int](5, 1, 3)

		for expectedIndex, value := range orderedSet.All() {
			index, found := orderedSet.IndexOf(value)
			Expect(found).To(BeTrue())
			Expect(index).To(Equal(expectedIndex))
		}
	})

	It("should return no index for a value of an empty ordered set", func() {
		var orderedSet utils.OrderedSet[int]

		_, found := orderedSet.IndexOf(1)
		Expect(found).To(BeFalse())
	})

	It("should return the lower bound of a value", func() {
		orderedSet := utils.NewOrderedSet[int](1, 3, 5)

		Expect(orderedSet.LowerBound(0)).To(Equal(0))
		Expect(orderedSet.LowerBound(1)).To(Equal(0))
		Expect(orderedSet.LowerBound(2)).To(Equal(1))
		Expect(orderedSet.LowerBound(3)).To(Equal(1))
		Expect(orderedSet.LowerBound(4)).To(Equal(2))
		Expect(orderedSet.LowerBound(5)).To(Equal(2))
	})

	It("should return the length as the lower bound of a value greater than all values", func() {
		orderedSet := utils.NewOrderedSet[int](1, 3, 5)

		Expect(orderedSet.LowerBound(6)).To(Equal(orderedSet.Length()))
	})

	It("should return the lower bound of a value of an empty ordered set", func() {
		var orderedSet utils.OrderedSet[int]

		Expect(orderedSet.LowerBound(1)).To(Equal(0))
	})
})

func BenchmarkOrderedSet_Add(b *testing.B) {
	for values := 2; values <= 64; values *= 2 {
		b.Run(fmt.Sprintf("Adding %d values ascending", values), func(b *testing.B) {
			for range b.N {
				orderedSet := utils.NewOrderedSet[int]()
				for i := range values {
					orderedSet.Add(i)
				}
			}
		})
	}

	for values := 2; values <= 64; values *= 2 {
		b.Run(fmt.Sprintf("Adding %d values descending", values), func(b *testing.B) {
			for range b.N {
				orderedSet := utils.NewOrderedSet[int]()
				for i := values - 1; 0 <= i; i-- {
					orderedSet.Add(i)
				}
			}
		})
	}
}

func BenchmarkOrderedSet_Hash(b *testing.B) {
	for values := 2; values <= 64; values *= 2 {
		b.Run(fmt.Sprintf("Hashing %d values", values), func(b *testing.B) {
			orderedSet := utils.NewOrderedSet[int]()
			for i := range values {
				orderedSet.Add(i)
			}
			for range b.N {
				orderedSet.Hash()
			}
		})
	}
}
