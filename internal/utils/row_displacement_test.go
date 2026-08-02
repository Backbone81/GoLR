package utils_test

import (
	"math/rand/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/utils"
)

// noEntry is the value the test tables use for a cell without an entry.
const noEntry = -1

var _ = Describe("RowDisplacement", func() {
	It("should provide an empty table for no rows", func() {
		result := utils.NewRowDisplacement(nil, noEntry)
		Expect(result.Base).To(BeEmpty())
		Expect(result.Next).To(BeEmpty())
		Expect(result.Check).To(BeEmpty())
	})

	It("should return the entries of a single row", func() {
		rows := [][]int{
			{noEntry, 10, noEntry, 30},
		}
		result := expectDecodeEquivalent(rows, noEntry)
		Expect(result.Lookup(0, 1)).To(Equal(10))
		Expect(result.Lookup(0, 3)).To(Equal(30))
	})

	It("should return the empty value for a column without an entry", func() {
		rows := [][]int{
			{noEntry, 10, noEntry, 30},
		}
		result := expectDecodeEquivalent(rows, noEntry)
		Expect(result.Lookup(0, 0)).To(Equal(noEntry))
		Expect(result.Lookup(0, 2)).To(Equal(noEntry))
	})

	It("should make the entries of one row fall into the holes of another", func() {
		rows := [][]int{
			{1, 2, 3, noEntry, noEntry},
			{noEntry, noEntry, noEntry, 4, 5},
		}
		result := expectDecodeEquivalent(rows, noEntry)

		// The entries of the second row land in the holes the first one leaves, so the two rows need six cells
		// instead of the ten the uncompressed table needs. They can not share a displacement, which is why the
		// second row starts one cell further along than it would have to.
		Expect(result.Next).To(HaveLen(6))
	})

	It("should give identical rows the same displacement", func() {
		rows := [][]int{
			{1, noEntry, 3},
			{noEntry, 2, noEntry},
			{1, noEntry, 3},
		}
		result := expectDecodeEquivalent(rows, noEntry)
		Expect(result.Base[2]).To(Equal(result.Base[0]))
		Expect(result.Base[1]).ToNot(Equal(result.Base[0]))
	})

	It("should give rows without any entry the same displacement", func() {
		rows := [][]int{
			{noEntry, noEntry},
			{1, 2},
			{noEntry, noEntry},
		}
		result := expectDecodeEquivalent(rows, noEntry)
		Expect(result.Base[2]).To(Equal(result.Base[0]))
	})

	It("should keep rows apart which have their entries in the same columns", func() {
		rows := [][]int{
			{1, noEntry},
			{2, noEntry},
		}
		expectDecodeEquivalent(rows, noEntry)
	})

	It("should support an empty value other than minus one", func() {
		rows := [][]int{
			{0, 10, 0},
			{20, 0, 30},
		}
		expectDecodeEquivalent(rows, 0)
	})

	It("should support rows of different width", func() {
		rows := [][]int{
			{1, noEntry, 3, noEntry, 5},
			{noEntry, 2},
			{},
		}
		expectDecodeEquivalent(rows, noEntry)
	})

	It("should support a table without any entry", func() {
		rows := [][]int{
			{noEntry, noEntry},
			{noEntry, noEntry},
		}
		expectDecodeEquivalent(rows, noEntry)
	})

	It("should support a table without any hole", func() {
		rows := [][]int{
			{1, 2, 3},
			{4, 5, 6},
		}
		result := expectDecodeEquivalent(rows, noEntry)

		// Neither row has a hole for the other one to use, so the packed table can not be smaller than the
		// dense one.
		Expect(result.Next).To(HaveLen(6))
	})

	DescribeTable("should return the same entries as the uncompressed table",
		func(rowCount int, width int, density float64) {
			random := rand.New(rand.NewPCG(uint64(rowCount), uint64(width)))
			rows := randomRows(random, rowCount, width, density)
			expectDecodeEquivalent(rows, noEntry)
		},
		Entry("for a table which is almost empty", 200, 64, 0.02),
		Entry("for a sparse table", 200, 64, 0.15),
		Entry("for a table which is half filled", 200, 64, 0.5),
		Entry("for a table which is almost full", 200, 64, 0.95),
		Entry("for a table with few but wide rows", 20, 500, 0.1),
		Entry("for a table with many but narrow rows", 1000, 8, 0.2),
	)

	It("should need fewer cells than the uncompressed table for a sparse table", func() {
		random := rand.New(rand.NewPCG(1, 2))
		rows := randomRows(random, 200, 64, 0.1)

		result := expectDecodeEquivalent(rows, noEntry)
		Expect(len(result.Next)).To(BeNumerically("<", 200*64/4))
	})
})

// expectDecodeEquivalent compresses the given rows and verifies that looking up every cell returns what the
// uncompressed rows hold, which is the property the whole compression rests on. It returns the compressed table so
// that a test can make additional assertions about it.
func expectDecodeEquivalent(rows [][]int, emptyValue int) utils.RowDisplacement {
	GinkgoHelper()

	result := utils.NewRowDisplacement(rows, emptyValue)
	Expect(result.Base).To(HaveLen(len(rows)))
	Expect(result.Check).To(HaveLen(len(result.Next)))
	Expect(result.EmptyValue).To(Equal(emptyValue))

	for rowIdx, row := range rows {
		for colIdx, value := range row {
			Expect(result.Lookup(rowIdx, colIdx)).To(
				Equal(value),
				"row %d column %d", rowIdx, colIdx,
			)
		}
	}
	return result
}

// randomRows builds a table of the given size in which the given fraction of the cells holds an entry.
func randomRows(random *rand.Rand, rowCount int, width int, density float64) [][]int {
	rows := make([][]int, rowCount)
	for rowIdx := range rows {
		rows[rowIdx] = make([]int, width)
		for colIdx := range rows[rowIdx] {
			if random.Float64() < density {
				rows[rowIdx][colIdx] = random.IntN(1000)
			} else {
				rows[rowIdx][colIdx] = noEntry
			}
		}
	}
	return rows
}
