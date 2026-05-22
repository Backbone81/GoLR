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
		grammar, err := golr.GrammarFromString(source)
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
		_, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			grammar, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
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
			grammar, err := golr.GrammarFromString(source)
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
			_, err := golr.GrammarFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})
})
