package counterexample_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/counterexample"
)

var _ = Describe("ProductConfiguration", func() {
	It("should add the cost to the successor and leave the original unchanged", func() {
		c := counterexample.ProductConfiguration{Cost: 3}
		successor := c.Successor(4)
		Expect(successor.Cost).To(Equal(7))
		Expect(c.Cost).To(Equal(3))
	})

	It("should hash separate configurations of equal contents equally", func() {
		Expect(hashOf(newProductConfiguration(1, 2, false))).To(Equal(hashOf(newProductConfiguration(1, 2, false))))
	})

	It("should hash configurations which differ in the shifted terminal differently", func() {
		Expect(hashOf(newProductConfiguration(1, 2, false))).ToNot(Equal(hashOf(newProductConfiguration(1, 2, true))))
	})

	It("should hash configurations with swapped parsers differently", func() {
		Expect(hashOf(newProductConfiguration(1, 2, false))).ToNot(Equal(hashOf(newProductConfiguration(2, 1, false))))
	})

	It("should hash configurations which differ only in the cost equally", func() {
		c := newProductConfiguration(1, 2, false)
		Expect(hashOf(c.Successor(5))).To(Equal(hashOf(c)))
	})
})

// newProductConfiguration returns a ProductConfiguration whose parsers start at the conflict items.
func newProductConfiguration(
	firstItemIdx int,
	secondItemIdx int,
	terminalShifted bool,
) *counterexample.ProductConfiguration {
	return &counterexample.ProductConfiguration{
		Parsers: [2]counterexample.SimulatedParser{
			counterexample.NewSimulatedParser(firstItemIdx),
			counterexample.NewSimulatedParser(secondItemIdx),
		},
		TerminalShifted: terminalShifted,
	}
}

// hashOf returns the hash of the ProductConfiguration.
func hashOf(c *counterexample.ProductConfiguration) uint64 {
	hash, _ := c.Hash(nil)
	return hash
}
