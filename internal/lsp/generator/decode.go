package generator

import (
	"encoding/json"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// decodeFileName is the file of the decode functions.
const decodeFileName = "decode.gen.go"

// unionArmsTestFileName is the test file decoding every item of every union.
const unionArmsTestFileName = "decode.gen_test.go"

// exportedTypeName matches the exported type names in a Go type.
var exportedTypeName = regexp.MustCompile(`\b[A-Z]\w*`)

// DecodeFromMetaModel writes a decode function for every structure, enumeration, type alias and union as Go source
// code. The decode functions go into the package of the generated types and use its hand written decode helpers.
func DecodeFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, decodeFileName+".template", model)
}

// UnionArmsTestFromMetaModel writes a test for the package of the generated types as Go source code, which decodes a
// sample of every item of every union and expects the item to be chosen.
func UnionArmsTestFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, unionArmsTestFileName+".template", model)
}

// Decoder returns the expression of the decode function of a type, a function from the decoded JSON value to the Go
// type of the type. A union gets the given name. An unsupported type is written with the name of its kind, which does
// not compile.
func (c TemplateContext) Decoder(unionName string, typ Type) string {
	switch typ.Kind {
	case TypeKindBase:
		return baseDecoder(typ.Name)
	case TypeKindReference:
		return "Decode" + typ.Name
	case TypeKindArray:
		return "arrayOf(" + c.Decoder(unionName, *typ.Element) + ")"
	case TypeKindMap:
		return "mapOf[" + c.goType(unionName, *typ.Key) + "](" + c.Decoder(unionName, *typ.Value) + ")"
	case TypeKindOr:
		items := nonNullItems(typ)
		switch {
		case len(items) > 1:
			return "Decode" + unionName
		case len(items) == 0:
			return BaseTypeNull
		case c.category(items[0]) == valueCategory:
			return "pointerOf(" + c.Decoder(unionName, items[0]) + ")"
		default:
			return c.Decoder(unionName, items[0])
		}
	case TypeKindTuple:
		return c.tupleDecoder(unionName, typ)
	case TypeKindLiteral:
		if len(typ.Literal.Properties) == 0 {
			return "decodeEmptyObject"
		}
	case TypeKindStringLiteral:
		return "stringLiteralOf(" + strconv.Quote(typ.StringValue) + ")"
	case TypeKindAnd, TypeKindIntegerLiteral, TypeKindBooleanLiteral:
	}
	return string(typ.Kind)
}

// tupleDecoder returns the expression of the decode function of a tuple of two items of the same type, the name of
// the kind otherwise.
func (c TemplateContext) tupleDecoder(unionName string, tuple Type) string {
	if len(tuple.Items) != 2 || c.tupleGoType(unionName, tuple) == string(tuple.Kind) {
		return string(tuple.Kind)
	}
	return "pairOf(" + c.Decoder(unionName, tuple.Items[0]) + ")"
}

// PropertyDecoder returns the expression of the decode function of the field of a property.
func (c TemplateContext) PropertyDecoder(property StructureProperty) string {
	decoder := c.Decoder(c.propertyUnionName(property), property.Type)
	switch c.fieldWrapper(property) {
	case pointerWrapper:
		return "pointerOf(" + decoder + ")"
	case optionalWrapper:
		return "optionalOf(" + decoder + ")"
	case noWrapper:
	}
	return decoder
}

// SkipsNull reports whether null in a property counts as absent: the property is optional and null is not one of its
// values.
func (c TemplateContext) SkipsNull(property StructureProperty) bool {
	return property.Optional && !isNullable(property.Type)
}

// TypeAliasDecoder returns the expression of the decode function of the type of a type alias. LSPAny keeps the
// decoded JSON value as it is.
func (c TemplateContext) TypeAliasDecoder(alias TypeAlias) string {
	if alias.Name == anyTypeAlias {
		return "decodeAny"
	}
	return c.Decoder(alias.Name, alias.Type)
}

// DecodeOrder returns the items of a union besides null in the order they are tried when decoding: by the number of
// required properties, descending, so the more specific item wins. Items with the same number keep their order. An
// array counts the required properties of its element.
func (c TemplateContext) DecodeOrder(union Union) []Type {
	items := slices.Clone(nonNullItems(union.Type))
	slices.SortStableFunc(items, func(a Type, b Type) int {
		return len(c.requiredProperties(armElement(b))) - len(c.requiredProperties(armElement(a)))
	})
	return items
}

// ArmDecoder returns the expression decoding the JSON value in the variable value as an item of a union, giving the
// decoded item and whether it fits.
func (c TemplateContext) ArmDecoder(union Union, item Type) string {
	element := armElement(item)
	required := c.requiredProperties(element)
	arguments := make([]string, 0, 2+len(required))
	arguments = append(arguments, "value", c.Decoder(union.Name, element))
	for _, name := range required {
		arguments = append(arguments, strconv.Quote(name))
	}
	if item.Kind == TypeKindArray {
		return "decodeArrayArm(" + strings.Join(arguments, ", ") + ")"
	}
	return "decodeArm(" + strings.Join(arguments, ", ") + ")"
}

