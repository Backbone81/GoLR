package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the rename tests of the plugins: GolrRenameTest.kt of the IntelliJ plugin and providers.test.ts of
// the VS Code plugin with its fixture sample.golr.
var _ = Describe("Rename", func() {
	DescribeTable("announces prepareRename only to a client which supports it",
		func(capabilities string, renameProvider string) {
			responses, err := run(false,
				`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":`+capabilities+`}}`,
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(2))
			Expect(responses[0]).To(ContainSubstring(`"renameProvider":` + renameProvider))
		},
		Entry("with prepareSupport", `{"textDocument":{"rename":{"prepareSupport":true}}}`,
			`{"prepareProvider":true}`),
		Entry("without prepareSupport", `{"textDocument":{"rename":{"prepareSupport":false}}}`, `true`),
		Entry("without rename capabilities", `{}`, `true`),
	)

	DescribeTable("prepares the rename of the name at the caret",
		func(source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.positionRequest(protocol.MethodTextDocumentPrepareRename),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.rangeResponse()))
		},
		Entry("a definition", utils.HereDoc(`
			@parser {
			    «a<caret>»
			        : a
			        ;
			}
		`)),
		Entry("a reference", utils.HereDoc(`
			@parser {
			    a
			        : «<caret>a»
			        ;
			}
		`)),
		Entry("an alias", utils.HereDoc(`
			@scanner {
			    PLUS: «"<caret>+"»;
			}
		`)),
	)

	DescribeTable("rejects preparing a rename without a name at the caret",
		func(source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.positionRequest(protocol.MethodTextDocumentPrepareRename),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(lsptest.BeErrorResponse(2, int(protocol.LSPErrorCodesRequestFailed)))
			Expect(responses[1]).To(ContainSubstring("no name to rename"))
		},
		Entry("the caret on no name", utils.HereDoc(`
			@parser {
			    a
			        :<caret> a
			        ;
			}
		`)),
		Entry("the caret on a keyword", utils.HereDoc(`
			@par<caret>ser {
			    a
			        : a
			        ;
			}
		`)),
	)

	DescribeTable("renames the declarations and references of the name at the caret",
		func(newName string, source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.renameRequest(newName),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.renameResponse(newName)))
		},
		Entry("from a definition", "x", utils.HereDoc(`
			@parser {
			    «a<caret>»
			        : b b
			        ;

			    b
			        : «a»
			        ;
			}
		`)),
		Entry("from a reference", "x", utils.HereDoc(`
			@parser {
			    «a»
			        : b b
			        ;

			    b
			        : «a<caret>»
			        ;
			}
		`)),
		Entry("a terminal, which keeps the references by its alias", "ADD", utils.HereDoc(`
			@scanner {
			    «PLUS<caret>»: "+";
			}

			@parser {
			    a
			        : a "+" a
			        | a «PLUS» a
			        ;
			}
		`)),
		Entry("an alias from a reference, which keeps the terminal", `"plus"`, utils.HereDoc(`
			@scanner {
			    PLUS: «"+"»;
			}

			@parser {
			    @precedence {
			        @left: «"+"»;
			    }

			    a
			        : a «"<caret>+"» a
			        | a PLUS a
			        ;
			}
		`)),
		Entry("an alias from its declaration", `"plus"`, utils.HereDoc(`
			@scanner {
			    PLUS: «"<caret>+"»;
			}

			@parser {
			    a
			        : a «"+"» a
			        | a PLUS a
			        ;
			}
		`)),
		Entry("a reference in the VS Code fixture", "factor", utils.HereDoc(`
			@scanner {
			    PLUS:    "+";
			    INTEGER: /[0-9]+/;
			}

			@parser {
			    @start: expression;

			    expression
			        : expression PLUS «<caret>term»
			        | «term»
			        ;

			    «term»
			        : INTEGER
			        ;
			}
		`)),
		Entry("a name declared twice", "B", utils.HereDoc(`
			@scanner {
			    «A»: "a";
			    «A»: "b";
			}

			@parser {
			    s
			        : «<caret>A»
			        ;
			}
		`)),
		Entry("an undefined name", "x", utils.HereDoc(`
			@parser {
			    a
			        : «<caret>undefined» «undefined»
			        ;
			}
		`)),
		Entry("to the old name", "a", utils.HereDoc(`
			@parser {
			    «a<caret>»
			        : «a»
			        ;
			}
		`)),
	)

	DescribeTable("rejects a rename",
		func(newName string, reason string, source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.renameRequest(newName),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(lsptest.BeErrorResponse(2, int(protocol.LSPErrorCodesRequestFailed)))
			Expect(responses[1]).To(ContainSubstring(reason))
		},
		Entry("of an alias to an identifier", "plus", "not a string", utils.HereDoc(`
			@scanner {
			    PLUS: "<caret>+";
			}
		`)),
		Entry("of an alias to two strings", `"a" "b"`, "not a string", utils.HereDoc(`
			@scanner {
			    PLUS: "<caret>+";
			}
		`)),
		Entry("of a symbol to a string", `"plus"`, "not an identifier", utils.HereDoc(`
			@scanner {
			    PLUS<caret>: "+";
			}
		`)),
		Entry("of a symbol to two identifiers", "a b", "not an identifier", utils.HereDoc(`
			@scanner {
			    PLUS<caret>: "+";
			}
		`)),
		Entry("of a symbol to an empty name", "", "not an identifier", utils.HereDoc(`
			@scanner {
			    PLUS<caret>: "+";
			}
		`)),
		Entry("of a symbol to a defined name", "b", "already defined", utils.HereDoc(`
			@parser {
			    a<caret>
			        : b
			        ;

			    b
			        : a
			        ;
			}
		`)),
		Entry("of a symbol to a referenced name", "undefined", "already referenced", utils.HereDoc(`
			@parser {
			    a<caret>
			        : undefined
			        ;
			}
		`)),
		Entry("of an alias to a defined alias", `"-"`, "already defined", utils.HereDoc(`
			@scanner {
			    PLUS:  "<caret>+";
			    MINUS: "-";
			}
		`)),
		Entry("without a name at the caret", "x", "no name to rename", utils.HereDoc(`
			@parser {
			    a
			        :<caret> a
			        ;
			}
		`)),
	)

	It("rejects a document which is not open", func() {
		document := newMarked(utils.HereDoc(`
			@parser {
			    a
			        : <caret>a
			        ;
			}
		`))
		responses, err := run(false,
			initializeRequest,
			document.renameRequest("x"),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
