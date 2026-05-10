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

			// All %token declarations
			Expect(grammar.Terminals).To(HaveLen(58))

			// All left hand sides of productions + error nonterminal
			// Note that grammar_declaration shows up twice and must be counted only once.
			Expect(grammar.Nonterminals).To(HaveLen(38 + 1))

			// All productions + alternatives
			// Note that not all alternatives start in the first column. symbols.1 has an alternative which is indented
			// and therefore easy to overlook with a regex search in the grammar file.
			Expect(grammar.Productions).To(HaveLen(39 + 80))
		})

		It("should correctly parse the Go 1.5.4 grammar", func() {
			grammar, err := bison.GrammarFromFile("testdata/go-1.5.4.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + char literals
			Expect(grammar.Terminals).To(HaveLen(46 + 24))

			// All left hand sides of productions + error nonterminal
			Expect(grammar.Nonterminals).To(HaveLen(127 + 1))

			// All productions + alternatives
			Expect(grammar.Productions).To(HaveLen(127 + 210))
		})

		It("should correctly parse the GCC 4.2.4 Java grammar", func() {
			grammar, err := bison.GrammarFromFile("testdata/gcc-4.2.4-java.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations
			// Note that there are duplicate %token declarations to assign a tag after declaration. Searching for all
			// %token declarations would therefore result in duplicate tokens.
			Expect(grammar.Terminals).To(HaveLen(109))

			// All left hand sides of productions + error nonterminal
			// Note that searching for identifiers at the start of the line with a colon at the end will turn up results
			// in comments which need to be ignored.
			Expect(grammar.Nonterminals).To(HaveLen(153 + 1))

			// All productions + alternatives
			// Note that one alternative is inside of a block comment starting with "Screws up thing". We need to remove
			// that from the result.
			Expect(grammar.Productions).To(HaveLen(153 + 352))
		})

		PIt("should correctly parse the GCC 2.95.3 C grammar", func() {
			grammar, err := bison.GrammarFromFile("testdata/gcc-2.95.3-c.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations
			Expect(grammar.Terminals).To(HaveLen(47))

			// All left hand sides of productions + error nonterminal
			Expect(grammar.Nonterminals).To(HaveLen(0 + 1))

			// All productions + alternatives
			Expect(grammar.Productions).To(HaveLen(0 + 0))
		})
	})
})
