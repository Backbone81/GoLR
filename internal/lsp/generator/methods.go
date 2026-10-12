package generator

import (
	"io"
	"strings"
)

// methodsFileName is the file of the method constants.
const methodsFileName = "methods.gen.go"

// MethodsFromMetaModel writes the constants of the methods of all requests and notifications as Go source code.
func MethodsFromMetaModel(writer io.Writer, model *MetaModel) error {
	return render(writer, methodsFileName+".template", model)
}

// methodName returns the name of the constant of a method: Method followed by the parts of the method separated by /,
// each with its first letter in upper case. The part $ of the methods starting with $/ is dropped.
func methodName(method string) string {
	var builder strings.Builder
	builder.WriteString("Method")
	for part := range strings.SplitSeq(method, "/") {
		if part != "$" {
			builder.WriteString(exported(part))
		}
	}
	return builder.String()
}
