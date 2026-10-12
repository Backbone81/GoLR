package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// TypeKind is the kind of a type in the meta model.
type TypeKind string

const (
	TypeKindBase           TypeKind = "base"
	TypeKindReference      TypeKind = "reference"
	TypeKindArray          TypeKind = "array"
	TypeKindMap            TypeKind = "map"
	TypeKindAnd            TypeKind = "and"
	TypeKindOr             TypeKind = "or"
	TypeKindTuple          TypeKind = "tuple"
	TypeKindLiteral        TypeKind = "literal"
	TypeKindStringLiteral  TypeKind = "stringLiteral"
	TypeKindIntegerLiteral TypeKind = "integerLiteral"
	TypeKindBooleanLiteral TypeKind = "booleanLiteral"
)

// The names of the base types.
const (
	BaseTypeURI         = "URI"
	BaseTypeDocumentUri = "DocumentUri"
	BaseTypeInteger     = "integer"
	BaseTypeUinteger    = "uinteger"
	BaseTypeDecimal     = "decimal"
	BaseTypeRegExp      = "RegExp"
	BaseTypeString      = "string"
	BaseTypeBoolean     = "boolean"
	BaseTypeNull        = "null"
)

// Type is a type of the meta model. Which fields are set depends on the kind.
type Type struct {
	Kind TypeKind

	// Name is the base type or the referenced type.
	Name string

	// Element is the element type of an array.
	Element *Type

	// Key and Value are the key and value types of a map.
	Key   *Type
	Value *Type

	// Items are the types of the kinds and, or and tuple.
	Items []Type

	// Literal is the structure of a literal.
	Literal *StructureLiteral

	StringValue  string
	IntegerValue int64
	BooleanValue bool
}

// UnmarshalJSON decodes a type. Unknown properties and kinds are an error.
func (t *Type) UnmarshalJSON(data []byte) error {
	var raw struct {
		Kind    TypeKind        `json:"kind"`
		Name    string          `json:"name"`
		Element *Type           `json:"element"`
		Key     *Type           `json:"key"`
		Value   json.RawMessage `json:"value"`
		Items   []Type          `json:"items"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	*t = Type{
		Kind:    raw.Kind,
		Name:    raw.Name,
		Element: raw.Element,
		Key:     raw.Key,
		Items:   raw.Items,
	}

	var err error
	switch raw.Kind {
	case TypeKindBase, TypeKindReference, TypeKindArray, TypeKindAnd, TypeKindOr, TypeKindTuple:
	case TypeKindMap:
		t.Value = &Type{}
		err = json.Unmarshal(raw.Value, t.Value)
	case TypeKindLiteral:
		t.Literal = &StructureLiteral{}
		err = json.Unmarshal(raw.Value, t.Literal)
	case TypeKindStringLiteral:
		err = json.Unmarshal(raw.Value, &t.StringValue)
	case TypeKindIntegerLiteral:
		err = json.Unmarshal(raw.Value, &t.IntegerValue)
	case TypeKindBooleanLiteral:
		err = json.Unmarshal(raw.Value, &t.BooleanValue)
	default:
		return fmt.Errorf("unknown type kind %q", raw.Kind)
	}
	if err != nil {
		return fmt.Errorf("decoding the value of a %s type: %w", raw.Kind, err)
	}
	return nil
}
