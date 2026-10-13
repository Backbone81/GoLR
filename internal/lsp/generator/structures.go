package generator

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

// structuresFileName is the file of the structures and type aliases.
const structuresFileName = "structures.gen.go"

// anyTypeAlias is the type alias of an arbitrary JSON value.
const anyTypeAlias = "LSPAny"

// goAliases are the type aliases written as Go aliases of their type, so the values encoding/json decodes into have
// these types at every depth.
var goAliases = []string{anyTypeAlias, "LSPObject", "LSPArray"}

// StructuresFromMetaModel writes the structures with their inherited properties and the type aliases which are no
// unions as Go source code.
func StructuresFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, structuresFileName+".template", model)
}

// StructureProperty is a property of a structure together with the structure declaring it.
type StructureProperty struct {
	Property

	// Structure is the name of the structure declaring the property, which can be a parent of the structure holding it.
	Structure string
}

// typeCategory tells how a Go type represents an absent value.
type typeCategory int

const (
	// valueCategory needs a pointer or Optional for an absent value.
	valueCategory typeCategory = iota

	// collectionCategory is a slice or map, nil when absent.
	collectionCategory

	// unionCategory is a union type or any, its zero value when absent.
	unionCategory
)

// IsGenerated reports whether the structure gets a Go type. A structure whose name starts with _ only exists to be
// inherited from.
func (c TemplateContext) IsGenerated(structure Structure) bool {
	return !strings.HasPrefix(structure.Name, "_")
}

// StructureProperties returns the properties of the structure including the ones inherited through extends and mixins,
// in this order. A property the structure redeclares replaces the inherited one at its place.
func (c TemplateContext) StructureProperties(structure Structure) []StructureProperty {
	var properties []StructureProperty
	for _, parent := range slices.Concat(structure.Extends, structure.Mixins) {
		if parentStructure, ok := c.structure(parent.Name); ok {
			properties = c.mergeProperties(properties, c.StructureProperties(parentStructure))
		}
	}
	own := make([]StructureProperty, 0, len(structure.Properties))
	for _, property := range structure.Properties {
		own = append(own, StructureProperty{Property: property, Structure: structure.Name})
	}
	return c.mergeProperties(properties, own)
}

// mergeProperties adds the properties to the list, replacing a property of the same name at its place.
func (c TemplateContext) mergeProperties(
	properties []StructureProperty,
	added []StructureProperty,
) []StructureProperty {
	for _, property := range added {
		index := slices.IndexFunc(properties, func(existing StructureProperty) bool {
			return existing.Name == property.Name
		})
		if index < 0 {
			properties = append(properties, property)
		} else {
			properties[index] = property
		}
	}
	return properties
}

// fieldWrapper tells how the field of a property wraps the Go type of the property.
type fieldWrapper int

const (
	// noWrapper is the Go type of the property itself.
	noWrapper fieldWrapper = iota

	// pointerWrapper is a pointer to the Go type of the property.
	pointerWrapper

	// optionalWrapper is an Optional of the Go type of the property.
	optionalWrapper
)

// PropertyGoType returns the Go type of the field of a property.
func (c TemplateContext) PropertyGoType(property StructureProperty) string {
	goType := c.goType(c.propertyUnionName(property), property.Type)
	switch c.fieldWrapper(property) {
	case pointerWrapper:
		return "*" + goType
	case optionalWrapper:
		return "Optional[" + goType + "]"
	case noWrapper:
	}
	return goType
}

// fieldWrapper returns how the field of a property wraps its Go type. An optional property is an Optional, except a
// collection or union which is absent as its zero value, and a reference to the declaring structure, which is a
// pointer.
func (c TemplateContext) fieldWrapper(property StructureProperty) fieldWrapper {
	if !property.Optional {
		return noWrapper
	}
	if property.Type.Kind == TypeKindReference && property.Type.Name == property.Structure {
		return pointerWrapper
	}
	switch c.category(property.Type) {
	case collectionCategory:
		if !isNullable(property.Type) {
			return noWrapper
		}
	case unionCategory:
		return noWrapper
	case valueCategory:
	}
	return optionalWrapper
}

// propertyUnionName returns the name of a union in the type of a property.
func (c TemplateContext) propertyUnionName(property StructureProperty) string {
	return property.Structure + exported(property.Name)
}

// propertyTag returns the struct tag of the field of a property.
func propertyTag(property Property) string {
	if property.Optional {
		return fmt.Sprintf("`json:%q`", property.Name+",omitzero")
	}
	return fmt.Sprintf("`json:%q`", property.Name)
}

