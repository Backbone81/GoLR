package golr_test

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lalr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// The behavioral differential test is the correct oracle for IELR(1): the parser table it produces is intentionally not
// isomorphic to LALR(1) or canonical LR(1) (different state count, numbering and splitting granularity), so a structural
// diff is the wrong tool. What IELR(1) does guarantee is behavioral — an IELR(1) parser accepts the same language and
// produces the same parses as canonical LR(1) under the same conflict-resolution policy. So the oracle is canonical
// LR(1), which is much simpler to build correctly (full LR(1) items, no splitting cleverness), and the test drives both
// resolved tables through the same generated sentences in lockstep, asserting they take the identical sequence of LR
// actions.
var _ = Describe("IELR(1) behavioral differential test", func() {
	// inputsPerGrammar is how many random sentences each grammar is checked with. A handful reaches most of the paths a
	// small grammar has; the corpus size below is the main lever for coverage, this one trades depth per grammar for
	// breadth across grammars.
	const inputsPerGrammar = 16

	// The system under test is ielr1golrcore.GrammarToParser, whose resolved table is compared, sentence by sentence,
	// against the resolved canonical LR(1) table for the same grammar under the same default conflict-resolution policy.
	// Both build from frontend.AugmentGrammar(grammar), so "reduce by production p" and "shift terminal t" mean the same
	// on both sides and the two action sequences are directly comparable.
	//
	// The paper's figures and the reduce/reduce grammar are the non-LALR shapes where phase 3 splitting fires or is
	// suppressed — the cases most likely to expose an IELR(1) bug.
	DescribeTable(
		"should agree action for action with resolved canonical LR(1) on curated grammars",
		func(grammar frontend.Grammar) {
			behaviorMatchesCanonicalLR1(grammar, inputsPerGrammar, rand.New(rand.NewSource(GinkgoRandomSeed())), "curated grammar")
		},
		Entry("the unambiguous test grammar for Fig. 1", ielr1golrcore.UnambiguousTestGrammarFig1),
		Entry("the ambiguous test grammar for Fig. 2", ielr1golrcore.AmbiguousTestGrammarFig2),
		Entry("the goto follows test grammar for Fig. 5", ielr1golrcore.GotoFollowsTestGrammarFig5),
		Entry("the goto follows caveats test grammar for Fig. 6", ielr1golrcore.GotoFollowsCaveatsTestGrammarFig6),
		Entry("the LR(1) but not LALR(1) reduce/reduce grammar", ielr1golrcore.ReduceReduceConflictTestGrammar),
	)

	// The random corpus is where the discriminating grammars come from: the generator's shared-nonterminal and
	// reduce/reduce scenarios are exactly the non-LALR shapes where canonical LR(1) and LALR(1) diverge, which is where
	// the split logic under test earns its keep. Each grammar is checked with a stream of derived sentences.
	Context("Random grammar corpus", func() {
		It("should agree action for action with resolved canonical LR(1) on a corpus of random grammars", func() {
			// grammarCount trades test time for coverage. The corpus runs under -race and builds a full canonical LR(1)
			// table per grammar (the dominant cost), so keep it to a size which still finishes in a few seconds; bump it
			// when hunting a suspected bug.
			const grammarCount = 1000

			var compared, skipped, discriminating, splittingFired int

			// A master RNG derives a distinct seed per grammar, so the corpus is grammarCount different grammars rather
			// than one repeated. The derived seed is reported on failure so a single failing grammar reconstructs on its
			// own by hand. Seeding the master from the Ginkgo random seed makes every run explore a fresh corpus while
			// staying reproducible with `ginkgo --seed=...`.
			masterRng := rand.New(rand.NewSource(GinkgoRandomSeed()))
			for range grammarCount {
				grammarSeed := masterRng.Int63()
				grammar := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(grammarSeed))).Generate()

				// The sentences for this grammar are drawn from an RNG seeded off the grammar seed, so a failing grammar
				// replays its exact sentence stream from the reported seed alone.
				inputRng := rand.New(rand.NewSource(grammarSeed))
				comparison := behaviorMatchesCanonicalLR1(grammar, inputsPerGrammar, inputRng, "grammar seed %d:\n%s", grammarSeed, grammar.String())
				if !comparison.compared {
					skipped++
					continue
				}
				compared++
				if comparison.discriminating {
					discriminating++
				}
				if comparison.splittingFired {
					splittingFired++
				}
			}

			GinkgoWriter.Printf(
				"random grammar corpus: %d compared, %d skipped (canonical LR(1) state limit), %d discriminating (LALR conflict LR(1) removes), %d split (|IELR| > |LALR|)\n",
				compared, skipped, discriminating, splittingFired,
			)

			// Guard against the generator degrading into grammars the oracle cannot build: if most grammars were skipped
			// the test would pass vacuously, which is the failure we care about the most.
			Expect(compared).To(BeNumerically(">", grammarCount/2))

			// The discriminating grammars — those where LALR has a conflict canonical LR(1) does not — are the whole
			// point of the corpus: they are the non-LALR grammars where phase 3 splitting earns its keep. Passing a
			// corpus of only trivially-LALR grammars would exercise none of the splitting under test and pass vacuously,
			// so assert the corpus keeps clearing a floor of them. The generator yields roughly 65 of them per thousand
			// grammars (observed 49–74 across runs, a ~6.5% rate with a binomial standard deviation near 8); a floor of 15
			// sits several deviations below that mean — never flaky — while still catching the generator degrading toward
			// all-trivial grammars.
			Expect(discriminating).To(BeNumerically(">", 15))
		})
	})
})

