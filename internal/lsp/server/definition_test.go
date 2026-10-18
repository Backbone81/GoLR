package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the go to definition tests of the plugins: GolrGoToDefinitionTest.kt of the IntelliJ plugin and
// providers.test.ts of the VS Code plugin with its fixture sample.golr.
var _ = Describe("Definition", func() {
	DescribeTable("returns the declarations of the name at the caret",
		func(source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.positionRequest(protocol.MethodTextDocumentDefinition),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.locationsResponse()))
		},
		Entry("a nonterminal in a production body", utils.HereDoc(`
			@parser {
			    a
			        : <caret>b
			        ;

			    «b»
			        : a
			        ;
			}
		`)),
		Entry("a terminal", utils.HereDoc(`
			@scanner {
			    «PLUS»: "+";
			}

			@parser {
			    expression
			        : <caret>PLUS
			        ;
			}
		`)),
		Entry("the symbol of @start", utils.HereDoc(`
			@parser {
			    @start: <caret>expression;

			    «expression»
			        : expression
			        ;
			}
		`)),
		Entry("an undefined name", utils.HereDoc(`
			@parser {
			    a
			        : <caret>undefined
			        ;
			}
		`)),
		Entry("an alias in a production body", utils.HereDoc(`
			@scanner {
			    PLUS: «"+"»;
			}

			@parser {
			    expression
			        : expression <caret>"+" expression
			        ;
			}
		`)),
		Entry("an alias in a precedence line", utils.HereDoc(`
			@scanner {
			    MINUS:  «"-"»;
			    UMINUS: @empty;
			}

			@parser {
			    @precedence {
			        @left:       <caret>"-";
			        @precedence: UMINUS;
			    }

			    expression
			        : "-" expression @precedence("-")
			        ;
			}
		`)),
		Entry("an alias in @precedence(...)", utils.HereDoc(`
			@scanner {
			    MINUS:  «"-"»;
			    UMINUS: @empty;
			}

			@parser {
			    @precedence {
			        @left:       "-";
			        @precedence: UMINUS;
			    }

			    expression
			        : "-" expression @precedence(<caret>"-")
			        ;
			}
		`)),
		Entry("a string of a fragment, which is no alias", utils.HereDoc(`
			@scanner {
			    DOT: "." @fragment;
			}

			@parser {
			    a
			        : <caret>"."
			        ;
			}
		`)),
		Entry("a reference in the VS Code fixture", utils.HereDoc(`
			@scanner {
			    PLUS:    "+";
			    INTEGER: /[0-9]+/;
			}

			@parser {
			    @start: expression;

			    expression
			        : expression PLUS <caret>term
			        | term
			        ;

			    «term»
			        : INTEGER
			        ;
			}
		`)),
		Entry("a definition", utils.HereDoc(`
			@parser {
			    <caret>«a»
			        : b
			        ;

			    b
			        : a
			        ;
			}
		`)),
		Entry("a name declared twice", utils.HereDoc(`
			@scanner {
			    «A»: "a";
			    «A»: "b";
			}

			@parser {
			    s
			        : <caret>A
			        ;
			}
		`)),
		Entry("the caret right after a name", utils.HereDoc(`
			@parser {
			    a
			        : b<caret>
			        ;

			    «b»
			        : a
			        ;
			}
		`)),
		Entry("the caret on no name", utils.HereDoc(`
			@parser {
			    a
			        :<caret> b
			        ;

			    b
			        : a
			        ;
			}
		`)),
		Entry("the caret on a keyword", utils.HereDoc(`
			@parser {
			    @sta<caret>rt: a;

			    a
			        : a
			        ;
			}
		`)),
		Entry("a character outside the Basic Multilingual Plane before the name", utils.HereDoc(`
			@parser {
			    /* 😀 */ a
			        : /* 😀 */ <caret>b
			        ;

			    /* 😀 */ «b»
			        : a
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
			document.positionRequest(protocol.MethodTextDocumentDefinition),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
