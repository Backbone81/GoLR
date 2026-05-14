package utils_test

import (
	"fmt"
	"testing"

	"golr/internal/utils"
)

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
