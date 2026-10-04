package counterexample_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/counterexample"
)

var _ = Describe("ProductConfigurationQueue", func() {
	It("should be empty when created", func() {
		queue := counterexample.NewProductConfigurationQueue()
		Expect(queue.IsEmpty()).To(BeTrue())
	})

	It("should remove the configurations cheapest first", func() {
		queue := counterexample.NewProductConfigurationQueue()
		for _, cost := range []int32{3, 0, 2, 1, 2} {
			queue.Add(counterexample.PendingConfiguration{Cost: cost})
		}
		Expect(queue.Len()).To(Equal(5))
		Expect(removeAllCosts(&queue)).To(Equal([]int32{0, 1, 2, 2, 3}))
		Expect(queue.IsEmpty()).To(BeTrue())
	})

	It("should keep the order while the costs rise far beyond the initial buckets", func() {
		queue := counterexample.NewProductConfigurationQueue()
		for cost := range int32(3) {
			queue.Add(counterexample.PendingConfiguration{Cost: cost})
		}
		// The costs of the actions of the search.
		actionCosts := []int32{1, 4, 20}
		var removedCosts []int32
		for i := range 1000 {
			c := queue.Remove()
			removedCosts = append(removedCosts, c.Cost)
			queue.Add(counterexample.PendingConfiguration{Cost: c.Cost + actionCosts[i%len(actionCosts)]})
		}
		Expect(slices.IsSorted(removedCosts)).To(BeTrue())
		Expect(removedCosts[len(removedCosts)-1]).To(BeNumerically(">", 100))
	})

	It("should add configurations far more expensive than the cheapest one", func() {
		queue := counterexample.NewProductConfigurationQueue()
		queue.Add(counterexample.PendingConfiguration{})
		Expect(queue.Remove().Cost).To(Equal(int32(0)))
		for _, cost := range []int32{150, 5, 100} {
			queue.Add(counterexample.PendingConfiguration{Cost: cost})
		}
		Expect(removeAllCosts(&queue)).To(Equal([]int32{5, 100, 150}))
	})
})

// removeAllCosts removes all configurations from the queue and returns their costs in the order of removal.
func removeAllCosts(queue *counterexample.ProductConfigurationQueue) []int32 {
	var result []int32
	for !queue.IsEmpty() {
		result = append(result, queue.Remove().Cost)
	}
	return result
}
