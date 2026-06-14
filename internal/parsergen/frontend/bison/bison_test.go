package bison_test

import (
	"bytes"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
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

	Context("well known Bison grammars", func() {
		It("should correctly parse the Bison 3.8.2 grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/bison-3.8.2.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + error token
			Expect(grammar.Terminals).To(HaveLen(58 + 1))

			// All left hand sides of productions
			// Note that grammar_declaration shows up twice and must be counted only once.
			Expect(grammar.Nonterminals).To(HaveLen(38))

			// All productions + alternatives
			// Note that not all alternatives start in the first column. symbols.1 has an alternative which is indented
			// and therefore easy to overlook with a regex search in the grammar file.
			Expect(grammar.Productions).To(HaveLen(39 + 80))
		})

		It("should correctly parse the GCC 2.95.3 C grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-2.95.3-c.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + error token + %left + %right + %nonassoc + char literals
			Expect(grammar.Terminals).To(HaveLen(47 + 1 + 19 + 7 + 2 + 6))

			// All left hand sides of productions
			// Note that the production for all_iter_stmt_with_decl is commented out and needs to be removed from the
			// list for a correct count.
			Expect(grammar.Nonterminals).To(HaveLen(117))

			// All productions + alternatives
			// Note that some alternatives are commented out and need to be removed from the count.
			Expect(grammar.Productions).To(HaveLen(117 + 247))
		})

		It("should correctly parse the GCC 2.95.3 Objective C grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-2.95.3-objc.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + error token + %left + %right + %nonassoc + char literals
			Expect(grammar.Terminals).To(HaveLen(47 + 1 + 19 + 7 + 2 + 6))

			// All left hand sides of productions
			// Note that the production for all_iter_stmt_with_decl is commented out and needs to be removed from the
			// list for a correct count.
			Expect(grammar.Nonterminals).To(HaveLen(162))

			// All productions + alternatives
			// Note that some alternatives are commented out and need to be removed from the count.
			Expect(grammar.Productions).To(HaveLen(162 + 340))
		})

		It("should correctly parse the GCC 3.3.6 C++ grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-3.3.6-cpp.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + %left + %right + %nonassoc + char literals
			// Note that some terminals show up as duplicates between %token and %nonassoc or %left and need to be
			// counted once only.
			Expect(grammar.Terminals).To(HaveLen(68 + 32 + 9 + 3))

			// All left hand sides of productions
			// Note that error was declared as a token and therefore does not show up in the list of nonterminals. In
			// addition the rule for primary_no_id is commented out and needs to be rmeoved.
			Expect(grammar.Nonterminals).To(HaveLen(238))

			// All productions + alternatives
			// Note that some alternatives are commented out and need to be removed from the count.
			Expect(grammar.Productions).To(HaveLen(238 + 633))
		})

		It("should correctly parse the GCC 4.2.4 Java grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/gcc-4.2.4-java.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + error token
			// Note that there are duplicate %token declarations to assign a tag after declaration. Searching for all
			// %token declarations would therefore result in duplicate tokens.
			Expect(grammar.Terminals).To(HaveLen(109 + 1))

			// All left hand sides of productions
			// Note that searching for identifiers at the start of the line with a colon at the end will turn up results
			// in comments which need to be ignored.
			Expect(grammar.Nonterminals).To(HaveLen(153))

			// All productions + alternatives
			// Note that one alternative is inside of a block comment starting with "Screws up thing". We need to remove
			// that from the result.
			Expect(grammar.Productions).To(HaveLen(153 + 352))
		})

		It("should correctly parse the Go 1.5.4 grammar", func() {
			grammar, err := bison.GrammarFromFile("../../../../testdata/go-1.5.4.y")
			Expect(err).ToNot(HaveOccurred())

			// All %token declarations + error token + %left + char literals
			// Note that some %left declarations are identical to %token and should not be counted twice.
			Expect(grammar.Terminals).To(HaveLen(46 + 1 + 3 + 24))

			// All left hand sides of productions
			Expect(grammar.Nonterminals).To(HaveLen(127))

			// All productions + alternatives
			Expect(grammar.Productions).To(HaveLen(127 + 210))
		})

		It("should correctly parse the PostgreSQL 18.4 grammar", func() {
			_, err := bison.GrammarFromFile("../../../../testdata/postgres-18.4.y")
			Expect(err).ToNot(HaveOccurred())

			//// All %token declarations + error token + %left + char literals
			//// Note that some %left declarations are identical to %token and should not be counted twice.
			//Expect(grammar.Terminals).To(HaveLen(46 + 1 + 3 + 24))
			//
			//// All left hand sides of productions
			//Expect(grammar.Nonterminals).To(HaveLen(127))
			//
			//// All productions + alternatives
			//Expect(grammar.Productions).To(HaveLen(127 + 210))
		})

	})
})

func BenchmarkToGrammar(b *testing.B) {
	b.Run("GNU Bison 3.8.2", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/bison-3.8.2.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GCC 2.95.3 C", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/gcc-2.95.3-c.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GCC 2.95.3 Objective C", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/gcc-2.95.3-objc.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GCC 3.3.6 C++", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/gcc-3.3.6-cpp.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GCC 4.2.4 Java", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/gcc-4.2.4-java.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Go 1.5.4", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/go-1.5.4.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("PostgreSQL 18.4", func(b *testing.B) {
		data, err := os.ReadFile("../../../../testdata/postgres-18.4.y")
		if err != nil {
			b.Fatal(err)
		}

		for b.Loop() {
			if _, err := bison.ToGrammar(bytes.NewReader(data), "in-memory"); err != nil {
				b.Fatal(err)
			}
		}
	})
}
