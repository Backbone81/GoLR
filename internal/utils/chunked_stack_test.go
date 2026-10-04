package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

var _ = Describe("ChunkedStack", func() {
	It("should report a new stack as empty", func() {
		var stack utils.ChunkedStack[int]
		Expect(stack.Size()).To(Equal(0))
		Expect(stack.IsEmpty()).To(BeTrue())
	})

	It("should return the pushed value from the top", func() {
		var stack utils.ChunkedStack[int]
		stack.Push(42)
		Expect(stack.Size()).To(Equal(1))
		Expect(stack.IsEmpty()).To(BeFalse())
		Expect(stack.Top()).To(Equal(42))
	})

	It("should return values in last in first out order", func() {
		var stack utils.ChunkedStack[int]
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

	It("should return values in last in first out order across many chunks", func() {
		var stack utils.ChunkedStack[int]
		for value := range 5000 {
			stack.Push(value)
		}
		Expect(stack.Size()).To(Equal(5000))
		for value := 4999; value >= 0; value-- {
			Expect(stack.Top()).To(Equal(value))
			stack.Pop()
		}
		Expect(stack.IsEmpty()).To(BeTrue())
	})

	It("should behave like a stack when pushes and pops interleave across chunks", func() {
		var stack utils.ChunkedStack[int]
		var reference utils.Stack[int]
		// Two pushes and a pop grow the stack, then a pop, a push and a pop shrink it again.
		for _, pattern := range [][]bool{{true, true, false}, {false, true, false}} {
			for step := range 9000 {
				if pattern[step%len(pattern)] {
					stack.Push(step)
					reference.Push(step)
				} else {
					stack.Pop()
					reference.Pop()
				}
				Expect(stack.Size()).To(Equal(reference.Size()))
				if !reference.IsEmpty() {
					Expect(stack.Top()).To(Equal(reference.Top()))
				}
			}
		}
		Expect(stack.IsEmpty()).To(BeTrue())
	})

	It("should be reusable after being emptied", func() {
		var stack utils.ChunkedStack[int]
		for value := range 3000 {
			stack.Push(value)
		}
		for range 3000 {
			stack.Pop()
		}
		Expect(stack.IsEmpty()).To(BeTrue())

		stack.Push(2)
		Expect(stack.Size()).To(Equal(1))
		Expect(stack.Top()).To(Equal(2))
	})

	It("should work with other element types", func() {
		var stack utils.ChunkedStack[string]
		stack.Push("a")
		stack.Push("b")
		Expect(stack.Top()).To(Equal("b"))
		stack.Pop()
		Expect(stack.Top()).To(Equal("a"))
	})
})
