// Package model builds the symbol model of a GoLR grammar file: the definitions, aliases and references the language
// features work on. It is a token-based pass over the generated golr scanner, so it answers on every text, broken or
// not. Offsets are bytes into the text the model was built from.
package model