// grammarComparison reports what a single grammar contributed to the corpus. compared is false when the grammar was
// skipped (its canonical LR(1) automaton exceeded the addressable state limit), in which case the other fields are
// meaningless. discriminating marks a grammar where LALR(1) has a conflict canonical LR(1) does not — the non-LALR
// shapes the corpus exists to find. splittingFired marks a grammar where the IELR(1) table has more states than the
// LALR(1) table, i.e. phase 3 actually split a state.
type grammarComparison struct {
	compared       bool
	discriminating bool
	splittingFired bool
}

// behaviorMatchesCanonicalLR1 builds the resolved canonical LR(1) oracle table and the IELR(1) table under test for the
// grammar and drives both through inputsPerGrammar generated sentences in lockstep, asserting they take the identical
// sequence of LR actions on every one. It also builds the LALR(1) table so it can assert the state-count size invariant
// |LALR(1)| <= |IELR(1)| <= |canonical LR(1)| and report the corpus-coverage flags in the returned grammarComparison.
// A grammar whose canonical LR(1) automaton exceeds the addressable state limit is skipped (compared=false) rather than
// failed — the oracle cannot be built then. The description and args are woven into every assertion so a failure names
// the grammar it came from (a curated title or a corpus seed).
func behaviorMatchesCanonicalLR1(grammar frontend.Grammar, inputsPerGrammar int, inputRng *rand.Rand, description string, args ...any) grammarComparison {
	// The input generator and the interpreters speak the augmented alphabet, so augment once here for the generator; the
	// GrammarToParser calls below augment the grammar the same way internally.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	// The oracle: canonical LR(1), resolved with the same default policy IELR(1) uses (both go through their core's
	// GrammarToParser, which resolves conflicts under the hood). A grammar whose canonical LR(1) automaton is too large
	// to address is skipped, not a failure of the builder under test; any other error means conflict resolution failed,
	// which the default policy never should for a generated grammar (no precedence declarations), so asserting the error
	// is the state limit doubles as the plan's precondition that resolution does not error.
	// Both the oracle and the system under test are built without the default-reduction compaction: the test compares
	// them action for action, and a default reduction reduces where canonical LR(1) would report an error, on a
	// lookahead partition that differs between the two automata. That is a correct optimization (same language, same
	// parses, only the error is reported one or more reductions later), but it is not what this test is checking, so it
	// is switched off on both sides to keep the comparison on the canonical resolved tables.
	oracleParser, lr1Conflicts, err := lr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy, core.WithoutDefaultReductions())
	if err != nil {
		Expect(err).To(MatchError(backend.ErrStateLimitExceeded), append([]any{description}, args...)...)
		return grammarComparison{compared: false}
	}

	// The system under test: the IELR(1) table, resolved with the same policy by its GrammarToParser and, like the
	// oracle above, without the default-reduction compaction so the two are compared as canonical resolved tables.
	sutParser, _, err := ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy, core.WithoutDefaultReductions())
	Expect(err).ToNot(HaveOccurred(), append([]any{description}, args...)...)

	// The LALR(1) table, built the same way, is the lower bound of the size invariant and the source of the
	// discriminating signal. It is always no larger than canonical LR(1), so if the oracle built without hitting the
	// state limit this one does too; the default policy resolves every conflict of a generated grammar, so any error is
	// a real failure.
	lalrParser, lalrConflicts, err := lalr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)
	Expect(err).ToNot(HaveOccurred(), append([]any{description}, args...)...)

	// Size invariant |LALR(1)| <= |IELR(1)| <= |canonical LR(1)| (CLAUDE.md). Conflict resolution never adds or removes
	// states, so comparing the resolved tables is valid. An IELR(1) table larger than canonical LR(1) or smaller than
	// LALR(1) is a correctness-preserving quality bug — splitting too eagerly or losing a required split.
	Expect(len(sutParser.States)).To(
		BeNumerically(">=", len(lalrParser.States)),
		append([]any{"IELR(1) has fewer states than LALR(1): %s", description}, args...)...,
	)
	Expect(len(sutParser.States)).To(
		BeNumerically("<=", len(oracleParser.States)),
		append([]any{"IELR(1) has more states than canonical LR(1): %s", description}, args...)...,
	)

	comparison := grammarComparison{
		compared: true,
		// A grammar is discriminating when LALR(1) reports more conflicts than canonical LR(1): the surplus are the
		// mysterious LALR conflicts LR(1) removes, the shapes where phase 3 splitting matters. Comparing conflict counts
		// is a conservative proxy — it never over-counts a discriminating grammar — which is all a coverage metric needs.
		discriminating: len(lalrConflicts) > len(lr1Conflicts),
		splittingFired: len(sutParser.States) > len(lalrParser.States),
	}

	generator := oracle.NewInputGenerator(augmentedGrammar, inputRng)
	for range inputsPerGrammar {
		input := generator.Generate()

		// Both interpreters get the same runaway step bound, sized off the larger of the two tables. A cyclic grammar
		// (the generator can produce one, e.g. N -> N) makes both tables reduce forever; with a shared bound they cut
		// that identical loop off at the same step and read as the agreement it is, rather than diverging only because
		// the smaller IELR(1) table's default bound fires earlier. The input length includes the EOF each interpreter
		// appends.
		sharedMaxSteps := oracle.DefaultMaxSteps(len(input)+1, max(len(sutParser.States), len(oracleParser.States)))

		// Each interpreter appends its own EOF and mutates its own input cursor, so hand each a private copy of the
		// sentence to keep them fully independent.
		sutInterpreter := oracle.NewParserInterpreter(sutParser, slices.Clone(input), oracle.WithMaxSteps(sharedMaxSteps))
		oracleInterpreter := oracle.NewParserInterpreter(oracleParser, slices.Clone(input), oracle.WithMaxSteps(sharedMaxSteps))

		// a is the IELR(1) table under test, b is the canonical LR(1) oracle, matching the "a=" / "b=" labels of the
		// divergence message.
		if err := oracle.RunInLockstep(sutInterpreter, oracleInterpreter); err != nil {
			// On a divergence, replay both tables with tracing on so the failure carries the two full action traces:
			// reading them against each other is what pins down the state and lookahead where the IELR(1) table and the
			// canonical LR(1) oracle first parted ways.
			Fail(fmt.Sprintf(
				"%s\ninput %v\n%v\n\n=== IELR(1) trace ===\n%s\n=== canonical LR(1) trace ===\n%s",
				fmt.Sprintf(description, args...), input, err,
				traceParse(sutParser, input, sharedMaxSteps),
				traceParse(oracleParser, input, sharedMaxSteps),
			))
		}
	}
	return comparison
}

// traceParse runs the parser table over the input with tracing on and returns the recorded trace, for the divergence
// diagnostics. It drives the interpreter to completion; the interpreter itself writes the readable per-step lines.
func traceParse(parser backend.Parser, input []int, maxSteps int) string {
	var trace strings.Builder
	interpreter := oracle.NewParserInterpreter(
		parser, slices.Clone(input),
		oracle.WithMaxSteps(maxSteps),
		oracle.WithTrace(&trace),
	)
	for interpreter.Next() {
	}
	return trace.String()
}
