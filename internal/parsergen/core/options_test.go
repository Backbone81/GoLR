package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	intcore "github.com/backbone81/golr/internal/parsergen/core"
	"github.com/backbone81/golr/pkg/parsergen/backend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	ielr1bison "github.com/backbone81/golr/pkg/parsergen/core/ielr1/bison"
	ielr1golr "github.com/backbone81/golr/pkg/parsergen/core/ielr1/golr"
	lalr1bison "github.com/backbone81/golr/pkg/parsergen/core/lalr1/bison"
	lalr1golr "github.com/backbone81/golr/pkg/parsergen/core/lalr1/golr"
	lr1bison "github.com/backbone81/golr/pkg/parsergen/core/lr1/bison"
	lr1golr "github.com/backbone81/golr/pkg/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	"github.com/backbone81/golr/pkg/parsergen/frontend/dsl"
	"github.com/backbone81/golr/pkg/utils"
)

type grammarToParser func(frontend.Grammar, ...core.Option) (
	backend.Parser,
	[]conflict.Conflict,
	[]utils.Warning,
	error,
)

var (
	golrCores = map[string]grammarToParser{
		"ielr1-golr": ielr1golr.GrammarToParser,
		"lalr1-golr": lalr1golr.GrammarToParser,
		"lr1-golr":   lr1golr.GrammarToParser,
	}
	bisonCores = map[string]grammarToParser{
		"ielr1-bison": ielr1bison.GrammarToParser,
		"lalr1-bison": lalr1bison.GrammarToParser,
		"lr1-bison":   lr1bison.GrammarToParser,
	}
)

var _ = Describe("Options", func() {
	DescribeTable("should make the GoLR cores fail on the selected kinds of conflicts",
		func(grammar frontend.Grammar, options []core.Option, wantFail bool) {
			for coreName, grammarToParser := range golrCores {
				_, _, _, err := grammarToParser(grammar, options...)
				if wantFail {
					Expect(conflict.UnresolvedConflictErrors(err)).ToNot(BeEmpty(), "core %s", coreName)
				} else {
					Expect(err).ToNot(HaveOccurred(), "core %s", coreName)
				}
			}
		},
		Entry("a shift/reduce conflict without options", shiftReduceGrammar(), nil, false),
		Entry("a shift/reduce conflict failing on shift/reduce conflicts",
			shiftReduceGrammar(), []core.Option{core.FailOnShiftReduceConflicts()}, true),
		Entry("a shift/reduce conflict failing on reduce/reduce conflicts",
			shiftReduceGrammar(), []core.Option{core.FailOnReduceReduceConflicts()}, false),
		Entry("a shift/reduce conflict failing on conflicts",
			shiftReduceGrammar(), []core.Option{core.FailOnConflicts()}, true),
		Entry("a reduce/reduce conflict without options", reduceReduceGrammar(), nil, false),
		Entry("a reduce/reduce conflict failing on shift/reduce conflicts",
			reduceReduceGrammar(), []core.Option{core.FailOnShiftReduceConflicts()}, false),
		Entry("a reduce/reduce conflict failing on reduce/reduce conflicts",
			reduceReduceGrammar(), []core.Option{core.FailOnReduceReduceConflicts()}, true),
		Entry("a reduce/reduce conflict failing on conflicts",
			reduceReduceGrammar(), []core.Option{core.FailOnConflicts()}, true),
	)

	// GNU Bison resolves the conflicts itself, so the Bison cores cannot apply these options. The check happens before
	// GNU Bison is run, so these tests do not need it installed.
	DescribeTable("should make the Bison cores reject the options which fail on conflicts",
		func(option core.Option, wantMessage string) {
			for coreName, grammarToParser := range bisonCores {
				_, _, _, err := grammarToParser(shiftReduceGrammar(), option)
				Expect(err).To(MatchError(intcore.ErrOptionNotSupported), "core %s", coreName)
				Expect(err).To(MatchError(ContainSubstring(wantMessage)), "core %s", coreName)
			}
		},
		Entry("failing on shift/reduce conflicts", core.FailOnShiftReduceConflicts(),
			"failing on shift/reduce conflicts is not supported by the GNU Bison cores"),
		Entry("failing on reduce/reduce conflicts", core.FailOnReduceReduceConflicts(),
			"failing on reduce/reduce conflicts is not supported by the GNU Bison cores"),
		Entry("failing on conflicts", core.FailOnConflicts(),
			"failing on shift/reduce and reduce/reduce conflicts is not supported by the GNU Bison cores"),
	)
})

// shiftReduceGrammar returns an ambiguous grammar whose only conflicts are shift/reduce conflicts without precedence.
//
//	E -> E + E | id
func shiftReduceGrammar() frontend.Grammar {
	grammar := dsl.NewGrammar()
	plus := grammar.Terminal("+")
	identity := grammar.Terminal("id")
	expression := grammar.Nonterminal("E")
	grammar.Production(expression).Rhs(expression, plus, expression)
	grammar.Production(expression).Rhs(identity)
	return grammar.Build()
}

// reduceReduceGrammar returns an ambiguous grammar whose only conflict is a reduce/reduce conflict.
//
//	S -> A | B
//	A -> id
//	B -> id
func reduceReduceGrammar() frontend.Grammar {
	grammar := dsl.NewGrammar()
	identity := grammar.Terminal("id")
	start := grammar.Nonterminal("S")
	a := grammar.Nonterminal("A")
	b := grammar.Nonterminal("B")
	grammar.Production(start).Rhs(a)
	grammar.Production(start).Rhs(b)
	grammar.Production(a).Rhs(identity)
	grammar.Production(b).Rhs(identity)
	return grammar.Build()
}
