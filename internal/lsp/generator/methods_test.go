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
	It("should name the constants after the method", func() {
		model := &generator.MetaModel{
			Requests: []generator.Request{
				{Method: "textDocument/semanticTokens/full/delta"},
				{Method: "initialize"},
			},
			Notifications: []generator.Notification{
				{Method: "$/cancelRequest"},
			},
		}
		Expect(generateMethods(model)).To(ContainSubstring(utils.HereDoc(`
			// The methods of the requests.
			const (
				MethodTextDocumentSemanticTokensFullDelta = "textDocument/semanticTokens/full/delta"
				MethodInitialize                          = "initialize"
			)

			// The methods of the notifications.
			const (
				MethodCancelRequest = "$/cancelRequest"
			)
		`)))
	})

	It("should mark a deprecated method without the documentation of the meta model", func() {
		model := &generator.MetaModel{
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
				// Deprecated: deprecated by the Language Server Protocol.
				MethodAB = "a/b"
			)
		`)))
	})
})
