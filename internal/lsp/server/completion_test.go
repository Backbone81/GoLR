package server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/lsp/jsonrpc"
	"github.com/backbone81/golr/internal/lsp/lsptest"
	"github.com/backbone81/golr/internal/lsp/protocol"
	"github.com/backbone81/golr/internal/utils"
)

// The cases follow the completion tests of the plugins: GolrCompletionTest.kt of the IntelliJ plugin,
// completion.test.ts and providers.test.ts of the VS Code plugin with its fixture sample.golr. The marked range is the
// word the items replace.
var _ = Describe("Completion", func() {
	DescribeTable("offers what is valid at the caret",
		func(source string, keywords []string, symbols []offeredSymbol) {
			document := newMarked(source)
			responses, err := run(false,
				initializeRequest,
				document.didOpen(),
				document.positionRequest(protocol.MethodTextDocumentCompletion),
				shutdownRequest,
				exitNotification,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(responses).To(HaveLen(3))
			Expect(responses[1]).To(MatchJSON(document.completionResponse(keywords, symbols)))
		},
		Entry("the section keywords at the top level", utils.HereDoc(`
			@scanner {
			    A: "a";
			}

			«<caret>»
		`), []string{"@scanner", "@parser"}, nil),
		Entry("the scanner annotations in a scanner rule body", utils.HereDoc(`
			@scanner {
			    A: "a" «<caret>»
			}
		`), []string{"@empty", "@skip", "@fragment"}, nil),
		Entry("the scanner annotations right after a string", utils.HereDoc(`
			@scanner {
			    A: "a"«<caret>»
			}
		`), []string{"@empty", "@skip", "@fragment"}, nil),
		Entry("the parser statement keywords at the start of a statement", utils.HereDoc(`
			@parser {
			    a
			        : a
			        ;

			    «<caret>»
			}
		`), []string{"@start", "@precedence"}, nil),
		Entry("the associativity keywords in a precedence block", utils.HereDoc(`
			@parser {
			    @precedence {
			        @left: a;
			        «<caret>»
			    }

			    a
			        : a
			        ;
			}
		`), []string{"@left", "@right", "@none", "@precedence"}, nil),
		Entry("@error and the symbols in a precedence line", utils.HereDoc(`
			@parser {
			    @precedence {
			        @left: a «<caret>»;
			    }

			    a
			        : a
			        ;
			}
		`), []string{"@error"}, []offeredSymbol{{"a", "nonterminal"}}),
		Entry("only the symbols after @start", utils.HereDoc(`
			@parser {
			    @start: «<caret>»;

			    a
			        : a
			        ;
			}
		`), nil, []offeredSymbol{{"a", "nonterminal"}}),
		Entry("only the symbols inside @precedence(...)", utils.HereDoc(`
			@parser {
			    a
			        : a @precedence(«<caret>»)
			        ;
			}
		`), nil, []offeredSymbol{{"a", "nonterminal"}}),
		Entry("the production body keywords and the nonterminals in a production body", utils.HereDoc(`
			@parser {
			    expression
			        : term «<caret>»
			        ;

			    term
			        : factor
			        ;

			    factor
			        : expression
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, []offeredSymbol{
			{"expression", "nonterminal"}, {"term", "nonterminal"}, {"factor", "nonterminal"},
		}),
		Entry("the terminals in a production body", utils.HereDoc(`
			@scanner {
			    INTEGER: /[0-9]+/;
			    PLUS:    "+";
			}

			@parser {
			    expression
			        : INTEGER «<caret>»
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, []offeredSymbol{
			{"INTEGER", "terminal"}, {"PLUS", "terminal"}, {"expression", "nonterminal"},
		}),
		Entry("the symbols of the VS Code fixture after INTEGER", utils.HereDoc(`
			@scanner {
			    PLUS:    "+";
			    INTEGER: /[0-9]+/;
			}

			@parser {
			    @start: expression;

			    expression
			        : expression PLUS term
			        | term
			        ;

			    term
			        : «INTEGER<caret>»
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, []offeredSymbol{
			{"PLUS", "terminal"}, {"INTEGER", "terminal"}, {"expression", "nonterminal"}, {"term", "nonterminal"},
		}),
		Entry("only the keywords after an @, replacing from the @", utils.HereDoc(`
			@parser {
			    a
			        : a «@na<caret>»
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, nil),
		Entry("the keywords and symbols for a partially typed identifier, replacing from its start", utils.HereDoc(`
			@parser {
			    expression
			        : term
			        ;

			    term
			        : «expr<caret>»
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, []offeredSymbol{
			{"expression", "nonterminal"}, {"term", "nonterminal"},
		}),
		Entry("a name declared twice once", utils.HereDoc(`
			@scanner {
			    A: "a";
			    A: "b";
			}

			@parser {
			    @start: «<caret>»;
			}
		`), nil, []offeredSymbol{{"A", "terminal"}}),
		Entry("no fragments, which are only valid in regular expressions", utils.HereDoc(`
			@scanner {
			    DIGIT:   /[0-9]/ @fragment;
			    INTEGER: /{DIGIT}+/;
			}

			@parser {
			    @start: «<caret>»;
			}
		`), nil, []offeredSymbol{{"INTEGER", "terminal"}}),
		Entry("the replaced range after a character outside the Basic Multilingual Plane", utils.HereDoc(`
			@parser {
			    a
			        : /* 😀 */ «a<caret>»
			        ;
			}
		`), []string{"@empty", "@error", "@precedence", "@name"}, []offeredSymbol{{"a", "nonterminal"}}),
		Entry("nothing inside @name(...)", utils.HereDoc(`
			@parser {
			    a
			        : a @name(<caret>)
			        ;
			}
		`), nil, nil),
		Entry("nothing between a rule name and its colon", utils.HereDoc(`
			@parser {
			    a <caret>
			        : a
			        ;
			}
		`), nil, nil),
		Entry("nothing in a comment", utils.HereDoc(`
			@parser {
			    // <caret>
			    a
			        : a
			        ;
			}
		`), nil, nil),
		Entry("nothing in a string", utils.HereDoc(`
			@scanner {
			    A: "a<caret>";
			}
		`), nil, nil),
		Entry("nothing in a regular expression", utils.HereDoc(`
			@scanner {
			    A: /a<caret>/;
			}
		`), nil, nil),
		Entry("nothing in an unterminated string", utils.HereDoc(`
			@parser {
			    a
			        : "a<caret>
			        ;
			}
		`), nil, nil),
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
			document.positionRequest(protocol.MethodTextDocumentCompletion),
			shutdownRequest,
			exitNotification,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(responses).To(HaveLen(3))
		Expect(responses[1]).To(lsptest.BeErrorResponse(2, jsonrpc.ErrorCodeInvalidParams))
	})
})
