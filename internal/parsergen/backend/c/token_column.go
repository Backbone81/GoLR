package c

// TokenColumn is one entry of the lookup which translates a token into the column of the action table holding the
// decisions for it.
//
// It carries the name of the enumerator rather than its value, because the value belongs to the scanner and this
// backend never learns it. The generated lookup switches on the enumerators themselves, which is what keeps the two
// generators free of a numbering they would both have to agree on, see table.NoTerminalColumn.
type TokenColumn struct {
	// Name is the name of the token enumerator.
	Name string

	// Column is the column of the action table which holds the decisions for the token.
	Column int
}