// requiredProperties returns the names of the required properties of a structure, looking through type aliases,
// nothing for another type.
func (c TemplateContext) requiredProperties(typ Type) []string {
	if typ.Kind != TypeKindReference {
		return nil
	}
	if alias, ok := c.typeAlias(typ.Name); ok {
		return c.requiredProperties(alias.Type)
	}
	structure, ok := c.structure(typ.Name)
	if !ok {
		return nil
	}
	var names []string
	for _, property := range c.StructureProperties(structure) {
		if !property.Optional {
			names = append(names, property.Name)
		}
	}
	return names
}

// enumerationDecoder returns the generic decode helper for the Go type an enumeration is defined over. An unsupported
// type is written with its name from the meta model, which does not compile.
func enumerationDecoder(enumeration Enumeration) string {
	if enumeration.Type.Kind == TypeKindBase {
		switch enumeration.Type.Name {
		case BaseTypeString:
			return "decodeString"
		case BaseTypeInteger, BaseTypeUinteger:
			return "decodeInteger"
		}
	}
	return enumeration.Type.Name
}

// baseDecoder returns the expression of the decode function of a base type. RegExp and null are written with their
// name, which does not compile.
func baseDecoder(name string) string {
	switch name {
	case BaseTypeString:
		return "decodeString[string]"
	case BaseTypeBoolean:
		return "decodeBoolean"
	case BaseTypeInteger, BaseTypeUinteger:
		return "decodeInteger[int]"
	case BaseTypeDecimal:
		return "decodeDecimal"
	case BaseTypeDocumentUri, BaseTypeURI:
		return "Decode" + name
	}
	return name
}

// armElement returns the element of an array, which decides the arm, and the type itself otherwise.
func armElement(item Type) Type {
	if item.Kind == TypeKindArray {
		return *item.Element
	}
	return item
}

// ArmGoType returns the Go type of an item of a union.
func (c TemplateContext) ArmGoType(union Union, item Type) string {
	return c.goType(union.Name, item)
}

// ArmSample returns the JSON text of a value of an item of a union, which the item has to decode as itself: an
// object holds the required properties only, an array one element.
func (c TemplateContext) ArmSample(item Type) string {
	text, err := json.Marshal(c.sample(item))
	if err != nil {
		return err.Error()
	}
	return string(text)
}

// sample returns a value of the type with the required properties of objects, one element in arrays, the first
// value of an enumeration and the first item of a union besides null. An unsupported type gives the name of its kind.
func (c TemplateContext) sample(typ Type) any {
	switch typ.Kind {
	case TypeKindBase:
		return baseSample(typ.Name)
	case TypeKindReference:
		return c.referenceSample(typ.Name)
	case TypeKindArray:
		return []any{c.sample(*typ.Element)}
	case TypeKindMap, TypeKindLiteral:
		return map[string]any{}
	case TypeKindOr:
		if items := nonNullItems(typ); len(items) > 0 {
			return c.sample(items[0])
		}
		return nil
	case TypeKindTuple:
		items := make([]any, 0, len(typ.Items))
		for _, item := range typ.Items {
			items = append(items, c.sample(item))
		}
		return items
	case TypeKindStringLiteral:
		return typ.StringValue
	case TypeKindAnd, TypeKindIntegerLiteral, TypeKindBooleanLiteral:
	}
	return string(typ.Kind)
}

// referenceSample returns a value of the structure, enumeration or type alias of the name.
func (c TemplateContext) referenceSample(name string) any {
	if structure, ok := c.structure(name); ok {
		object := map[string]any{}
		for _, property := range c.StructureProperties(structure) {
			if !property.Optional {
				object[property.Name] = c.sample(property.Type)
			}
		}
		return object
	}
	if alias, ok := c.typeAlias(name); ok {
		return c.sample(alias.Type)
	}
	index := slices.IndexFunc(c.MetaModel.Enumerations, func(enumeration Enumeration) bool {
		return enumeration.Name == name
	})
	if index >= 0 && len(c.MetaModel.Enumerations[index].Values) > 0 {
		return c.MetaModel.Enumerations[index].Values[0].Value
	}
	return name
}

// baseSample returns a value of a base type.
func baseSample(name string) any {
	switch name {
	case BaseTypeInteger, BaseTypeUinteger, BaseTypeDecimal:
		return 0
	case BaseTypeBoolean:
		return true
	case BaseTypeNull:
		return nil
	}
	return ""
}

// qualified returns the Go type with every exported type name qualified by the package protocol, for code outside of
// it.
func qualified(goType string) string {
	return exportedTypeName.ReplaceAllString(goType, "protocol.$0")
}
