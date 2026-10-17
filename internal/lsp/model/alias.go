package model

// Alias is the string of a terminal declared with a string, like "+" in `PLUS: "+";`. The name includes the quotes,
// as it is the text of every reference to it.
type Alias struct {
	Name     string
	Terminal string
	Span     Span
}
