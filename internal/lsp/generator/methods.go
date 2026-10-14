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

// methodAnchors are the anchors of the sections of the specification which are not named after their method.
//
//nolint:gosec // The anchors are no credentials.
var methodAnchors = map[string]string{
	"textDocument/semanticTokens/full":       "semanticTokens_fullRequest",
	"textDocument/semanticTokens/full/delta": "semanticTokens_deltaRequest",
	"textDocument/semanticTokens/range":      "semanticTokens_rangeRequest",
	"workspace/semanticTokens/refresh":       "semanticTokens_refreshRequest",
	"workspace/diagnostic/refresh":           "diagnostic_refresh",
	"workspace/codeLens/refresh":             "codeLens_refresh",
	"workspace/textDocumentContent/refresh":  "workspace_textDocumentContentRefresh",
	"workspaceSymbol/resolve":                "workspace_symbolResolve",
	"textDocument/rangesFormatting":          "textDocument_rangeFormatting",
}

// methodAnchor returns the anchor of the section of the specification of a method: the method without a leading $/
// and with _ instead of /, unless the section has another anchor.
func methodAnchor(method string) string {
	if anchor, ok := methodAnchors[method]; ok {
		return anchor
	}
	return strings.ReplaceAll(strings.TrimPrefix(method, "$/"), "/", "_")
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
