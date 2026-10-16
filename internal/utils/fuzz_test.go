package utils_test

import (
	"testing"

	"github.com/backbone81/golr/internal/utils"
)

// FuzzRowDisplacement packs arbitrary sparse tables. A lookup of every row at every column has to return what the
// table holds there.
func FuzzRowDisplacement(f *testing.F) {
	f.Add([]byte{3, 1, 0, 0, 0, 2, 0, 0, 0, 3})
	f.Add([]byte{4, 1, 2, 0, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2})
	f.Add([]byte{2, 0, 0, 0, 0})
	f.Add([]byte{16, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7})

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) == 0 {
			return
		}
		// The first byte is the width of the table, every following byte a cell, where 0 is a cell without an entry.
		const emptyValue = -1
		width := 1 + int(input[0])%16
		var rows [][]int
		for cells := input[1:]; len(cells) >= width; cells = cells[width:] {
			row := make([]int, width)
			for colIdx := range row {
				row[colIdx] = int(cells[colIdx]%8) - 1
			}
			rows = append(rows, row)
		}

		displacement := utils.NewRowDisplacement(rows, emptyValue)
		for rowIdx, row := range rows {
			for colIdx, want := range row {
				if got := displacement.Lookup(rowIdx, colIdx); got != want {
					t.Fatalf("row %d column %d is %d instead of %d in %v", rowIdx, colIdx, got, want, rows)
				}
			}
		}
	})
}
