package golr_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

var _ = Describe("GoLR Grammar Files", func() {
	It("should correctly parse the most minimal grammar", func() {
		source := `
			@scanner {
			}
			@parser {
				file: @empty;
			}
		`
		_, grammar, err := golr.GrammarFromString(source)
		Expect(err).ToNot(HaveOccurred())
		Expect(grammar).To(Equal(frontend.Grammar{
			Terminals: nil,
			Nonterminals: []frontend.Symbol{
				{
					Name: "file",
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

	It("should reject a grammar without productions", func() {
		source := `
			@scanner {
			}
			@parser {
			}
		`
		_, _, err := golr.GrammarFromString(source)
		Expect(err).To(HaveOccurred())
	})

	Context("Tokens", func() {
		It("should accept a token with a regular expression", func() {
			source := `
				@scanner {
					FOO: /foo/;
				}
				@parser {
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
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

		It("should accept a token with a string literal", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
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
						Name: "file",
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

		It("should accept an empty token", func() {
			source := `
				@scanner {
					FOO: @empty;
				}
				@parser {
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
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

		It("should accept multiple tokens", func() {
			source := `
				@scanner {
					FOO: /foo/;
					BAR: "bar";
					BAZ: @empty;
				}
				@parser {
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
					{
						Name:  "BAR",
						Alias: `"bar"`,
					},
					{
						Name: "BAZ",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
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

		It("should reject duplicate token declarations", func() {
			source := `
				@scanner {
					FOO: /foo/;
					FOO: "bar";
				}
				@parser {
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject duplicate token aliases", func() {
			source := `
				@scanner {
					FOO: "baz";
					BAR: "baz";
				}
				@parser {
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject a token with an invalid regular expression", func() {
			source := `
				@scanner {
					FOO: /[unclosed/;
				}
				@parser {
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Associativity and Precedence of Tokens", func() {
		It("should correctly set associativity and precedence", func() {
			source := `
				@scanner {
					FOO: @empty;
					BAR: @empty;
					BAZ: @empty;
					BAT: @empty;
				}
				@parser {
					@precedence {
						@left: FOO;
						@right: BAR;
						@none: BAZ;
						@precedence: BAT;
					}
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:          "FOO",
						Precedence:    math.MaxInt - 1,
						Associativity: frontend.AssociativityLeft,
					},
					{
						Name:          "BAR",
						Precedence:    math.MaxInt - 2,
						Associativity: frontend.AssociativityRight,
					},
					{
						Name:          "BAZ",
						Precedence:    math.MaxInt - 3,
						Associativity: frontend.AssociativityNone,
					},
					{
						Name:       "BAT",
						Precedence: math.MaxInt - 4,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
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

		It("should support multiple tokens on the same precedence", func() {
			source := `
				@scanner {
					FOO: @empty;
					BAR: "bar";
					BAZ: @empty;
				}
				@parser {
					@precedence {
						@left: FOO "bar" BAZ;
					}
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:          "FOO",
						Precedence:    math.MaxInt - 1,
						Associativity: frontend.AssociativityLeft,
					},
					{
						Name:          "BAR",
						Alias:         `"bar"`,
						Precedence:    math.MaxInt - 1,
						Associativity: frontend.AssociativityLeft,
					},
					{
						Name:          "BAZ",
						Precedence:    math.MaxInt - 1,
						Associativity: frontend.AssociativityLeft,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
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

		It("should reject unknown terminals in precedence", func() {
			source := `
				@scanner {
				}
				@parser {
					@precedence {
						@left: FOO;
					}
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Fragments", func() {
		It("should exclude a fragment token from grammar terminals", func() {
			source := `
				@scanner {
					DIGIT: /[0-9]/ @fragment;
					NUMBER: /{DIGIT}+/;
				}
				@parser {
					file: NUMBER;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(1))
			Expect(grammar.Terminals[0].Name).To(Equal("NUMBER"))
		})

		It("should accept a token with a regex referencing a string-literal fragment", func() {
			source := `
				@scanner {
					KW_IF: "if" @fragment;
					IF: /{KW_IF}/ ;
				}
				@parser {
					file: IF;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(1))
			Expect(grammar.Terminals[0].Name).To(Equal("IF"))
		})

		It("should accept a token with a regex referencing an @empty fragment", func() {
			source := `
				@scanner {
					NOTHING: @empty @fragment;
					FOO: /{NOTHING}/;
				}
				@parser {
					file: FOO;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(1))
			Expect(grammar.Terminals[0].Name).To(Equal("FOO"))
		})

		It("should accept a token referencing a nested fragment", func() {
			source := `
				@scanner {
					DIGIT: /[0-9]/ @fragment;
					HEX_DIGIT: /[a-f]{DIGIT}/ @fragment;
					HEX: /{HEX_DIGIT}+/;
				}
				@parser {
					file: HEX;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(1))
			Expect(grammar.Terminals[0].Name).To(Equal("HEX"))
		})

		It("should accept multiple tokens referencing the same fragment", func() {
			source := `
				@scanner {
					DIGIT: /[0-9]/ @fragment;
					DEC: /{DIGIT}+/;
					OCT: /0[0-7]{DIGIT}*/;
				}
				@parser {
					file: DEC | OCT;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(2))
		})

		It("should reject a duplicate fragment declaration", func() {
			source := `
				@scanner {
					DIGIT: /[0-9]/ @fragment;
					DIGIT: /[0-9]/ @fragment;
				}
				@parser {
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject a fragment name that collides with a token name", func() {
			source := `
				@scanner {
					DIGIT: /[0-9]/;
					DIGIT: /[0-9]/ @fragment;
				}
				@parser {
					file: DIGIT;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject a token referencing an unknown fragment", func() {
			source := `
				@scanner {
					NUMBER: /{DIGIT}+/;
				}
				@parser {
					file: NUMBER;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject a cyclic fragment reference", func() {
			source := `
				@scanner {
					A: /{B}/ @fragment;
					B: /{A}/ @fragment;
					TOKEN: /{A}/;
				}
				@parser {
					file: TOKEN;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Error recovery", func() {
		It("should not add the error symbol to a grammar which does not reference it", func() {
			source := `
				@scanner {
					FOO: /foo/;
				}
				@parser {
					file: FOO;
				}
			`
			rules, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			for _, terminal := range grammar.Terminals {
				Expect(terminal.Name).ToNot(Equal(frontend.SymbolError.Name))
			}
			for _, rule := range rules {
				Expect(rule.Name).ToNot(Equal(frontend.SymbolError.Name))
			}
		})

		It("should add the error symbol as a terminal without a scanner rule when @error is referenced", func() {
			source := `
				@scanner {
					SEMI: ";";
				}
				@parser {
					file
						: SEMI
						| @error ";"
						;
				}
			`
			rules, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(2))
			Expect(grammar.Terminals[1].Name).To(Equal(frontend.SymbolError.Name))

			// No input can produce the error symbol, so it gets no scanner rule. The generated scanner declares the
			// constant which names it among its reserved tokens instead.
			Expect(rules).To(HaveLen(1))
			Expect(rules[0].Name).To(Equal("SEMI"))

			Expect(grammar.Productions[1].SymbolRefs).To(Equal([]frontend.SymbolRef{
				frontend.NewTerminalRef(1),
				frontend.NewTerminalRef(0),
			}))
		})

		It("should add the error symbol only once when @error is referenced repeatedly", func() {
			source := `
				@scanner {
					SEMI: ";";
				}
				@parser {
					file
						: @error ";"
						| file @error ";"
						;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals).To(HaveLen(2))
		})

		It("should accept @error in a precedence declaration", func() {
			// GNU Bison permits giving the error token a precedence and grammars in the wild rely on it, so GoLR has to
			// be able to express it as well.
			source := `
				@scanner {
					FOO: /foo/;
				}
				@parser {
					@precedence {
						@left: @error;
					}
					file: FOO;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Terminals[1].Name).To(Equal(frontend.SymbolError.Name))
			Expect(grammar.Terminals[1].Associativity).To(Equal(frontend.AssociativityLeft))
		})

		It("should not let a scanner section declare a terminal which collides with the error symbol", func() {
			// The reserved name carries a leading dollar sign, which the IDENTIFIER pattern does not allow, so a collision
			// can only be attempted and never succeed.
			source := `
				@scanner {
					$error: /foo/;
				}
				@parser {
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Productions", func() {
		It("should accept tokens with regex, literal string and empty on the right hand side", func() {
			source := `
				@scanner {
					FOO: /foo/;
					BAR: "bar";
					BAZ: @empty;
				}
				@parser {
					file: FOO "bar" BAZ;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
					{
						Name:  "BAR",
						Alias: `"bar"`,
					},
					{
						Name: "BAZ",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
							frontend.NewTerminalRef(1),
							frontend.NewTerminalRef(2),
						},
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept nonterminals on the right hand side", func() {
			source := `
				@scanner {
				}
				@parser {
					file: content;
					content: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
					{
						Name: "content",
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
						NonterminalIdx: 1,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept a mix of terminals and nonterminals on the right hand side", func() {
			source := `
				@scanner {
					FOO: @empty;
				}
				@parser {
					file: content FOO;
					content: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
					{
						Name: "content",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
							frontend.NewTerminalRef(0),
						},
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should reject production with a terminal name", func() {
			source := `
				@scanner {
					FOO: @empty;
				}
				@parser {
					FOO: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject production with undeclared nonterminal", func() {
			source := `
				@scanner {
				}
				@parser {
					file: content;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject production with undeclared terminal", func() {
			source := `
				@scanner {
				}
				@parser {
					file: "foo";
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Precedence of Productions", func() {
		It("should accept precedence on the right hand side", func() {
			source := `
				@scanner {
					FOO: /foo/;
					BAR: "bar";
					BAZ: @empty;
				}
				@parser {
					@precedence {
						@left: FOO;
						@right: "bar";
						@none: BAZ;
					}
					file: content @precedence(FOO);
					content: line @precedence("bar");
					line: BAZ @precedence(BAZ);
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			ptrTo0 := 0
			ptrTo1 := 1
			ptrTo2 := 2
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    math.MaxInt - 1,
					},
					{
						Name:          "BAR",
						Alias:         `"bar"`,
						Associativity: frontend.AssociativityRight,
						Precedence:    math.MaxInt - 2,
					},
					{
						Name:          "BAZ",
						Associativity: frontend.AssociativityNone,
						Precedence:    math.MaxInt - 3,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
					{
						Name: "content",
					},
					{
						Name: "line",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(1),
						},
						PrecedenceTerminalIdx: &ptrTo0,
					},
					{
						NonterminalIdx: 1,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewNonterminalRef(2),
						},
						PrecedenceTerminalIdx: &ptrTo1,
					},
					{
						NonterminalIdx: 2,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(2),
						},
						PrecedenceTerminalIdx: &ptrTo2,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept precedence on an empty alternative", func() {
			source := `
				@scanner {
					FOO: @empty;
				}
				@parser {
					@precedence {
						@left: FOO;
					}
					file: FOO | @empty @precedence(FOO);
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			ptrTo0 := 0
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name:          "FOO",
						Associativity: frontend.AssociativityLeft,
						Precedence:    math.MaxInt - 1,
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
						},
					},
					{
						NonterminalIdx:        0,
						SymbolRefs:            nil,
						PrecedenceTerminalIdx: &ptrTo0,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple alternatives", func() {
			source := `
				@scanner {
					FOO: @empty;
					BAR: @empty;
				}
				@parser {
					file: FOO | BAR;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
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
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
						},
					},
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

		It("should accept an empty alternative", func() {
			source := `
				@scanner {
					FOO: @empty;
				}
				@parser {
					file: FOO | @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: []frontend.Symbol{
					{
						Name: "FOO",
					},
				},
				Nonterminals: []frontend.Symbol{
					{
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
						},
					},
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 0,
			}))
		})

		It("should accept multiple alternatives as separate rules", func() {
			source := `
				@scanner {
					FOO: @empty;
					BAR: @empty;
				}
				@parser {
					file: FOO;
					file: BAR;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
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
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs: []frontend.SymbolRef{
							frontend.NewTerminalRef(0),
						},
					},
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

		It("should reject production precedence with undeclared terminal", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					file: "foo" @precedence(BAR);
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Names of Productions", func() {
		It("should set the explicit name of a production from an @name annotation", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					file: FOO @name(my_rule);
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Productions[0].Name).To(HaveValue(Equal("my_rule")))
		})

		It("should name each alternative independently", func() {
			source := `
				@scanner {
					FOO: "foo";
					BAR: "bar";
				}
				@parser {
					file
						: FOO @name(from_foo)
						| BAR @name(from_bar)
						;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Productions[0].Name).To(HaveValue(Equal("from_foo")))
			Expect(grammar.Productions[1].Name).To(HaveValue(Equal("from_bar")))
		})

		It("should leave a production without an @name annotation unnamed", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					file: FOO;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Productions[0].Name).To(BeNil())
		})

		It("should accept @name next to @precedence on the same alternative", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					@precedence {
						@left: FOO;
					}
					file: FOO @name(my_rule) @precedence(FOO);
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar.Productions[0].Name).To(HaveValue(Equal("my_rule")))
			Expect(grammar.Productions[0].PrecedenceTerminalIdx).ToNot(BeNil())
		})

		It("should reject two productions asking for the same name", func() {
			source := `
				@scanner {
					FOO: "foo";
					BAR: "bar";
				}
				@parser {
					file: FOO @name(dup);
					other: BAR @name(dup);
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`"dup"`))
		})

		It("should round trip an @name annotation through the GoLR writer", func() {
			source := `
				@scanner {
					FOO: "foo";
				}
				@parser {
					file: FOO @name(my_rule);
				}
			`
			rules, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())

			formatted, err := golr.GrammarToString(rules, grammar)
			Expect(err).ToNot(HaveOccurred())
			Expect(formatted).To(ContainSubstring("@name(my_rule)"))

			_, reparsed, err := golr.GrammarFromString(formatted)
			Expect(err).ToNot(HaveOccurred())
			Expect(reparsed.Productions[0].Name).To(HaveValue(Equal("my_rule")))
		})
	})

	Context("Start", func() {
		It("should respect a starting nonterminal", func() {
			source := `
				@scanner {
				}
				@parser {
					@start: file;

					content: @empty;
					file: @empty;
				}
			`
			_, grammar, err := golr.GrammarFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(grammar).To(Equal(frontend.Grammar{
				Terminals: nil,
				Nonterminals: []frontend.Symbol{
					{
						Name: "content",
					},
					{
						Name: "file",
					},
				},
				Productions: []frontend.Production{
					{
						NonterminalIdx: 0,
						SymbolRefs:     nil,
					},
					{
						NonterminalIdx: 1,
						SymbolRefs:     nil,
					},
				},
				StartNonterminalIdx: 1,
			}))
		})

		It("should reject nonexisting start nonterminal", func() {
			source := `
				@scanner {
				}
				@parser {
					@start: content;
					file: @empty;
				}
			`
			_, _, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})
})
