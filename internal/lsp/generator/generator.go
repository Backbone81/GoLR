package generator

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

//go:embed templates/*.go.template
var templateFiles embed.FS

var parsedTemplates = template.Must(template.New("").Funcs(template.FuncMap{
	"deprecation":          deprecation,
	"methodName":           methodName,
	"enumerationGoType":    enumerationGoType,
	"enumerationValueName": enumerationValueName,
	"enumerationValue":     enumerationValue,
	"enumerationDecoder":   enumerationDecoder,
	"exported":             exported,
	"propertyTag":          propertyTag,
	"typeList":             typeList,
	"join":                 strings.Join,
	"hasPrefix":            strings.HasPrefix,
	"qualified":            qualified,
}).ParseFS(templateFiles, "templates/*.go.template"))

type TemplateContext struct {
	MetaModel *MetaModel
}

// MetaModelToDirectory writes the files generated from the meta model into the directory, always the same set of
// files.
func MetaModelToDirectory(directory string, model *MetaModel) error {
	if err := toFile(filepath.Join(directory, methodsFileName), model, MethodsFromMetaModel); err != nil {
		return err
	}
	if err := toFile(filepath.Join(directory, enumerationsFileName), model, EnumerationsFromMetaModel); err != nil {
		return err
	}
	if err := toFile(filepath.Join(directory, structuresFileName), model, StructuresFromMetaModel); err != nil {
		return err
	}
	if err := toFile(filepath.Join(directory, unionsFileName), model, UnionsFromMetaModel); err != nil {
		return err
	}
	if err := toFile(filepath.Join(directory, decodeFileName), model, DecodeFromMetaModel); err != nil {
		return err
	}
	return toFile(filepath.Join(directory, unionArmsTestFileName), model, UnionArmsTestFromMetaModel)
}

// toFile writes the output of the write function into the file.
func toFile(
	filePath string,
	model *MetaModel,
	write func(writer io.Writer, model *MetaModel) error,
) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the Go file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return write(file, model)
}

// render writes the named template for the meta model as formatted Go source code. Source code which cannot be
// formatted is written as it is, together with the error.
func render(writer io.Writer, templateName string, model *MetaModel) error {
	var buffer bytes.Buffer
	if err := parsedTemplates.ExecuteTemplate(&buffer, templateName, TemplateContext{
		MetaModel: model,
	}); err != nil {
		return fmt.Errorf("rendering the template: %w", err)
	}
	source := buffer.Bytes()

	var joinedErr error
	formatted, err := format.Source(source)
	if err != nil {
		joinedErr = errors.Join(joinedErr, err)
	} else {
		source = formatted
	}

	if _, err := writer.Write(source); err != nil {
		joinedErr = errors.Join(joinedErr, err)
	}
	return joinedErr
}

// exported returns the name with its first letter in upper case.
func exported(name string) string {
	first, size := utf8.DecodeRuneInString(name)
	return string(unicode.ToUpper(first)) + name[size:]
}

// deprecation returns the Deprecated: comment line for a deprecated element, empty otherwise. The reason from the meta
// model is not repeated.
func deprecation(annotations Annotations) string {
	if annotations.Deprecated == "" {
		return ""
	}
	return "// Deprecated: deprecated by the Language Server Protocol.\n"
}
