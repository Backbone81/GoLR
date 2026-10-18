package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the reference count tests of the plugins: GolrReferencesCodeVisionTest.kt of the IntelliJ plugin,
// codeLens.test.ts and providers.test.ts of the VS Code plugin with its fixture sample.golr.
var _ = Describe("Code lens", func() {
	DescribeTable("shows the reference count above every definition",
		func(source string, titles []string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.documentRequest(protocol.MethodTextDocumentCodeLens),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.codeLensResponse(titles...)))
		},
		Entry("with one and two references", utils.HereDoc(`
			@parser {
			    «a»
			        : b b
			        ;

			    «b»
			        : a
			        ;
			}
		`), []string{"1 reference", "2 references"}),
		Entry("with no references", utils.HereDoc(`
			@parser {
			    «SourceFiles»
			        : PackageClause ";"
			        | @empty
			        ;

			    «PackageClause»
			        : "package" identifier
			        ;
			}
		`), []string{"0 references", "1 reference"}),
		Entry("of the VS Code fixture", utils.HereDoc(`
			@scanner {
			    «PLUS»:    "+";
			    «INTEGER»: /[0-9]+/;
			}

			@parser {
			    @start: expression;

			    «expression»
			        : expression PLUS term
			        | term
			        ;

			    «term»
			        : INTEGER
			        ;
			}
		`), []string{"1 reference", "1 reference", "2 references", "2 references"}),
		Entry("counting the references by alias", utils.HereDoc(`
			@scanner {
			    «PLUS»: "+";
			}

			@parser {
			    «e»
			        : e PLUS e
			        | e "+" e
			        ;
			}
		`), []string{"2 references", "4 references"}),
		Entry("without fragments", utils.HereDoc(`
			@scanner {
			    DIGIT:     /[0-9]/ @fragment;
			    «INTEGER»: /{DIGIT}+/;
			}
		`), []string{"0 references"}),
		Entry("on both declarations of a name declared twice", utils.HereDoc(`
			@scanner {
			    «A»: "a";
			    «A»: "b";
			}

			@parser {
			    «s»
			        : A A "a"
			        ;
			}
		`), []string{"3 references", "3 references", "0 references"}),
		Entry("without a lens for an unresolved name", utils.HereDoc(`
			@parser {
			    «a»
			        : UNDEFINED
			        ;
			}
		`), []string{"0 references"}),
		Entry("of a document without definitions", "@parser {\n}\n", []string{}),
	)

	It("rejects a document which is not open", func() {
		document := newMarked("@parser {\n}\n")
		responses, err := run(false,
			initializeRequest,
			document.documentRequest(protocol.MethodTextDocumentCodeLens),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
