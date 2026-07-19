package bison_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
	bisonfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/bison"
	"github.com/backbone81/golr/testdata"
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
			Terminals: []frontend.Symbol{
				{
					Name: "error",
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

	It("should accept %empty", func() {
		bisonGrammar := `
			%%
			s: %empty
		`
		grammar, err := bison.GrammarFromString(bisonGrammar)
		Expect(err).ToNot(HaveOccurred())
		Expect(grammar).To(Equal(frontend.Grammar{
			Terminals: []frontend.Symbol{
				{
					Name: "error",
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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
						Name: "error",
					},
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

	Context("%left", func() {
		It("should accept single %left with one token", func() {
			bisonGrammar := `
				%left FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %left with multiple tokens", func() {
			bisonGrammar := `
				%left FOO BAR BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should assign increasing precedence levels across multiple %left declarations", func() {
			bisonGrammar := `
				%left FOO
				%left BAR
				%left BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityLeft,
						Precedence:    2,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityLeft,
						Precedence:    3,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should update associativity and precedence when terminal is already declared via %token", func() {
			bisonGrammar := `
				%token FOO
				%left FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("%right", func() {
		It("should accept single %right with one token", func() {
			bisonGrammar := `
				%right FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %right with multiple tokens", func() {
			bisonGrammar := `
				%right FOO BAR BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should assign increasing precedence levels across multiple %right declarations", func() {
			bisonGrammar := `
				%right FOO
				%right BAR
				%right BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityRight,
						Precedence:    2,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityRight,
						Precedence:    3,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should update associativity and precedence when terminal is already declared via %token", func() {
			bisonGrammar := `
				%token FOO
				%right FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityRight,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("%nonassoc", func() {
		It("should accept single %nonassoc with one token", func() {
			bisonGrammar := `
				%nonassoc FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %nonassoc with multiple tokens", func() {
			bisonGrammar := `
				%nonassoc FOO BAR BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should assign increasing precedence levels across multiple %nonassoc declarations", func() {
			bisonGrammar := `
				%nonassoc FOO
				%nonassoc BAR
				%nonassoc BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityNone,
						Precedence:    2,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityNone,
						Precedence:    3,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should update associativity and precedence when terminal is already declared via %token", func() {
			bisonGrammar := `
				%token FOO
				%nonassoc FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityNone,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("%precedence", func() {
		It("should accept single %precedence with one token", func() {
			bisonGrammar := `
				%precedence FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept single %precedence with multiple tokens", func() {
			bisonGrammar := `
				%precedence FOO BAR BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should assign increasing precedence levels across multiple %precedence declarations", func() {
			bisonGrammar := `
				%precedence FOO
				%precedence BAR
				%precedence BAZ
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
					{
						Name:          "BAR",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    2,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    3,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should update associativity and precedence when terminal is already declared via %token", func() {
			bisonGrammar := `
				%token FOO
				%precedence FOO
				%%
				s:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityUndeclared,
						Precedence:    1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
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
						Name: "error",
					},
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
							frontend.NewTerminalRef(1),
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
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
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
						Name: "error",
					},
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
							frontend.NewTerminalRef(1),
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(2),
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
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
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
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
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
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
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
						Name: "error",
					},
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
							frontend.NewTerminalRef(1),
							frontend.NewNonterminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("%prec", func() {
		It("should set PrecedenceTerminalIdx on a production", func() {
			bisonGrammar := `
				%token FOO
				%%
				s: FOO %prec FOO
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			precedenceTerminalIdx := 1
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
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
							frontend.NewTerminalRef(1),
						},
						PrecedenceTerminalIdx: &precedenceTerminalIdx,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should set PrecedenceTerminalIdx only on the production it appears in", func() {
			bisonGrammar := `
				%token FOO BAR
				%%
				s: FOO %prec FOO | BAR
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			precedenceTerminalIdx := 1
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
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
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(1),
						},
						PrecedenceTerminalIdx: &precedenceTerminalIdx,
					},
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should allow %prec to reference a terminal not used in the production", func() {
			bisonGrammar := `
				%token FOO BAR
				%%
				s: FOO %prec BAR
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			precedenceTerminalIdx := 2
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
					{
						Name: "FOO",
					},
					{
						Name: "BAR",
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "s"},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(1),
						},
						PrecedenceTerminalIdx: &precedenceTerminalIdx,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("%start", func() {
		It("should set StartNonterminalIdx to the declared start nonterminal", func() {
			bisonGrammar := `
				%start b
				%%
				a:
				b:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "a",
					},
					{
						Name: "b",
					},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
					{NonterminalIdx: 1},
				},
				StartNonterminalIdx: 1,
			}))
		})

		It("should set StartNonterminalIdx when specified in the grammar declaration", func() {
			bisonGrammar := `
				%%
				a:
				b:
				%start b;
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "a"},
					{Name: "b"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
					{NonterminalIdx: 1},
				},
				StartNonterminalIdx: 1,
			}))
		})

		It("should use the first %start when multiple are declared", func() {
			bisonGrammar := `
				%start b
				%start a
				%%
				a:
				b:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "a"},
					{Name: "b"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
					{NonterminalIdx: 1},
				},
				StartNonterminalIdx: 1,
			}))
		})

		It("should default to the first nonterminal when no %start is declared", func() {
			bisonGrammar := `
				%%
				a:
				b:
			`
			grammar, err := bison.GrammarFromString(bisonGrammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "error",
					},
				},
				Nonterminals: []frontend.Symbol{
					{Name: "a"},
					{Name: "b"},
				},
				Productions: []frontend.Production{
					{NonterminalIdx: 0},
					{NonterminalIdx: 1},
				},
				StartNonterminalIdx: 0,
			}))
		})
	})

	Context("well known grammars", func() {
		for _, wellKnownGrammar := range testdata.WellKnownGrammars {
			It("should correctly parse the "+wellKnownGrammar.Title+" grammar", func() {
				grammar, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(grammar.Terminals).To(HaveLen(wellKnownGrammar.Terminals))
				Expect(grammar.Nonterminals).To(HaveLen(wellKnownGrammar.Nonterminals))
				Expect(grammar.Productions).To(HaveLen(wellKnownGrammar.Productions))
			})
		}
	})
})

func BenchmarkToGrammar(b *testing.B) {
	for _, wellKnownGrammar := range testdata.WellKnownGrammars {
		b.Run(wellKnownGrammar.Title, func(b *testing.B) {
			for b.Loop() {
				if _, err := bisonfrontend.ToGrammar(
					bytes.NewBuffer(wellKnownGrammar.Content),
					wellKnownGrammar.FileName,
				); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
