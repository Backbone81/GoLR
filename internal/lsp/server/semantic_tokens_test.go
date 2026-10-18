package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the semantic tokens provider of the VS Code plugin (semanticTokensProvider.ts) and its test in
// providers.test.ts with the fixture sample.golr. The IntelliJ plugin highlights by the lexer alone.
var _ = Describe("Semantic tokens", func() {
	DescribeTable("marks the definitions and references",
		func(source string, tokens []string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.documentRequest(protocol.MethodTextDocumentSemanticTokensFull),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.semanticTokensResponse(tokens...)))
		},
		Entry("of the VS Code fixture", utils.HereDoc(`
			@scanner {
			    «PLUS»:    "+";
			    «INTEGER»: /[0-9]+/;
			}

			@parser {
			    @start: «expression»;

			    «expression»
			        : «expression» «PLUS» «term»
			        | «term»
			        ;

			    «term»
			        : «INTEGER»
			        ;
			}
		`), []string{
			"enum declaration", "enum declaration",
			"class",
			"class declaration", "class", "enum", "class", "class",
			"class declaration", "enum",
		}),
		Entry("without aliases and references by alias", utils.HereDoc(`
			@scanner {
			    «PLUS»: "+";
			}

			@parser {
			    @precedence {
			        @left: "+";
			    }

			    «a»
			        : «a» "+" «a»
			        ;
			}
		`), []string{"enum declaration", "class declaration", "class", "class"}),
		Entry("with an unresolved reference as a nonterminal", utils.HereDoc(`
			@parser {
			    «a»
			        : «UNDEFINED»
			        ;
			}
		`), []string{"class declaration", "class"}),
		Entry("with a fragment as a terminal and no references inside regular expressions", utils.HereDoc(`
			@scanner {
			    «DIGIT»:   /[0-9]/ @fragment;
			    «INTEGER»: /{DIGIT}+/;
			}
		`), []string{"enum declaration", "enum declaration"}),
		Entry("with a reference of the kind of the first definition with its name", utils.HereDoc(`
			@scanner {
			    «A»: "a";
			}

			@parser {
			    «A»
			        : «A»
			        ;
			}
		`), []string{"enum declaration", "class declaration", "enum"}),
		Entry("in precedence lines and @precedence(...)", utils.HereDoc(`
			@scanner {
			    «MINUS»:  "-";
			    «UMINUS»: @empty;
			}

			@parser {
			    @precedence {
			        @left:       «MINUS»;
			        @precedence: «UMINUS»;
			    }

			    «e»
			        : «MINUS» «e» @precedence(«UMINUS»)
			        ;
			}
		`), []string{
			"enum declaration", "enum declaration",
			"enum", "enum",
			"class declaration", "enum", "class", "enum",
		}),
		Entry("of a document without symbols", "@parser {\n}\n", []string{}),
	)

	DescribeTable("counts in the negotiated position encoding",
		func(positionEncodings string, data string) {
			// The comment before b is two characters longer in UTF-8 than in UTF-16.
			didOpen := `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{
				"uri":"file:///a.golr","languageId":"golr","version":1,
				"text":"@parser {\n    a\n        : /* 😀 */ bb\n        ;\n}\n"}}}`
			responses, err := run(false,
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"processId":null,"capabilities":{
					"general":{"positionEncodings":`+positionEncodings+`}}}}`,
				didOpen,
				`{"jsonrpc":"2.0","id":2,"method":"textDocument/semanticTokens/full","params":{
					"textDocument":{"uri":"file:///a.golr"}}}`,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(`{"jsonrpc":"2.0","id":2,"result":{"data":` + data + `}}`))
		},
		Entry("UTF-8 when the client supports it", `["utf-8"]`, `[1,4,1,0,1, 1,21,2,0,0]`),
		Entry("UTF-16 otherwise", `["utf-16"]`, `[1,4,1,0,1, 1,19,2,0,0]`),
	)

	It("rejects a document which is not open", func() {
		document := newMarked("@parser {\n}\n")
		responses, err := run(false,
			initializeRequest,
			document.documentRequest(protocol.MethodTextDocumentSemanticTokensFull),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
