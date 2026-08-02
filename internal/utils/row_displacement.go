package utils

import (
	"encoding/binary"
	"slices"
	"strings"
)

// NoColumn is the entry used in RowDisplacement.Check for a cell which does not belong to any row. Column indices are
// never negative, so this can not be confused with a cell holding a valid column.
const NoColumn = -1

// RowDisplacement stores a sparse table as a single shared array, by displacing every row by an amount which makes its
// occupied cells fall into the holes of the rows placed before it.
//
// This is the row displacement method of "Storing a Sparse Table" by Tarjan and Yao, with the displacements chosen by
// the first-fit-decreasing algorithm of its appendix. The paper only computes the displacements themselves. The Check
// array added here is what makes a lookup safe, because it tells whether the cell a row lands on really belongs to that
// row.
//
// The zero value is the displacement of a table without any rows.
type RowDisplacement struct {
	// Base holds the displacement of every row, indexed by the row index.
	Base []int

	// Next holds the entries of all rows packed into a single array. The entry of a row at a column lives at
	// Base[row] + column. Cells which no row occupies hold EmptyValue.
	Next []int

	// Check holds the column every cell of Next belongs to, or NoColumn for a cell which no row occupies. A lookup
	// is only valid when the cell it lands on holds the column which was looked up.
	Check []int

	// EmptyValue is the value the rows use for a cell without an entry, and the value a lookup returns for a
	// column a row has no entry for.
	EmptyValue int
}

// NewRowDisplacement packs the given rows into a single array. A cell of a row holding emptyValue is a hole which the
// rows placed after it may use.
//
// Rows with identical content share a single displacement, which is the additional trick for LR tables mentioned in
// section 5 of "Storing a Sparse Table" by Tarjan and Yao. Sharing is only sound because rows which are not identical
// are given distinct displacements: two distinct rows on the same displacement would make one of them accept the cells
// of the other, as Check only records the column of a cell and not which row put it there.
func NewRowDisplacement(rows [][]int, emptyValue int) RowDisplacement {
	builder := newRowDisplacementBuilder(rows, emptyValue)
	for _, rowIdx := range builder.placementOrder() {
		builder.placeRow(rowIdx)
	}
	return builder.finish()
}

// Lookup returns the entry of the given row at the given column, or EmptyValue if the row has no entry there. This is
// the access code a generated table driver performs.
func (r *RowDisplacement) Lookup(rowIdx int, colIdx int) int {
	cellIdx := r.Base[rowIdx] + colIdx
	if r.Check[cellIdx] != colIdx {
		return r.EmptyValue
	}
	return r.Next[cellIdx]
}

// rowDisplacementBuilder places the rows of a sparse table one after the other. It holds the bookkeeping the placement
// needs but the resulting table does not.
type rowDisplacementBuilder struct {
	// result is the table built so far.
	result RowDisplacement

	// rows are the rows to place, in the order the row indices refer to them.
	rows [][]int

	// occupiedColumns holds, for every row, the columns it has an entry in, in ascending order. This is the list
	// the paper's appendix calls list(i). Working from it keeps the cost of trying a displacement proportional to
	// the number of entries of the row instead of to its width.
	occupiedColumns [][]int

	// lowZero is the index of the lowest cell which is still a hole. Every cell below it is occupied, so a
	// displacement which would put the first entry of a row below it can not fit.
	lowZero int

	// baseByContent holds the displacement of a row already placed, keyed by its content, so that rows which are
	// identical can share it.
	baseByContent map[string]int

	// usedBase holds the displacements handed out so far, so that rows which are not identical never share one.
	usedBase map[int]bool
}

// newRowDisplacementBuilder prepares the placement of the given rows.
func newRowDisplacementBuilder(rows [][]int, emptyValue int) *rowDisplacementBuilder {
	occupiedColumns := make([][]int, len(rows))
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			if value != emptyValue {
				occupiedColumns[rowIdx] = append(occupiedColumns[rowIdx], colIdx)
			}
		}
	}

	return &rowDisplacementBuilder{
		result: RowDisplacement{
			Base:       make([]int, len(rows)),
			EmptyValue: emptyValue,
		},
		rows:            rows,
		occupiedColumns: occupiedColumns,
		baseByContent:   make(map[string]int),
		usedBase:        make(map[int]bool),
	}
}

