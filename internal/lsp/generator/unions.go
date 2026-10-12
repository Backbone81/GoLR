package generator

import (
	"io"
	"strings"
)

// unionsFileName is the file of the union types.
const unionsFileName = "unions.gen.go"

// UnionsFromMetaModel writes a union type for every union with more than one item besides null as Go source code.
func UnionsFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, unionsFileName+".template", model)
}

// Union is a union of the meta model with the name of its Go type.
type Union struct {
	Annotations

	Name string
	Type Type
}

// Unions returns the unions of the meta model in its order: in the properties of structures, named after structure
// and property, as type alias, named after the alias, and in the results of requests, named after the type name of
// the request and Result or PartialResult.
func (c TemplateContext) Unions() []Union {
	var unions []Union
	for _, structure := range c.MetaModel.Structures {
		for _, property := range structure.Properties {
			unions = c.appendUnions(unions, structure.Name+exported(property.Name), property.Type)
		}
	}
	for _, alias := range c.MetaModel.TypeAliases {
		if alias.Name == anyTypeAlias {
			continue
		}
		if unions = c.appendUnions(unions, alias.Name, alias.Type); len(nonNullItems(alias.Type)) > 1 {
			unions[len(unions)-1].Annotations = alias.Annotations
		}
	}
	for _, request := range c.MetaModel.Requests {
		if request.Result != nil {
			unions = c.appendUnions(unions, request.TypeName+"Result", *request.Result)
		}
		if request.PartialResult != nil {
			unions = c.appendUnions(unions, request.TypeName+"PartialResult", *request.PartialResult)
		}
	}
	return unions
}

// appendUnions appends the unions with more than one item besides null in the type, all with the given name.
func (c TemplateContext) appendUnions(unions []Union, name string, typ Type) []Union {
	switch typ.Kind {
	case TypeKindArray:
		return c.appendUnions(unions, name, *typ.Element)
	case TypeKindMap:
		return c.appendUnions(c.appendUnions(unions, name, *typ.Key), name, *typ.Value)
	case TypeKindOr, TypeKindAnd, TypeKindTuple:
		for _, item := range typ.Items {
			unions = c.appendUnions(unions, name, item)
		}
		if len(nonNullItems(typ)) > 1 {
			unions = append(unions, Union{Name: name, Type: typ})
		}
	case TypeKindBase, TypeKindReference, TypeKindLiteral, TypeKindStringLiteral, TypeKindIntegerLiteral,
		TypeKindBooleanLiteral:
	}
	return unions
}

// UnionItems returns the Go types of the items of a union besides null.
func (c TemplateContext) UnionItems(union Union) []string {
	items := nonNullItems(union.Type)
	goTypes := make([]string, 0, len(items))
	for _, item := range items {
		goTypes = append(goTypes, c.goType(union.Name, item))
	}
	return goTypes
}

// typeList joins the Go types for a sentence: a, b or c.
func typeList(goTypes []string) string {
	if len(goTypes) < 2 {
		return strings.Join(goTypes, "")
	}
	return strings.Join(goTypes[:len(goTypes)-1], ", ") + " or " + goTypes[len(goTypes)-1]
}
