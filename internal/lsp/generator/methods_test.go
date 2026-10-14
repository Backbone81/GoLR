package generator_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/generator"
	"github.com/backbone81/golr/internal/utils"
)

// generateMethods returns the method constants generated from the meta model.
func generateMethods(model *generator.MetaModel) string {
	var builder strings.Builder
	Expect(generator.MethodsFromMetaModel(&builder, model)).To(Succeed())
	return builder.String()
}

var _ = Describe("MethodsFromMetaModel", func() {
	It("should name the constants after the method and link their section of the specification", func() {
		model := &generator.MetaModel{
			MetaData: generator.MetaData{Version: "3.18.0"},
			Requests: []generator.Request{
				{Method: "textDocument/semanticTokens/full/delta"},
				{Method: "initialize"},
			},
			Notifications: []generator.Notification{
				{Method: "$/cancelRequest"},
				{Method: "textDocument/didOpen"},
			},
		}
		//nolint:lll // The expected comments hold the links to the specification.
		Expect(generateMethods(model)).To(ContainSubstring(utils.HereDoc(`
			// The methods of the requests.
			const (
				// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#semanticTokens_deltaRequest.
				MethodTextDocumentSemanticTokensFullDelta = "textDocument/semanticTokens/full/delta"
				// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#initialize.
				MethodInitialize = "initialize"
			)

			// The methods of the notifications.
			const (
				// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#cancelRequest.
				MethodCancelRequest = "$/cancelRequest"
				// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#textDocument_didOpen.
				MethodTextDocumentDidOpen = "textDocument/didOpen"
			)
		`)))
	})

	It("should mark a deprecated method without the documentation of the meta model", func() {
		model := &generator.MetaModel{
			MetaData: generator.MetaData{Version: "3.18.0"},
			Notifications: []generator.Notification{{
				Annotations: generator.Annotations{
					Documentation: "Sent from the client.",
					Deprecated:    "use something else",
				},
				Method: "a/b",
			}},
		}
		Expect(generateMethods(model)).To(ContainSubstring(utils.HereDoc(`
			// The methods of the notifications.
			const (
				// See https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/#a_b.
				//
				// Deprecated: deprecated by the Language Server Protocol.
				MethodAB = "a/b"
			)
		`)))
	})
})
