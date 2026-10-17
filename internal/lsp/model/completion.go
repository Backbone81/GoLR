package model

// Completion is what completion offers at a caret.
type Completion struct {
	// Offset is where the word being completed starts, including a leading "@".
	Offset int

	// Keywords are the keywords valid at the caret.
	Keywords []string

	// Symbols reports whether the names of the symbols are valid at the caret.
	Symbols bool
}
