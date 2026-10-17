package model

// Definition is the name of a scanner or parser rule.
type Definition struct {
	Name string
	Kind Kind
	Span Span
}
