package utils_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Deque", func() {
	It("should report the zero value as empty", func() {
		var deque utils.Deque[int]
		Expect(deque.Len()).To(Equal(0))
		Expect(deque.AppendAll(nil)).To(BeEmpty())
	})

	It("should keep the values pushed at the back in order", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2).PushBack(3)
		Expect(deque.Len()).To(Equal(3))
		Expect(deque.First()).To(Equal(1))
		Expect(deque.Last()).To(Equal(3))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1, 2, 3}))
	})

	It("should keep the values pushed at the front in reverse order", func() {
		deque := utils.Deque[int]{}.PushFront(1).PushFront(2).PushFront(3)
		Expect(deque.Len()).To(Equal(3))
		Expect(deque.First()).To(Equal(3))
		Expect(deque.Last()).To(Equal(1))
		Expect(deque.AppendAll(nil)).To(Equal([]int{3, 2, 1}))
	})

	It("should combine values pushed at both ends", func() {
		deque := utils.Deque[int]{}.PushBack(3).PushFront(2).PushBack(4).PushFront(1)
		Expect(deque.First()).To(Equal(1))
		Expect(deque.Last()).To(Equal(4))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1, 2, 3, 4}))
	})

	It("should keep the first value at hand while values are pushed and dropped at the back", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2)
		Expect(deque.First()).To(Equal(1))
		deque = deque.DropBack(1).PushBack(3).PushBack(4).DropBack(2)
		Expect(deque.First()).To(Equal(1))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1}))
	})

	It("should keep the first value at hand when popping rebuilds the sequence", func() {
		deque := utils.Deque[int]{}.PushBack(3).PushFront(2).PushFront(1)
		deque, _ = deque.PopBack(2, nil)
		Expect(deque.First()).To(Equal(1))
		deque = deque.PushBack(4)
		Expect(deque.First()).To(Equal(1))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1, 4}))
	})

	It("should append to the values already in the buffer", func() {
		deque := utils.Deque[int]{}.PushBack(2).PushFront(1)
		Expect(deque.AppendAll([]int{0})).To(Equal([]int{0, 1, 2}))
	})

	It("should drop values from the back", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2).PushBack(3).DropBack(2)
		Expect(deque.Len()).To(Equal(1))
		Expect(deque.Last()).To(Equal(1))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1}))
	})

	It("should pop values from the back in order", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2).PushBack(3)
		deque, popped := deque.PopBack(2, []int{0})
		Expect(popped).To(Equal([]int{0, 2, 3}))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1}))
	})

	It("should pop values which were pushed at the front", func() {
		deque := utils.Deque[int]{}.PushBack(3).PushFront(2).PushFront(1)
		deque, popped := deque.PopBack(2, nil)
		Expect(popped).To(Equal([]int{2, 3}))
		Expect(deque.Len()).To(Equal(1))
		Expect(deque.Last()).To(Equal(1))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1}))
	})

	It("should keep the last value at hand when dropping every value pushed at the back", func() {
		deque := utils.Deque[int]{}.PushBack(3).PushFront(2).PushFront(1).DropBack(1)
		Expect(deque.First()).To(Equal(1))
		Expect(deque.Last()).To(Equal(2))
		Expect(deque.AppendAll(nil)).To(Equal([]int{1, 2}))
	})

	It("should become empty when dropping every value", func() {
		deque := utils.Deque[int]{}.PushBack(2).PushFront(1).DropBack(2)
		Expect(deque.Len()).To(Equal(0))
		Expect(deque.AppendAll(nil)).To(BeEmpty())

		deque = deque.PushFront(3)
		Expect(deque.First()).To(Equal(3))
		Expect(deque.Last()).To(Equal(3))

		deque = deque.DropBack(1).PushBack(4)
		Expect(deque.First()).To(Equal(4))
		Expect(deque.Last()).To(Equal(4))
	})

	It("should leave the sequence it was called on unchanged", func() {
		original := utils.Deque[int]{}.PushBack(2).PushFront(1).PushBack(3)
		_ = original.PushBack(4)
		_ = original.PushFront(0)
		_ = original.DropBack(3)
		_, _ = original.PopBack(2, nil)
		Expect(original.Len()).To(Equal(3))
		Expect(original.First()).To(Equal(1))
		Expect(original.Last()).To(Equal(3))
		Expect(original.AppendAll(nil)).To(Equal([]int{1, 2, 3}))
	})

	It("should keep sequences derived from the same sequence apart", func() {
		shared := utils.Deque[int]{}.PushBack(1).PushBack(2)
		first := shared.PushBack(3)
		second := shared.DropBack(1).PushBack(4)
		Expect(first.AppendAll(nil)).To(Equal([]int{1, 2, 3}))
		Expect(second.AppendAll(nil)).To(Equal([]int{1, 4}))
		Expect(shared.AppendAll(nil)).To(Equal([]int{1, 2}))
	})

	It("should report the values it holds", func() {
		deque := utils.Deque[int]{}.PushBack(2).PushFront(1)
		Expect(utils.DequeContains(deque, 1)).To(BeTrue())
		Expect(utils.DequeContains(deque, 2)).To(BeTrue())
		Expect(utils.DequeContains(deque, 3)).To(BeFalse())
		Expect(utils.DequeContains(utils.Deque[int]{}, 1)).To(BeFalse())
	})

	It("should hash equal sequences alike no matter how they were built", func() {
		pushedBack := utils.Deque[int]{}.PushBack(1).PushBack(2).PushBack(3)
		pushedFront := utils.Deque[int]{}.PushFront(3).PushFront(2).PushFront(1)
		mixed := utils.Deque[int]{}.PushBack(2).PushFront(1).PushBack(3)
		rebuilt, _ := utils.Deque[int]{}.PushBack(2).PushFront(1).PushBack(3).PushBack(4).PopBack(1, nil)
		Expect(hashOf(pushedFront)).To(Equal(hashOf(pushedBack)))
		Expect(hashOf(mixed)).To(Equal(hashOf(pushedBack)))
		Expect(hashOf(rebuilt)).To(Equal(hashOf(pushedBack)))
	})

	It("should hash different sequences differently", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2)
		Expect(hashOf(utils.Deque[int]{}.PushBack(2).PushBack(1))).ToNot(Equal(hashOf(deque)))
		Expect(hashOf(deque.PushBack(3))).ToNot(Equal(hashOf(deque)))
		Expect(hashOf(deque.DropBack(1))).ToNot(Equal(hashOf(deque)))
		Expect(hashOf(utils.Deque[int]{})).ToNot(Equal(hashOf(deque)))
	})

	It("should hash alike with a reused buffer and return the buffer", func() {
		deque := utils.Deque[int]{}.PushBack(1).PushBack(2)
		hash, buffer := deque.Hash(make([]int, 0, 8))
		Expect(hash).To(Equal(hashOf(deque)))
		Expect(buffer).To(HaveCap(8))

		hash, _ = deque.PushBack(3).Hash(buffer)
		Expect(hash).To(Equal(hashOf(deque.PushBack(3))))
	})

	It("should panic when dropping more values than it holds", func() {
		if !utils.EnableDebugAssertions {
			Skip("debug assertions are disabled")
		}
		deque := utils.Deque[int]{}.PushBack(1)
		Expect(func() { deque.DropBack(2) }).To(Panic())
	})

	It("should hold the same values as a slice after the same random operations", func() {
		random := rand.New(rand.NewSource(GinkgoRandomSeed()))
		for range 100 {
			var values []int
			var deque utils.Deque[int]
			for range random.Intn(10) {
				value := random.Int()
				values = append(values, value)
				deque = deque.PushBack(value)
			}

			for range 100 {
				switch random.Intn(4) {
				case 0:
					value := random.Int()
					values = append(values, value)
					deque = deque.PushBack(value)
				case 1:
					value := random.Int()
					values = append([]int{value}, values...)
					deque = deque.PushFront(value)
				case 2:
					count := random.Intn(len(values) + 1)
					var popped []int
					deque, popped = deque.PopBack(count, nil)
					Expect(popped).To(HaveExactElements(values[len(values)-count:]))
					values = values[:len(values)-count]
				case 3:
					count := random.Intn(len(values) + 1)
					deque = deque.DropBack(count)
					values = values[:len(values)-count]
				}

				Expect(deque.Len()).To(Equal(len(values)))
				Expect(deque.AppendAll(nil)).To(HaveExactElements(values))

				var rebuilt utils.Deque[int]
				for _, value := range values {
					rebuilt = rebuilt.PushBack(value)
				}
				Expect(hashOf(deque)).To(Equal(hashOf(rebuilt)))
				if len(values) > 0 {
					Expect(deque.First()).To(Equal(values[0]))
					Expect(deque.Last()).To(Equal(values[len(values)-1]))
				}
			}
		}
	})
})

// hashOf returns the hash of the deque.
func hashOf(d utils.Deque[int]) uint64 {
	hash, _ := d.Hash(nil)
	return hash
}
