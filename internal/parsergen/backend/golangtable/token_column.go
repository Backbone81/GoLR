package golangtable

// TokenColumn is one entry of the table which translates a token into the column of the action table holding the
// decisions for it.
//
// It carries the name of the token constant rather than its value, because the value belongs to the scanner and this
// backend never learns it. The generated table is an array whose indexes are written as the token constants themselves,
// which lets the Go compiler place every entry at whatever value the scanner gave it. That is what keeps the two
// generators free of a numbering they would both have to agree on, see table.NoTerminalColumn.
type TokenColumn struct {
	// Name is the name of the token constant, which the generated table uses as the index of the entry.
	Name string

	// Column is the column of the action table which holds the decisions for the token.
	Column int
}
