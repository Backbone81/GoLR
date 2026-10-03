package counterexample_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/counterexample"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/utils"
)

// reduceReduceSpec is an ambiguous grammar with a reduce/reduce conflict, one of whose reductions is to the empty
// string.
var reduceReduceSpec = utils.HereDoc(`
	@scanner {
	    X: "x";
	    Z: "z";
	}

	@parser {
	    s : a "z" | b "z" ;
	    a : "x" e ;
	    b : "x" ;
	    e : @empty ;
	}
`)

var _ = Describe("Counterexample", func() {
	It("should render a unifying counterexample with a single example", func() {
		_, grammar, err := golrfrontend.GrammarFromString(figure1Spec)
		Expect(err).ToNot(HaveOccurred())
		dot := counterexample.NewDotDerivation()
		ifStmt := `stmt -> "if" expr "then" stmt`
		ifElseStmt := `stmt -> "if" expr "then" stmt "else" stmt`

		ce := counterexample.Counterexample{
			Unifying:       true,
			NonterminalIdx: nonterminalIdx(grammar, "stmt"),
			Derivations: [2]counterexample.Derivation{
				expand(grammar, ifElseStmt,
					leaf(grammar, `"if"`), leaf(grammar, "expr"), leaf(grammar, `"then"`),
					expand(grammar, ifStmt,
						leaf(grammar, `"if"`), leaf(grammar, "expr"), leaf(grammar, `"then"`), leaf(grammar, "stmt"), dot,
					),
					leaf(grammar, `"else"`), leaf(grammar, "stmt"),
				),
				expand(grammar, ifStmt,
					leaf(grammar, `"if"`), leaf(grammar, "expr"), leaf(grammar, `"then"`),
					expand(grammar, ifElseStmt,
						leaf(grammar, `"if"`), leaf(grammar, "expr"), leaf(grammar, `"then"`), leaf(grammar, "stmt"), dot,
						leaf(grammar, `"else"`), leaf(grammar, "stmt"),
					),
				),
			},
		}
		Expect(ce.Header(grammar)).To(Equal("ambiguous for stmt:"))
		Expect(strings.Join(ce.Lines(grammar), "\n") + "\n").To(Equal(utils.HereDoc(`
			example: "if" expr "then" "if" expr "then" stmt • "else" stmt
			using the reduction:
			  stmt -> "if" expr "then" [stmt] "else" stmt
			    stmt -> "if" expr "then" stmt •
			using the shift:
			  stmt -> "if" expr "then" [stmt]
			    stmt -> "if" expr "then" stmt • "else" stmt
		`)))
	})

	It("should render a nonunifying counterexample with an example per derivation", func() {
		_, grammar, err := golrfrontend.GrammarFromString(figure3Spec)
		Expect(err).ToNot(HaveOccurred())
		dot := counterexample.NewDotDerivation()

		ce := counterexample.Counterexample{
			NonterminalIdx: nonterminalIdx(grammar, "s"),
			Derivations: [2]counterexample.Derivation{
				expand(grammar, "s -> s t",
					expand(grammar, "s -> t",
						expand(grammar, "t -> x",
							expand(grammar, `x -> "a"`, leaf(grammar, `"a"`), dot),
						),
					),
					expand(grammar, "t -> x",
						expand(grammar, `x -> "a"`, leaf(grammar, `"a"`)),
					),
				),
				expand(grammar, "s -> t",
					expand(grammar, "t -> y",
						expand(grammar, `y -> "a" "a" "b"`, leaf(grammar, `"a"`), dot, leaf(grammar, `"a"`), leaf(grammar, `"b"`)),
					),
				),
			},
		}
		Expect(ce.Header(grammar)).To(Equal("conflict within s:"))
		Expect(strings.Join(ce.Lines(grammar), "\n") + "\n").To(Equal(utils.HereDoc(`
			example: "a" • "a"
			using the reduction:
			  s -> [s] [t]
			    s -> [t]
			      t -> [x]
			        x -> "a" •
			    t -> [x]
			      x -> "a"
			example: "a" • "a" "b"
			using the shift:
			  s -> [t]
			    t -> [y]
			      y -> "a" • "a" "b"
		`)))
	})

	It("should render a reduce/reduce conflict with a reduction to the empty string", func() {
		_, grammar, err := golrfrontend.GrammarFromString(reduceReduceSpec)
		Expect(err).ToNot(HaveOccurred())
		dot := counterexample.NewDotDerivation()

		ce := counterexample.Counterexample{
			Unifying:       true,
			NonterminalIdx: nonterminalIdx(grammar, "s"),
			Derivations: [2]counterexample.Derivation{
				expand(grammar, `s -> a "z"`,
					expand(grammar, `a -> "x" e`, leaf(grammar, `"x"`), expand(grammar, "e -> (empty)", dot)),
					leaf(grammar, `"z"`),
				),
				expand(grammar, `s -> b "z"`,
					expand(grammar, `b -> "x"`, leaf(grammar, `"x"`), dot),
					leaf(grammar, `"z"`),
				),
			},
		}
		Expect(strings.Join(ce.Lines(grammar), "\n") + "\n").To(Equal(utils.HereDoc(`
			example: "x" • "z"
			using the first reduction:
			  s -> [a] "z"
			    a -> "x" [e]
			      e -> (empty) •
			using the second reduction:
			  s -> [b] "z"
			    b -> "x" •
		`)))
	})
})

// expand returns the nonterminal of the production, written as FormatProduction writes it, expanded into the children.
func expand(
	grammar frontend.Grammar,
	production string,
	children ...counterexample.Derivation,
) counterexample.Derivation {
	return counterexample.NewExpandedDerivation(grammar, productionIdx(grammar, production), children)
}

// leaf returns the symbol with the name, which is not expanded.
func leaf(grammar frontend.Grammar, name string) counterexample.Derivation {
	for i, terminal := range grammar.Terminals {
		if terminal.String() == name {
			return counterexample.NewLeafDerivation(frontend.NewTerminalRef(i))
		}
	}
	return counterexample.NewLeafDerivation(frontend.NewNonterminalRef(nonterminalIdx(grammar, name)))
}
