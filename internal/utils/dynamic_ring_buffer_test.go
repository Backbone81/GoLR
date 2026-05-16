package utils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("DynamicRingBuffer", func() {
	It("should create an instance with the default capacity", func() {
		buffer := utils.NewDynamicRingBuffer[int]()
		Expect(buffer.Capacity()).To(Equal(1024))
	})

	It("should return the same elements on Remove() which were added with Add() before", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}

		for i := range 8 {
			Expect(buffer.Remove()).To(Equal(i))
		}
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))
	})

	It("should correctly handle wraparound", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		// add and remove some values to move into the middle of the buffer
		for i := range 5 {
			buffer.Add(i)
			buffer.Remove()
		}

		// now do the real test which will wrap around the end of the buffer
		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}

		for i := range 8 {
			Expect(buffer.Remove()).To(Equal(i))
		}
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))
	})

	It("should grow the buffer when his capacity is exceeded", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 12 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(12))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(15))

		for i := range 12 {
			Expect(buffer.Get(i)).To(Equal(i))
		}

		for i := range 12 {
			Expect(buffer.Remove()).To(Equal(i))
		}
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(15))
	})

	It("should always add 50% to the capacity when growing the buffer", func() {
		tests := []struct {
			items    int
			capacity int
		}{
			{7, 10},
			{11, 15},
			{16, 22},
			{23, 33},
			{34, 49},
		}
		for _, test := range tests {
			buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
			for i := range test.items {
				buffer.Add(i)
			}
			Expect(buffer.Length()).To(Equal(test.items))
			Expect(buffer.IsEmpty()).To(BeFalse())
			Expect(buffer.Capacity()).To(Equal(test.capacity))

			for i := range test.items {
				Expect(buffer.Get(i)).To(Equal(i))
			}

			for i := range test.items {
				Expect(buffer.Remove()).To(Equal(i))
			}
			Expect(buffer.Length()).To(Equal(0))
			Expect(buffer.IsEmpty()).To(BeTrue())
			Expect(buffer.Capacity()).To(Equal(test.capacity))
		}
	})

	It("should correctly remove all items when reset", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}

		buffer.Reset()
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))
	})

	It("should correctly resize the capacity when increased", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		buffer.Resize(20)
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(20))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}
	})

	It("should correctly resize the capacity when decreased", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](20)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(20))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(20))

		buffer.Resize(10)
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}
	})

	It("should not resize the capacity below the number of items stored", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		buffer.Resize(3)
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(8))

		for i := range 8 {
			Expect(buffer.Get(i)).To(Equal(i))
		}
	})

	It("should correctly handle RemoveN()", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 8 {
			buffer.Add(i)
		}
		Expect(buffer.Length()).To(Equal(8))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		buffer.RemoveN(4)
		Expect(buffer.Length()).To(Equal(4))
		Expect(buffer.IsEmpty()).To(BeFalse())
		Expect(buffer.Capacity()).To(Equal(10))

		for i := range 4 {
			Expect(buffer.Get(i)).To(Equal(4 + i))
		}

		for i := range 4 {
			Expect(buffer.Remove()).To(Equal(4 + i))
		}
		Expect(buffer.Length()).To(Equal(0))
		Expect(buffer.IsEmpty()).To(BeTrue())
		Expect(buffer.Capacity()).To(Equal(10))
	})

	It("should panic when calling Remove() on an empty buffer", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		Expect(func() { buffer.Remove() }).To(Panic())
	})

	It("should panic when calling RemoveN() with more items than available", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		for i := range 8 {
			buffer.Add(i)
		}
		Expect(func() { buffer.RemoveN(9) }).To(Panic())
	})

	It("should panic when calling Get() out of bounds", func() {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](10)
		for i := range 8 {
			buffer.Add(i)
		}
		Expect(func() { buffer.Get(-2) }).To(Panic())
		Expect(func() { buffer.Get(9) }).To(Panic())
	})
})

func BenchmarkDynamicRingBuffer(b *testing.B) {
	b.Run("Add() & Remove()", func(b *testing.B) {
		buffer := utils.NewDynamicRingBufferWithCapacity[int](b.N)
		b.ResetTimer()
		for i := range b.N {
			buffer.Add(i)
			buffer.Remove()
		}
	})
}