// placementOrder returns the row indices ordered by descending number of entries, which is the order the
// first-fit-decreasing method places the rows in. Placing the crowded rows first leaves the rows which are easy to fit
// for the end, when the table has the most holes to offer.
//
// This is the bucket sort of step 2 of the appendix of "Storing a Sparse Table" by Tarjan and Yao.
func (b *rowDisplacementBuilder) placementOrder() []int {
	var maxCount int
	for _, columns := range b.occupiedColumns {
		maxCount = max(maxCount, len(columns))
	}

	buckets := make([][]int, maxCount+1)
	for rowIdx, columns := range b.occupiedColumns {
		buckets[len(columns)] = append(buckets[len(columns)], rowIdx)
	}

	result := make([]int, 0, len(b.occupiedColumns))
	for _, v := range slices.Backward(buckets) {
		result = append(result, v...)
	}
	return result
}

// placeRow gives the given row a displacement and writes its entries into the cells that displacement puts them in.
func (b *rowDisplacementBuilder) placeRow(rowIdx int) {
	content := rowContent(b.rows[rowIdx])
	if base, ok := b.baseByContent[content]; ok {
		b.result.Base[rowIdx] = base
		return
	}

	base := b.firstFit(rowIdx)
	b.result.Base[rowIdx] = base
	b.baseByContent[content] = base
	b.usedBase[base] = true

	row := b.rows[rowIdx]
	for _, colIdx := range b.occupiedColumns[rowIdx] {
		b.grow(base + colIdx + 1)
		b.result.Next[base+colIdx] = row[colIdx]
		b.result.Check[base+colIdx] = colIdx
	}
	b.advanceLowZero()
}

// firstFit returns the smallest displacement which is not in use yet and on which no entry of the given row collides
// with an entry already placed.
//
// This is step 3 of the appendix of "Storing a Sparse Table" by Tarjan and Yao. The search does not start at zero but
// at the displacement which puts the first entry of the row on the lowest cell which is still a hole, because every
// smaller displacement would put that entry on an occupied cell.
func (b *rowDisplacementBuilder) firstFit(rowIdx int) int {
	columns := b.occupiedColumns[rowIdx]

	var base int
	if len(columns) > 0 {
		base = max(0, b.lowZero-columns[0])
	}
	for ; ; base++ {
		if !b.usedBase[base] && b.fits(columns, base) {
			return base
		}
	}
}

// fits reports whether the entries in the given columns can be placed at the given displacement without landing on a
// cell which is occupied already.
func (b *rowDisplacementBuilder) fits(columns []int, base int) bool {
	for _, colIdx := range columns {
		cellIdx := base + colIdx
		if cellIdx < len(b.result.Check) && b.result.Check[cellIdx] != NoColumn {
			return false
		}
	}
	return true
}

// advanceLowZero moves the lowest hole up past the cells which have just been occupied.
func (b *rowDisplacementBuilder) advanceLowZero() {
	for b.lowZero < len(b.result.Check) && b.result.Check[b.lowZero] != NoColumn {
		b.lowZero++
	}
}

// grow extends the table to at least the given number of cells, filling the new cells with holes.
func (b *rowDisplacementBuilder) grow(cellCount int) {
	for len(b.result.Next) < cellCount {
		b.result.Next = append(b.result.Next, b.result.EmptyValue)
		b.result.Check = append(b.result.Check, NoColumn)
	}
}

// finish extends the table so that every column of every row can be looked up without leaving it, even when the row has
// no entry there. This keeps a table driver from having to guard the lookup with a range check of its own.
func (b *rowDisplacementBuilder) finish() RowDisplacement {
	for rowIdx, row := range b.rows {
		b.grow(b.result.Base[rowIdx] + len(row))
	}
	return b.result
}

// rowContent returns a string which is equal for two rows exactly when they hold the same entries in the same columns,
// so that identical rows can be recognized. The values are encoded as variable length integers, which is self
// delimiting and therefore keeps rows of different length apart.
func rowContent(row []int) string {
	var builder strings.Builder
	var buffer [binary.MaxVarintLen64]byte
	for _, value := range row {
		builder.Write(buffer[:binary.PutVarint(buffer[:], int64(value))])
	}
	return builder.String()
}
