package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// The decode functions take the tree encoding/json decodes into any with UseNumber: map[string]any, []any, string,
// json.Number, bool and nil.

// DecodeError is a value which does not decode into its type.
type DecodeError struct {
	// Path leads to the value through properties, array indexes and map keys, e.g. edits[2].range. It is empty for the
	// value passed to the decode function.
	Path string

	Message string
}

func (e *DecodeError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return e.Path + ": " + e.Message
}

// withPathSegment returns the error with the segment prepended to the path. A segment starts with [ for an index or
// key, which then needs no separator.
func withPathSegment(err error, segment string) error {
	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		decodeErr = &DecodeError{Message: err.Error()}
	}
	switch {
	case decodeErr.Path == "":
		decodeErr.Path = segment
	case strings.HasPrefix(decodeErr.Path, "["):
		decodeErr.Path = segment + decodeErr.Path
	default:
		decodeErr.Path = segment + "." + decodeErr.Path
	}
	return decodeErr
}

// kindError reports a value which is not of the expected kind.
func kindError(expected string, value any) error {
	return &DecodeError{Message: fmt.Sprintf("expected %s, got %s", expected, jsonKind(value))}
}

// jsonKind returns the JSON kind of the value.
func jsonKind(value any) string {
	switch value := value.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case json.Number:
		return "number " + value.String()
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", value)
	}
}

// decodeObject returns the properties of an object.
func decodeObject(value any) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, kindError("object", value)
	}
	return object, nil
}

// decodeEmptyObject decodes an object of any properties into the empty literal.
func decodeEmptyObject(value any) (struct{}, error) {
	_, err := decodeObject(value)
	return struct{}{}, err
}

// decodeAny keeps the value as it is.
func decodeAny(value any) (any, error) {
	return value, nil
}

// decodeString decodes a string into a type based on string.
//
//nolint:ireturn // T is a string type, not an interface.
func decodeString[T ~string](value any) (T, error) {
	text, ok := value.(string)
	if !ok {
		return "", kindError("string", value)
	}
	return T(text), nil
}

// decodeBoolean decodes a boolean.
func decodeBoolean(value any) (bool, error) {
	boolean, ok := value.(bool)
	if !ok {
		return false, kindError("boolean", value)
	}
	return boolean, nil
}

// decodeInteger decodes an integral number into a type based on int.
//
//nolint:ireturn // T is an integer type, not an interface.
func decodeInteger[T ~int](value any) (T, error) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, kindError("integer", value)
	}
	integer, err := strconv.Atoi(number.String())
	if err != nil {
		return 0, kindError("integer", value)
	}
	return T(integer), nil
}

// decodeDecimal decodes a number.
func decodeDecimal(value any) (float64, error) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, kindError("decimal", value)
	}
	decimal, err := number.Float64()
	if err != nil {
		return 0, kindError("decimal", value)
	}
	return decimal, nil
}

// stringLiteralOf returns a decode function accepting only the string.
func stringLiteralOf(expected string) func(any) (string, error) {
	return func(value any) (string, error) {
		text, err := decodeString[string](value)
		if err != nil {
			return "", err
		}
		if text != expected {
			return "", &DecodeError{Message: fmt.Sprintf("expected %q, got %q", expected, text)}
		}
		return text, nil
	}
}

// arrayOf returns a decode function for an array of the elements. null decodes to nil.
func arrayOf[T any](decodeElement func(any) (T, error)) func(any) ([]T, error) {
	return func(value any) ([]T, error) {
		if value == nil {
			return nil, nil
		}
		array, ok := value.([]any)
		if !ok {
			return nil, kindError("array", value)
		}
		elements := make([]T, 0, len(array))
		for index, element := range array {
			decoded, err := decodeElement(element)
			if err != nil {
				return nil, withPathSegment(err, "["+strconv.Itoa(index)+"]")
			}
			elements = append(elements, decoded)
		}
		return elements, nil
	}
}

// pairOf returns a decode function for an array of exactly two elements.
func pairOf[T any](decodeElement func(any) (T, error)) func(any) ([2]T, error) {
	return func(value any) ([2]T, error) {
		var pair [2]T
		elements, err := arrayOf(decodeElement)(value)
		if err != nil {
			return pair, err
		}
		if len(elements) != len(pair) {
			return pair, &DecodeError{Message: fmt.Sprintf("expected 2 elements, got %d", len(elements))}
		}
		return [2]T(elements), nil
	}
}

// mapOf returns a decode function for an object whose properties are the entries of a map. null decodes to nil.
func mapOf[K ~string, V any](decodeValue func(any) (V, error)) func(any) (map[K]V, error) {
	return func(value any) (map[K]V, error) {
		if value == nil {
			return nil, nil
		}
		object, err := decodeObject(value)
		if err != nil {
			return nil, err
		}
		entries := make(map[K]V, len(object))
		for key, entry := range object {
			decoded, err := decodeValue(entry)
			if err != nil {
				return nil, withPathSegment(err, "["+strconv.Quote(key)+"]")
			}
			entries[K(key)] = decoded
		}
		return entries, nil
	}
}

// pointerOf returns a decode function for a pointer to the value. null decodes to nil.
func pointerOf[T any](decode func(any) (T, error)) func(any) (*T, error) {
	return func(value any) (*T, error) {
		if value == nil {
			return nil, nil
		}
		decoded, err := decode(value)
		if err != nil {
			return nil, err
		}
		return &decoded, nil
	}
}

// optionalOf returns a decode function for a present optional value.
func optionalOf[T any](decode func(any) (T, error)) func(any) (Optional[T], error) {
	return func(value any) (Optional[T], error) {
		decoded, err := decode(value)
		if err != nil {
			return Optional[T]{}, err
		}
		return NewOptional(decoded), nil
	}
}

// decodeArm decodes the value as an arm of a union. It fits when it is an object with the required properties and
// decodes. Without required properties, only decoding counts.
//
//nolint:ireturn // T is the type of the arm, not an interface.
func decodeArm[T any](value any, decode func(any) (T, error), required ...string) (T, bool) {
	if !hasProperties(value, required) {
		var zero T
		return zero, false
	}
	decoded, err := decode(value)
	return decoded, err == nil
}

// decodeArrayArm decodes the value as an arm of a union which is an array. The first element decides like a
// non-array arm, an empty array fits.
func decodeArrayArm[T any](value any, decodeElement func(any) (T, error), required ...string) ([]T, bool) {
	array, ok := value.([]any)
	if !ok || len(array) > 0 && !hasProperties(array[0], required) {
		return nil, false
	}
	decoded, err := arrayOf(decodeElement)(value)
	return decoded, err == nil
}

// hasProperties reports whether the value is an object with all the properties. Without properties, every value has
// them.
func hasProperties(value any, names []string) bool {
	if len(names) == 0 {
		return true
	}
	object, ok := value.(map[string]any)
	if !ok {
		return false
	}
	for _, name := range names {
		if _, ok := object[name]; !ok {
			return false
		}
	}
	return true
}
