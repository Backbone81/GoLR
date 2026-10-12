package generator

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
)

// enumerationsFileName is the file of the enumerations.
const enumerationsFileName = "enums.gen.go"

// EnumerationsFromMetaModel writes the enumerations as Go source code, each as defined type with a constant per value.
func EnumerationsFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, enumerationsFileName+".template", model)
}

// enumerationGoType returns the Go type an enumeration is defined over. An unsupported type is written with its name
// from the meta model, which does not compile.
func enumerationGoType(enumeration Enumeration) string {
	if enumeration.Type.Kind == TypeKindBase {
		switch enumeration.Type.Name {
		case BaseTypeString:
			return "string"
		case BaseTypeInteger, BaseTypeUinteger:
			return "int"
		}
	}
	return enumeration.Type.Name
}

// enumerationValueName returns the name of the constant of a value: the name of the enumeration followed by the name
// of the value with its first letter in upper case.
func enumerationValueName(enumeration Enumeration, entry EnumerationEntry) string {
	return enumeration.Name + exported(entry.Name)
}

// enumerationValue returns the Go literal of a value. A number is written as it is, so the compiler rejects one which
// does not fit the type of the enumeration.
func enumerationValue(entry EnumerationEntry) string {
	var text string
	if err := json.Unmarshal(entry.Value, &text); err == nil {
		return strconv.Quote(text)
	}
	return string(bytes.TrimSpace(entry.Value))
}
