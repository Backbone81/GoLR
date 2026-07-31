package cmd

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/backbone81/golr/pkg/parsergen/selftest"
)

var (
	selftestConfig selftest.Config
)

var selftestCmd = &cobra.Command{
	Use:          "selftest",
	Short:        "Fuzz tests the IELR(1) parser core against a canonical LR(1) oracle.",
	Long:         `Fuzz tests the IELR(1) parser core against a canonical LR(1) oracle.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		startTimestamp := time.Now()

		log.Printf(
			"Self-test: workers=%d grammars=%d duration=%s sentences-per-grammar=%d memory-limit=%dMiB\n",
			selftestConfig.Workers,
			selftestConfig.GrammarCount,
			selftestConfig.Duration,
			selftestConfig.SentencesPerGrammar,
			selftestConfig.MemoryLimitMiB,
		)

		progress, err := selftest.Run(ctx, selftestConfig)
		if err != nil {
			return err
		}

		// Reading to the close of the channel is what delivers the final snapshot, so the loop always ends with the summary
		// of the whole run in hand.
		var finalState selftest.Progress
		for currentProgress := range progress {
			finalState = currentProgress
			log.Printf(
				"generated=%d compared=%d skipped=%d failed=%d discriminating=%d split=%d sentences=%d (%.1f grammars/s)\n",
				currentProgress.Generated,
				currentProgress.Compared,
				currentProgress.Skipped,
				currentProgress.Failed,
				currentProgress.Discriminating,
				currentProgress.SplittingFired,
				currentProgress.SentencesParsed,
				float64(currentProgress.Generated)/time.Since(startTimestamp).Seconds(),
			)
		}
		printSelftestSummary(finalState, time.Since(startTimestamp))

		if finalState.Failed > 0 {
			return fmt.Errorf("%d of %d grammars failed the check", finalState.Failed, finalState.Generated)
		}
		return nil
	},
}

// printSelftestSummary reports the totals of the finished run. The scenario breakdown is only printed here, not on
// every tick, because it is what a finished run is judged by rather than something to watch while it goes.
func printSelftestSummary(progress selftest.Progress, elapsed time.Duration) {
	fmt.Println()
	fmt.Printf("Self-test summary after %s:\n", elapsed)
	fmt.Printf("  %-34s %12d\n", "generated", progress.Generated)
	fmt.Printf("  %-34s %12d\n", "compared", progress.Compared)
	fmt.Printf("  %-34s %12d\n", "skipped", progress.Skipped)
	fmt.Printf("  %-34s %12d\n", "failed", progress.Failed)
	fmt.Printf("  %-34s %12d\n", "discriminating", progress.Discriminating)
	fmt.Printf("  %-34s %12d\n", "splitting fired", progress.SplittingFired)
	fmt.Printf("  %-34s %12d\n", "sentences parsed", progress.SentencesParsed)
}

func init() {
	rootCmd.AddCommand(selftestCmd)

	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.Workers,
		"workers",
		selftest.DefaultConfig.Workers,
		"The number of random grammars to check concurrently. Defaults to the number of CPU cores.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.GrammarCount,
		"grammar-count",
		selftest.DefaultConfig.GrammarCount,
		"The number of random grammars to check in total. Use 0 for unlimited.",
	)
	selftestCmd.PersistentFlags().DurationVar(
		&selftestConfig.Duration,
		"duration",
		selftest.DefaultConfig.Duration,
		"The duration to check random grammars. Supports Go durations with 30s or 10m. Use 0 for unlimited.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.SentencesPerGrammar,
		"sentences-per-grammar",
		selftest.DefaultConfig.SentencesPerGrammar,
		"The number of random sentences each grammar is checked with.",
	)
	selftestCmd.PersistentFlags().StringVar(
		&selftestConfig.FailureDir,
		"failure-dir",
		selftest.DefaultConfig.FailureDir,
		"The directory to dump a failing grammar and its action traces into.",
	)
	selftestCmd.PersistentFlags().BoolVar(
		&selftestConfig.StopOnFailure,
		"stop-on-failure",
		selftest.DefaultConfig.StopOnFailure,
		"Exit the application on the first failing grammar instead of continuing.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.MemoryLimitMiB,
		"memory-limit",
		selftest.DefaultConfig.MemoryLimitMiB,
		"The megabytes of heap to fill before collecting garbage. Higher values check more grammars per second at the "+
			"cost of memory. Use 0 to leave the Go garbage collector at its defaults.",
	)

	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.MaxTerminalCount,
		"max-terminal-count",
		selftest.DefaultConfig.MaxTerminalCount,
		"The largest number of terminals a random grammar may have.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.MaxNonterminalCount,
		"max-nonterminal-count",
		selftest.DefaultConfig.MaxNonterminalCount,
		"The largest number of nonterminals a random grammar may have.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.MaxProductionCountPerNonterminal,
		"max-production-count-per-nonterminal",
		selftest.DefaultConfig.MaxProductionCountPerNonterminal,
		"The largest number of productions a generated nonterminal may have.",
	)
	selftestCmd.PersistentFlags().IntVar(
		&selftestConfig.MaxRHSSymbolCount,
		"max-rhs-symbol-count",
		selftest.DefaultConfig.MaxRHSSymbolCount,
		"The largest number of symbols on the right hand side of a generated production.",
	)
}
