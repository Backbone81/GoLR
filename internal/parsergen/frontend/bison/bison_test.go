package bison_test

import (
	"golr/internal/parsergen/frontend"
	"golr/internal/parsergen/frontend/bison"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bison Grammar Files", func() {
	It("should correctly parse the most minimal grammar", func() {
		bisonGrammar := `
			%%
			s:
		`
		grammar, err := bison.GrammarFromString(bisonGrammar)
		Expect(err).ToNot(HaveOccurred())
		Expect(grammar).To(Equal(frontend.Grammar{
			Terminals: nil,
			Nonterminals: []frontend.Symbol{
				{
					Name: "s",
				},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs:     nil,
				},
			},
			StartNonterminalIdx: 0,
		}))
	})

	It("should accept %empty", func() {
		bisonGrammar := `
			%%
			s: %empty
		`
		grammar, err := bison.GrammarFromString(bisonGrammar)
		Expect(err).ToNot(HaveOccurred())
		Expect(grammar).To(Equal(frontend.Grammar{
			Terminals: nil,
			Nonterminals: []frontend.Symbol{
				{
					Name: "s",
				},
			},
			Productions: []frontend.Production{
				{
					NonterminalIdx: 0,
					SymbolRefs:     nil,
				},
			},
			StartNonterminalIdx: 0,
		}))
	})

	Context("%token", func() {
		It("should accept single %token", func() {
			bisonGrammar := `
				%token FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %token with string alias", func() {
			bisonGrammar := `
				%token FOO "foo"
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `"foo"`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %token with tstring alias", func() {
			bisonGrammar := `
				%token FOO _("foo")
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `_("foo")`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple %token", func() {
			bisonGrammar := `
				%token FOO
				%token BAR
				%token BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
					{
						Name: "BAR",
					},
					{
						Name: "BAZ",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple %token with string alias", func() {
			bisonGrammar := `
				%token FOO "foo"
				%token BAR "bar"
				%token BAZ "baz"
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `"foo"`,
					},
					{
						Name:  "BAR",
						Alias: `"bar"`,
					},
					{
						Name:  "BAZ",
						Alias: `"baz"`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple %token with tstring alias", func() {
			bisonGrammar := `
				%token FOO _("foo")
				%token BAR _("bar")
				%token BAZ _("baz")
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `_("foo")`,
					},
					{
						Name:  "BAR",
						Alias: `_("bar")`,
					},
					{
						Name:  "BAZ",
						Alias: `_("baz")`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %token with multiple values", func() {
			bisonGrammar := `
				%token FOO BAR BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
					{
						Name: "BAR",
					},
					{
						Name: "BAZ",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %token with multiple values with string aliases", func() {
			bisonGrammar := `
				%token FOO "foo" BAR "bar" BAZ "baz"
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `"foo"`,
					},
					{
						Name:  "BAR",
						Alias: `"bar"`,
					},
					{
						Name:  "BAZ",
						Alias: `"baz"`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %token with multiple values with tstring aliases", func() {
			bisonGrammar := `
				%token FOO _("foo") BAR _("bar") BAZ _("baz")
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `_("foo")`,
					},
					{
						Name:  "BAR",
						Alias: `_("bar")`,
					},
					{
						Name:  "BAZ",
						Alias: `_("baz")`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("rules", func() {
		It("should accept a single terminal", func() {
			bisonGrammar := `
				%token FOO
				%%
				s: FOO
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept a single nonterminal", func() {
			bisonGrammar := `
				%%
				s: foo
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "foo",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept a mix of terminals and nonterminals", func() {
			bisonGrammar := `
				%token FOO BAR
				%%
				s: FOO baz BAR bat
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
					{
						Name: "BAR",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "baz",
					},
					{
						Name: "bat",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(1),
							frontend.NewNonterminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple alternatives", func() {
			bisonGrammar := `
				%%
				s: foo | bar | baz
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
					{
						Name: "baz",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(2),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(3),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept %empty as one alternatives", func() {
			bisonGrammar := `
				%%
				s: %empty | foo | bar
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple alternatives as separate rules", func() {
			bisonGrammar := `
				%%
				s: foo
				s: bar
				s: baz
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
					{
						Name: "baz",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(2),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(3),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should correctly map string aliases", func() {
			bisonGrammar := `
				%token FOO "foo"
				%%
				s: bar "foo" baz
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:  "FOO",
						Alias: `"foo"`,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "s",
					},
					{
						Name: "bar",
					},
					{
						Name: "baz",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(0),
							frontend.NewNonterminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("well known Bison grammars", func() {
		It("should correctly parse the Bison 3.8.2 grammar", func() {
			grammar, err := bison.GrammarFromFile("testdata/bison-3.8.2.y")
			Expect(err).ToNot(HaveOccurred())

			Expect(grammar.Terminals).To(HaveLen(54 + 1 + 1 + 1 + 1))
			Expect(grammar.Nonterminals).To(HaveLen(39))
			Expect(grammar.Productions).To(HaveLen(119))
		})

		It("should correctly parse the Go 1.5.4 grammar", func() {
			grammar, err := bison.GrammarFromFile("testdata/go-1.5.4.y")
			Expect(err).ToNot(HaveOccurred())

			Expect(grammar.Terminals).To(HaveLen(46 + 24))
			Expect(grammar.Nonterminals).To(HaveLen(127 + 1))
			Expect(grammar.Productions).To(HaveLen(127 + 210))
		})
	})
})
