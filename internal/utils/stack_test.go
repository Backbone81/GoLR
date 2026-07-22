package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("Stack", func() {
	It("should report a new stack as empty", func() {
		var stack utils.Stack[int]
		Expect(stack.Size()).To(Equal(0))
		Expect(stack.IsEmpty()).To(BeTrue())
	})

	It("should return the pushed value from the top", func() {
		var stack utils.Stack[int]
		stack.Push(42)
		Expect(stack.Size()).To(Equal(1))
		Expect(stack.IsEmpty()).To(BeFalse())
		Expect(stack.Top()).To(Equal(42))
	})

	It("should return values in last in first out order", func() {
		var stack utils.Stack[int]
		stack.Push(1)
		stack.Push(2)
		stack.Push(3)
		Expect(stack.Size()).To(Equal(3))

		Expect(stack.Top()).To(Equal(3))
		stack.Pop()
		Expect(stack.Top()).To(Equal(2))
		stack.Pop()
		Expect(stack.Top()).To(Equal(1))
		stack.Pop()

		Expect(stack.IsEmpty()).To(BeTrue())
		Expect(stack.Size()).To(Equal(0))
	})

	It("should not change the value on top when calling Top repeatedly", func() {
		var stack utils.Stack[int]
		stack.Push(7)
		Expect(stack.Top()).To(Equal(7))
		Expect(stack.Top()).To(Equal(7))
		Expect(stack.Size()).To(Equal(1))
	})

	It("should handle interleaved pushes and pops", func() {
		var stack utils.Stack[int]
		stack.Push(1)
		stack.Push(2)
		stack.Pop()
		Expect(stack.Top()).To(Equal(1))
		stack.Push(3)
		Expect(stack.Top()).To(Equal(3))
		Expect(stack.Size()).To(Equal(2))
	})

	It("should be reusable after being emptied", func() {
		var stack utils.Stack[int]
		stack.Push(1)
		stack.Pop()
		Expect(stack.IsEmpty()).To(BeTrue())

		stack.Push(2)
		Expect(stack.IsEmpty()).To(BeFalse())
		Expect(stack.Top()).To(Equal(2))
	})

	It("should work with other element types", func() {
		var stack utils.Stack[string]
		stack.Push("a")
		stack.Push("b")
		Expect(stack.Top()).To(Equal("b"))
		stack.Pop()
		Expect(stack.Top()).To(Equal("a"))
	})
})
