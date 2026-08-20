package utils

// PythonConstantName returns the given identifier in the upper case with underscores which PEP 8 spells a constant and
// an enumeration member with, so that TransitionBase becomes TRANSITION_BASE.
func PythonConstantName(identifier string) string {
	return upperSnakeName(identifier)
}

// NewPythonIntArray returns the given values as a table. It is the Python counterpart of NewIntArray, which picks a Go
// type for the same values.
//
// Python has one integer type and no fixed width, so there is nothing to narrow: the entries are written as they are
// and the table carries no type name. Every other backend picks the narrowest type its values fit into, which for
// Python would only decide between an array module type code and a tuple - and a tuple indexes faster than either.
func NewPythonIntArray(values []int) IntArray {
	return NewTypedIntArray("", values)
}
