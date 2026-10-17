package model

import "errors"

var (
	// ErrNotIdentifier rejects a new name of a symbol which is not exactly one IDENTIFIER.
	ErrNotIdentifier = errors.New("not an identifier")

	// ErrNotString rejects a new name of an alias which is not exactly one STRING.
	ErrNotString = errors.New("not a string")

	// ErrNameDefined rejects a new name which is defined already, which would merge two symbols.
	ErrNameDefined = errors.New("already defined")

	// ErrNameReferenced rejects a new name which is referenced already, which would take over its references.
	ErrNameReferenced = errors.New("already referenced")
)
