package utils

import (
	"math"
	"strings"
)

// kotlinValuesPerFunction is the number of table entries one generated function holds.
//
// Kotlin compiles to the same virtual machine Java does, so a method may hold only 64 KB of bytecode and an array
// literal costs about eight bytes per entry. The reasoning of javaValuesPerMethod applies unchanged, and the limit is
// the same one, so a table is split over several functions instead of being written as one literal.
//
// A top level property of a Kotlin file is initialized in the one class initializer the compiler writes for that file,
// exactly the way a static field of a Java class is. Moving the entries out of the initializer and into a function each
// is therefore what buys the room here too.
const kotlinValuesPerFunction = 4000

// KotlinIntType returns the name of the narrowest Kotlin integer type which can hold every value from zero up to the
// given one. It is the Kotlin counterpart of GoUintType.
//
// The ladder is the Java one, because Kotlin inherits the signed integer types of the virtual machine and its unsigned
// types are library types over them rather than types an array of primitives can hold. So the bound of each step is
// the positive range of a signed type and not its width.
//
// Unlike in Java, none of these widens on its own in an expression. A lookup in a table typed this way is turned into
// an Int by KotlinTable.ToInt, which is what lets it be compared with an index or with another lookup.
func KotlinIntType(maxValue int) string {
	switch {
	case maxValue <= math.MaxInt8:
		return "Byte"
	case maxValue <= math.MaxInt16:
		return "Short"
	default:
		return "Int"
	}
}

// NewKotlinIntArray returns the given values as a table typed by the largest value it has to hold. It is the Kotlin
// counterpart of NewIntArray, which picks a Go type for the same values.
func NewKotlinIntArray(values []int) IntArray {
	var maxValue int
	for _, value := range values {
		maxValue = max(maxValue, value)
	}
	return NewTypedIntArray(KotlinIntType(maxValue), values)
}

// KotlinConstantName returns the given identifier in the upper case with underscores which Kotlin spells a constant
// with, so that transitionBase becomes TRANSITION_BASE.
func KotlinConstantName(identifier string) string {
	return upperSnakeName(identifier)
}

// KotlinString returns the given text as a Kotlin string literal, quotes included.
//
// Kotlin reads a dollar sign in a string literal as the start of a template expression, so a name which carries one -
// the augmented start symbol $accept and the end of input terminal $end both do - has to be escaped or the generated
// file does not compile.
func KotlinString(text string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
	)
	return `"` + replacer.Replace(text) + `"`
}

// KotlinArray is how the generated Kotlin code spells an array of one of the integer types KotlinIntType returns.
type KotlinArray struct {
	// Type is the Kotlin type of the array itself, which is one of the primitive array types such as ByteArray.
	Type string

	// ArrayOf is the name of the function which builds such an array from its entries, such as byteArrayOf.
	ArrayOf string

	// ToInt is what turns one entry into an Int, which is either the empty string for an array of Int or the call
	// which widens a narrower entry. Every lookup in the generated code carries it, because Kotlin widens no integer
	// type on its own.
	ToInt string

	// FromInt is what turns an Int into one entry, which is either the empty string for an array of Int or the call
	// which narrows it. It is the counterpart of ToInt and is carried by a value the generated code writes into such
	// an array as an Int literal, because Kotlin narrows no integer type on its own either.
	FromInt string
}

// NewKotlinArray returns how an array of the given entry type is spelled. Pass a type KotlinIntType returned.
func NewKotlinArray(elementType string) KotlinArray {
	toInt := ".toInt()"
	fromInt := ".to" + elementType + "()"
	if elementType == "Int" {
		toInt = ""
		fromInt = ""
	}

	return KotlinArray{
		Type:    elementType + "Array",
		ArrayOf: strings.ToLower(elementType[:1]) + elementType[1:] + "ArrayOf",
		ToInt:   toInt,
		FromInt: fromInt,
	}
}

// KotlinNameTable is a lookup table whose entries are the names of constants rather than numbers. It is chunked the
// same way a KotlinTable is, because the entries cost the same bytecode either way.
type KotlinNameTable struct {
	// Name is the name of the property holding the whole table.
	Name string

	// Function is the name of the functions holding the entries. Each one carries the index of its chunk as a suffix.
	Function string

	// Type is the Kotlin type of the whole table, which is an array of the type of one entry.
	Type string

	// ElementType is the Kotlin type of one entry, which the entries are qualified with.
	ElementType string

	// Chunks are the names of the entries, split into pieces small enough for one function each.
	Chunks [][]string
}

// NewKotlinNameTable returns the given entries under the given function name, split into as many chunks as they need.
func NewKotlinNameTable(function string, elementType string, names []string) KotlinNameTable {
	var chunks [][]string
	for start := 0; start < len(names); start += kotlinValuesPerFunction {
		chunks = append(chunks, names[start:min(start+kotlinValuesPerFunction, len(names))])
	}
	if len(chunks) == 0 {
		// A table with no entries at all still needs the one function the property is built from.
		chunks = append(chunks, nil)
	}

	return KotlinNameTable{
		Name:        KotlinConstantName(function),
		Function:    function,
		Type:        "Array<" + elementType + ">",
		ElementType: elementType,
		Chunks:      chunks,
	}
}

// KotlinTable is a lookup table in the form the generated Kotlin code holds it: a property joined from the arrays of
// several functions, because no single function can hold them all.
type KotlinTable struct {
	KotlinArray

	// Name is the name of the property holding the whole table.
	Name string

	// Function is the name of the functions holding the entries. Each one carries the index of its chunk as a suffix.
	Function string

	// Chunks are the entries, split into pieces small enough for one function each.
	Chunks []IntArray
}

// NewKotlinTable returns the given table under the given function name, split into as many chunks as it needs. The name
// of the property follows from the function name, so that the two can never drift apart.
func NewKotlinTable(function string, array IntArray) KotlinTable {
	var chunks []IntArray
	for start := 0; start < len(array.Values); start += kotlinValuesPerFunction {
		end := min(start+kotlinValuesPerFunction, len(array.Values))
		chunks = append(chunks, NewTypedIntArray(array.Type, array.Values[start:end]))
	}
	if len(chunks) == 0 {
		// A table with no entries at all still needs the one function the property is built from.
		chunks = append(chunks, NewTypedIntArray(array.Type, nil))
	}

	return KotlinTable{
		KotlinArray: NewKotlinArray(array.Type),
		Name:        KotlinConstantName(function),
		Function:    function,
		Chunks:      chunks,
	}
}
