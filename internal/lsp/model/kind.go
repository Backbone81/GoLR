package model

// Kind is the kind of symbol a definition declares.
type Kind int

const (
	// KindTerminal is a terminal declared in the scanner section.
	KindTerminal Kind = iota

	// KindFragment is a terminal of the scanner section marked with @fragment.
	KindFragment

	// KindNonterminal is a nonterminal declared in the parser section.
	KindNonterminal
)
