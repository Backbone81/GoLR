package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the find usages tests of the plugins: GolrFindUsagesTest.kt of the IntelliJ plugin and
// providers.test.ts of the VS Code plugin with its fixture sample.golr.
var _ = Describe("References", func() {
	DescribeTable("returns the references of the name at the caret",
		func(includeDeclaration bool, source string) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.referencesRequest(includeDeclaration),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.locationsResponse()))
		},
		Entry("a nonterminal referenced twice", false, utils.HereDoc(`
			@parser {
			    a
			        : «b» «b»
			        ;

			    <caret>b
			        : a
			        ;
			}
		`)),
		Entry("a nonterminal referenced once", false, utils.HereDoc(`
			@parser {
			    <caret>a
			        : b b
			        ;

			    b
			        : «a»
			        ;
			}
		`)),
		Entry("a terminal referenced in @parser", false, utils.HereDoc(`
			@scanner {
			    <caret>PLUS: "+";
			}

			@parser {
			    expression
			        : expression «PLUS» expression
			        ;
			}
		`)),
		Entry("a nonterminal referencing itself", false, utils.HereDoc(`
			@scanner {
			    PLUS: "+";
			}

			@parser {
			    <caret>expression
			        : «expression» PLUS «expression»
			        ;
			}
		`)),
		Entry("an unused symbol", false, utils.HereDoc(`
			@parser {
			    a
			        : a
			        ;

			    <caret>unused
			        : a
			        ;
			}
		`)),
		Entry("a terminal with the references by its alias", false, utils.HereDoc(`
			@scanner {
			    <caret>PLUS: "+";
			}

			@parser {
			    @precedence {
			        @left: «"+"»;
			    }

			    expression
			        : expression «"+"» expression
			        | expression «PLUS» expression
			        ;
			}
		`)),
		Entry("an alias with only the references by its string", false, utils.HereDoc(`
			@scanner {
			    PLUS: <caret>"+";
			}

			@parser {
			    expression
			        : expression «"+"» expression
			        | expression PLUS expression
			        ;
			}
		`)),
		Entry("a reference to a terminal", false, utils.HereDoc(`
			@scanner {
			    PLUS: "+";
			}

			@parser {
			    expression
			        : expression «"+"» expression
			        | expression <caret>«PLUS» expression
			        ;
			}
		`)),
		Entry("a reference by an alias", false, utils.HereDoc(`
			@scanner {
			    PLUS: "+";
			}

			@parser {
			    expression
			        : expression <caret>«"+"» expression
			        | expression PLUS expression
			        ;
			}
		`)),
		Entry("a nonterminal with its definition", true, utils.HereDoc(`
			@parser {
			    a
			        : «b» «b»
			        ;

			    <caret>«b»
			        : a
			        ;
			}
		`)),
		Entry("a terminal with its definition and the references by its alias", true, utils.HereDoc(`
			@scanner {
			    «PLUS»: "+";
			}

			@parser {
			    expression
			        : expression «"+"» expression
			        | expression <caret>«PLUS» expression
			        ;
			}
		`)),
		Entry("an alias with its declaration", true, utils.HereDoc(`
			@scanner {
			    PLUS: «"+"»;
			}

			@parser {
			    expression
			        : expression <caret>«"+"» expression
			        | expression PLUS expression
			        ;
			}
		`)),
		Entry("a name declared twice with both declarations", true, utils.HereDoc(`
			@scanner {
			    «A»: "a";
			    «A»: "b";
			}

			@parser {
			    s
			        : <caret>«A»
			        ;
			}
		`)),
		Entry("the definition in the VS Code fixture", true, utils.HereDoc(`
			@scanner {
			    PLUS:    "+";
			    INTEGER: /[0-9]+/;
			}

			@parser {
			    @start: expression;

			    expression
			        : expression PLUS «term»
			        | «term»
			        ;

			    <caret>«term»
			        : INTEGER
			        ;
			}
		`)),
		Entry("the caret on no name", true, utils.HereDoc(`
			@parser {
			    a
			        :<caret> b
			        ;

			    b
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
			document.referencesRequest(true),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
