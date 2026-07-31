package golr_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/selftest"
)

// The behavioral differential test drives selftest.CompareBehavior, which builds the resolved canonical LR(1) oracle
// table and the IELR(1) table under test for a grammar and compares them action for action on generated sentences. The
// comparison itself lives in the selftest package because the `golr selftest` command runs the very same check for
// hours across all cores; this test is the short, deterministic corpus which every `make test` run affords.
var _ = Describe("IELR(1) behavioral differential test", func() {
	// inputsPerGrammar is how many random sentences each grammar is checked with. A handful reaches most of the paths a
	// small grammar has; the corpus size below is the main lever for coverage, this one trades depth per grammar for
	// breadth across grammars.
	const inputsPerGrammar = 16

	// The paper's figures and the reduce/reduce grammar are the non-LALR shapes where phase 3 splitting fires or is
	// suppressed - the cases most likely to expose an IELR(1) bug. A failure names the entry it came from, so the
	// assertion needs no description of its own.
	DescribeTable(
		"should agree action for action with resolved canonical LR(1) on curated grammars",
		func(grammar frontend.Grammar) {
			_, err := selftest.CompareBehavior(grammar, inputsPerGrammar, rand.New(rand.NewSource(GinkgoRandomSeed())))
			Expect(err).ToNot(HaveOccurred())
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
			// table per grammar (the dominant cost), so keep it to a size which still finishes in a few seconds. Hunting
			// a suspected bug over a bigger corpus is what the `golr selftest` command is for.
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
				outcome, err := selftest.CompareBehavior(grammar, inputsPerGrammar, inputRng)
				Expect(err).ToNot(HaveOccurred(), "grammar seed %d:\n%s", grammarSeed, grammar.String())

				if !outcome.Compared {
					skipped++
					continue
				}
				compared++
				if outcome.Discriminating {
					discriminating++
				}
				if outcome.SplittingFired {
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

			// The discriminating grammars - those where LALR has a conflict canonical LR(1) does not - are the whole
			// point of the corpus: they are the non-LALR grammars where phase 3 splitting earns its keep. Passing a
			// corpus of only trivially-LALR grammars would exercise none of the splitting under test and pass vacuously,
			// so assert the corpus keeps clearing a floor of them. The generator yields roughly 65 of them per thousand
			// grammars (observed 49-74 across runs, a ~6.5% rate with a binomial standard deviation near 8); a floor of 15
			// sits several deviations below that mean - never flaky - while still catching the generator degrading toward
			// all-trivial grammars.
			Expect(discriminating).To(BeNumerically(">", 15))
		})
	})
})