// TypeAliasDefinition returns the type definition of a type alias after its name, with = for a Go alias. A union is
// written as union type instead and gets an empty definition.
func (c TemplateContext) TypeAliasDefinition(alias TypeAlias) string {
	switch {
	case alias.Name == anyTypeAlias:
		return "= any"
	case slices.Contains(goAliases, alias.Name), alias.Type.Kind == TypeKindReference:
		return "= " + c.goType(alias.Name, alias.Type)
	case len(nonNullItems(alias.Type)) > 1:
		return ""
	default:
		return c.goType(alias.Name, alias.Type)
	}
}

// goType returns the Go type of a type. A union gets the given name. An unsupported type is written with the name of
// its kind, which does not compile.
func (c TemplateContext) goType(unionName string, typ Type) string {
	switch typ.Kind {
	case TypeKindBase:
		return baseGoType(typ.Name)
	case TypeKindReference:
		return typ.Name
	case TypeKindArray:
		return "[]" + c.goType(unionName, *typ.Element)
	case TypeKindMap:
		return "map[" + c.goType(unionName, *typ.Key) + "]" + c.goType(unionName, *typ.Value)
	case TypeKindOr:
		items := nonNullItems(typ)
		switch {
		case len(items) > 1:
			return unionName
		case len(items) == 0:
			return BaseTypeNull
		case c.category(items[0]) == valueCategory:
			return "*" + c.goType(unionName, items[0])
		default:
			return c.goType(unionName, items[0])
		}
	case TypeKindTuple:
		return c.tupleGoType(unionName, typ)
	case TypeKindLiteral:
		if len(typ.Literal.Properties) == 0 {
			return "struct{}"
		}
	case TypeKindStringLiteral:
		return "string"
	case TypeKindAnd, TypeKindIntegerLiteral, TypeKindBooleanLiteral:
	}
	return string(typ.Kind)
}

// tupleGoType returns the Go array of a tuple whose items all have the same type, the name of the kind otherwise.
func (c TemplateContext) tupleGoType(unionName string, tuple Type) string {
	if len(tuple.Items) == 0 {
		return string(tuple.Kind)
	}
	element := c.goType(unionName, tuple.Items[0])
	for _, item := range tuple.Items[1:] {
		if c.goType(unionName, item) != element {
			return string(tuple.Kind)
		}
	}
	return fmt.Sprintf("[%d]%s", len(tuple.Items), element)
}

// category returns how the Go type of a type represents an absent value, looking through type aliases.
func (c TemplateContext) category(typ Type) typeCategory {
	switch typ.Kind {
	case TypeKindArray, TypeKindMap:
		return collectionCategory
	case TypeKindOr:
		items := nonNullItems(typ)
		switch len(items) {
		case 0:
			return valueCategory
		case 1:
			return c.category(items[0])
		default:
			return unionCategory
		}
	case TypeKindReference:
		if alias, ok := c.typeAlias(typ.Name); ok {
			return c.category(alias.Type)
		}
	case TypeKindBase, TypeKindAnd, TypeKindTuple, TypeKindLiteral, TypeKindStringLiteral,
		TypeKindIntegerLiteral, TypeKindBooleanLiteral:
	}
	return valueCategory
}

// structure returns the structure of the name.
func (c TemplateContext) structure(name string) (Structure, bool) {
	index := slices.IndexFunc(c.MetaModel.Structures, func(structure Structure) bool {
		return structure.Name == name
	})
	if index < 0 {
		return Structure{}, false
	}
	return c.MetaModel.Structures[index], true
}

// typeAlias returns the type alias of the name.
func (c TemplateContext) typeAlias(name string) (TypeAlias, bool) {
	index := slices.IndexFunc(c.MetaModel.TypeAliases, func(alias TypeAlias) bool {
		return alias.Name == name
	})
	if index < 0 {
		return TypeAlias{}, false
	}
	return c.MetaModel.TypeAliases[index], true
}

// baseGoType returns the Go type of a base type. RegExp and null are written with their name, which does not compile.
func baseGoType(name string) string {
	switch name {
	case BaseTypeString:
		return "string"
	case BaseTypeBoolean:
		return "bool"
	case BaseTypeInteger, BaseTypeUinteger:
		return "int"
	case BaseTypeDecimal:
		return "float64"
	}
	return name
}

// isNullable reports whether null is one of the items of a union.
func isNullable(typ Type) bool {
	return typ.Kind == TypeKindOr && len(nonNullItems(typ)) < len(typ.Items)
}

// nonNullItems returns the items of a union without null, nothing for another kind.
func nonNullItems(typ Type) []Type {
	if typ.Kind != TypeKindOr {
		return nil
	}
	var items []Type
	for _, item := range typ.Items {
		if item.Kind != TypeKindBase || item.Name != BaseTypeNull {
			items = append(items, item)
		}
	}
	return items
}
