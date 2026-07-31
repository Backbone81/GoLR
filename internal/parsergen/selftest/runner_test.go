package selftest_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/selftest"
)

// The runner is the engine behind the `golr selftest` command, so these specs check the contract a caller depends on -
// the run ends on every one of its three stop conditions, the progress channel always closes, and the final snapshot
// adds up - rather than the correctness of the comparison itself, which the behavioral differential test covers. Every
// spec keeps the grammar count tiny, because a single grammar builds a full canonical LR(1) table and these run under
// -race like the rest of the suite.
var _ = Describe("Runner", func() {
	// smallRun is the configuration the specs start from: few grammars and few sentences each, which keeps a spec well
	// below the progress interval of the runner, so that the only snapshot a spec sees is the final one and the specs do
	// not depend on the timing of ticks. The failure directory is a temporary one, so that a spec which does hit a
	// failure dumps the grammar somewhere it can be picked up from instead of into the package directory.
	smallRun := func() selftest.Config {
		return selftest.Config{
			Workers:             2,
			GrammarCount:        20,
			SentencesPerGrammar: 2,
			FailureDir:          GinkgoT().TempDir(),
		}
	}

	It("should check the configured number of grammars and close the progress channel", func(ctx SpecContext) {
		progress, err := selftest.Run(ctx, smallRun())
		Expect(err).ToNot(HaveOccurred())

		final := drainProgress(progress)

		// The workers claim their slots from one shared counter, so the configured budget is what they produce between
		// them - no overshoot from the last grammars racing each other, and nothing left unchecked.
		Expect(final.Generated).To(BeEquivalentTo(20))

		// Every grammar either got compared against the oracle or was skipped because the oracle exceeded the state
		// limit. A grammar counted in neither would mean the aggregator lost a result.
		Expect(final.Compared + final.Skipped).To(Equal(final.Generated))
		Expect(final.Failed).To(BeZero())

		// The coverage flags are subsets of the compared grammars, never of the skipped ones.
		Expect(final.Discriminating).To(BeNumerically("<=", final.Compared))
		Expect(final.SplittingFired).To(BeNumerically("<=", final.Compared))
	})

	It("should drive every compared grammar through the configured number of sentences", func(ctx SpecContext) {
		progress, err := selftest.Run(ctx, smallRun())
		Expect(err).ToNot(HaveOccurred())

		final := drainProgress(progress)

		// The sentence counter comes from the comparison of every single grammar. A run reporting fewer sentences than
		// it was asked for would mean the per-grammar counts never reached the totals, which is what makes the
		// statistics of a soak run worthless. A skipped grammar contributes no sentences, and a failing one stops early,
		// so the equality only holds for a run without failures.
		Expect(final.Failed).To(BeZero())
		Expect(final.SentencesParsed).To(BeEquivalentTo(2 * final.Compared))
	})

	It("should end the run when the configured duration is up", func(ctx SpecContext) {
		config := smallRun()
		// Unlimited grammars, so the duration is the only thing which can end this run.
		config.GrammarCount = 0
		config.Duration = 100 * time.Millisecond

		progress, err := selftest.Run(ctx, config)
		Expect(err).ToNot(HaveOccurred())

		final := drainProgress(progress)
		Expect(final.Generated).To(BeNumerically(">", 0))
	})

	It("should end the run when the caller cancels the context", func(ctx SpecContext) {
		config := smallRun()
		// Unlimited grammars, so the cancellation is the only thing which can end this run.
		config.GrammarCount = 0

		runCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		progress, err := selftest.Run(runCtx, config)
		Expect(err).ToNot(HaveOccurred())

		cancel()

		// A cancelled run still delivers its final snapshot and closes the channel, which is what lets the command print
		// a summary after a Ctrl-C. The workers finish the grammar they are on, so the counts are whatever they reached.
		final := drainProgress(progress)
		Expect(final.Failed).To(BeZero())
	})

	It("should create the failure directory up front", func(ctx SpecContext) {
		config := smallRun()
		config.FailureDir = filepath.Join(GinkgoT().TempDir(), "failures")

		progress, err := selftest.Run(ctx, config)
		Expect(err).ToNot(HaveOccurred())
		defer drainProgress(progress)

		// The directory is prepared before the first grammar is checked, so that a run which will need to dump a grammar
		// finds out at the start instead of hours in, when it finally has something to report.
		Expect(config.FailureDir).To(BeADirectory())
	})

	It("should fail up front when the failure directory cannot be created", func(ctx SpecContext) {
		// A path below a regular file can never become a directory, which is the cheapest way to be certain the
		// preparation fails.
		blocker := filepath.Join(GinkgoT().TempDir(), "not-a-directory")
		Expect(os.WriteFile(blocker, []byte("blocker"), 0o600)).To(Succeed())

		config := smallRun()
		config.FailureDir = filepath.Join(blocker, "failures")

		progress, err := selftest.Run(ctx, config)
		Expect(err).To(HaveOccurred())
		Expect(progress).To(BeNil())
	})
})

// drainProgress reads the progress channel to its close and returns the final snapshot, which is the summary of the
// run. It fails the spec rather than blocking forever when the channel does not close, so a runner which loses its
// workers or never releases the aggregator shows up as a failure instead of a hung suite.
func drainProgress(progress <-chan selftest.Progress) selftest.Progress {
	GinkgoHelper()

	// The bound only has to be far above the runtime of a spec; it is a deadlock detector, not a performance
	// assertion.
	deadline := time.NewTimer(2 * time.Minute)
	defer deadline.Stop()

	var last selftest.Progress
	for {
		select {
		case snapshot, ok := <-progress:
			if !ok {
				return last
			}
			last = snapshot
		case <-deadline.C:
			Fail("the progress channel did not close in time")
		}
	}
}
