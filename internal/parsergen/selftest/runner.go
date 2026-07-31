package selftest

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// Run checks randomly generated grammars against the canonical LR(1) oracle until the context is cancelled, the
// configured grammar count or duration is reached, or a failure ends the run under Config.StopOnFailure. The work is
// spread over Config.Workers goroutines, so a run saturates as many cores as it is given.
//
// It returns a channel of cumulative Progress snapshots, one every few seconds and a final one when the run ends, after
// which the channel is closed. The final snapshot is the summary of the run: a caller reads until the channel closes
// and reports the last value it received, whose Failed count tells it whether anything was found. Snapshots are dropped
// rather than queued when the caller is not keeping up, so a slow reader cannot slow the run down; only the final
// snapshot is guaranteed to be delivered, which means the caller has to drain the channel to the close.
//
// An error is returned only for a configuration which cannot work, in particular a FailureDir which cannot be written
// to. That check happens up front on purpose: discovering it when the first grammar fails could be hours into a run.
func Run(ctx context.Context, config Config) (<-chan Progress, error) {
	return NewRunner(config).Run(ctx)
}

// Runner is one self-test run. It is used once: Run closes both channels when the run ends, so a second run needs a
// second Runner.
type Runner struct {
	config Config

	// progress carries the cumulative snapshots out to the caller, results carries the per grammar counts of the
	// workers in to the aggregator. Only the aggregator touches the totals, which is what keeps them free of locks.
	progress chan Progress
	results  chan Progress

	waitGroup sync.WaitGroup

	// grammarCounter is the shared budget of Config.GrammarCount. Every worker claims its next grammar from it, so
	// the workers produce the configured number of grammars between them instead of each producing that many.
	grammarCounter atomic.Int64
}

// NewRunner creates a runner for the config, filling in the default for every field which was left zero and has one.
func NewRunner(config Config) *Runner {
	if config.Workers == 0 {
		config.Workers = runtime.NumCPU()
	}
	if config.SentencesPerGrammar == 0 {
		config.SentencesPerGrammar = DefaultConfig.SentencesPerGrammar
	}
	if config.FailureDir == "" {
		config.FailureDir = DefaultConfig.FailureDir
	}
	return &Runner{
		config:   config,
		progress: make(chan Progress, 1),
		results:  make(chan Progress, config.Workers),
	}
}

// Run starts the workers and the aggregator and returns immediately with the progress channel of the run, which comes
// with the contract described on the package level Run.
func (r *Runner) Run(ctx context.Context) (<-chan Progress, error) {
	// The failure directory is prepared up front, so that a run which cannot write there finds out at the start
	// instead of hours in, when it finally has something to report.
	//nolint:gosec // The dumped grammars do not contain sensitive information.
	if err := os.MkdirAll(r.config.FailureDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating the failure directory %q: %w", r.config.FailureDir, err)
	}

	// Letting the heap grow to the limit before collecting is what lets a run saturate the cores it was given, see
	// Config.MemoryLimitMiB for why. Switching the growth target off is safe next to a limit the collector has to
	// respect: the limit is a soft one, so a run whose live heap ever approaches it collects more often rather than
	// running out of memory.
	if r.config.MemoryLimitMiB > 0 {
		debug.SetMemoryLimit(int64(r.config.MemoryLimitMiB) * 1024 * 1024)
		debug.SetGCPercent(-1)
	}

	var runCtx context.Context
	var cancel context.CancelFunc
	if r.config.Duration > 0 {
		runCtx, cancel = context.WithTimeout(ctx, r.config.Duration)
	} else {
		runCtx, cancel = context.WithCancel(ctx)
	}

	//nolint:gosec // We do not need crypographically strong random numbers here.
	runRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for range r.config.Workers {
		// The seed is drawn here and not inside the worker, because a rand.Rand is not safe for concurrent use and
		// this one is shared by all the workers being started.
		seed := runRand.Int63()
		r.waitGroup.Go(func() {
			r.runWorker(runCtx, cancel, seed)
		})
	}

	// Closing the results channel is what tells the aggregator that the run is over, and only the last worker
	// finishing may do it. The aggregator has to stay out of the wait group for that, because it keeps draining
	// results while the workers finish.
	go func() {
		r.waitGroup.Wait()
		close(r.results)
	}()

	go r.runAggregator(cancel)

	return r.progress, nil
}

// runWorker checks grammars until the run ends. Each grammar is built from a fresh seed drawn from the worker's own
// random source, and that one seed is enough to reconstruct both the grammar and the sentences it was checked with, so
// a failure is reproducible from the seed alone no matter how many workers were running.
func (r *Runner) runWorker(ctx context.Context, cancel context.CancelFunc, seed int64) {
	//nolint:gosec // We do not need crypographically strong random numbers here.
	workerRng := rand.New(rand.NewSource(seed))
	for {
		if ctx.Err() != nil {
			return
		}

		// Increment the grammar counter before doing any work. That way we can exit when we overshot the limit.
		if r.config.GrammarCount > 0 && r.grammarCounter.Add(1) > int64(r.config.GrammarCount) {
			return
		}

		//nolint:contextcheck // We do not need to pass a context only for creating the trace down in the call tree.
		result, err := r.testSingleGrammar(workerRng.Int63())

		if err != nil && r.config.StopOnFailure {
			// Report the failure before ending the run, so the aggregator counts it and the final snapshot describes it.
			select {
			case r.results <- result:
			case <-ctx.Done():
			}
			cancel()
			return
		}

		// Submit the result or wait for the context to be done.
		select {
		case r.results <- result:
		case <-ctx.Done():
			return
		}
	}
}

// runAggregator owns the totals of the run. Every worker reports through the results channel and only this goroutine
// adds them up, which keeps the counters free of locks and atomics. It emits a snapshot on every tick and a final one
// when the workers are done, then closes the progress channel.
func (r *Runner) runAggregator(cancel context.CancelFunc) {
	defer cancel()
	defer close(r.progress)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var totals Progress
	for {
		select {
		case result, ok := <-r.results:
			if !ok {
				// The workers are all done. We send the totals and are done.
				r.progress <- totals
				return
			}
			totals.Add(result)
		case <-ticker.C:
			select {
			case r.progress <- totals:
			default:
				// Do not block when reader is not ready. Drop it and try again on the next tick.
			}
		}
	}
}

// testSingleGrammar generates the grammar for the seed and runs the behavioral comparison on it. The sentence RNG is
// seeded from the same seed as the grammar, so the seed alone replays the whole check.
func (r *Runner) testSingleGrammar(seed int64) (Progress, error) {
	//nolint:gosec // We do not need crypographically strong random numbers here.
	rng := rand.New(rand.NewSource(seed))

	grammarGenerator := oracle.DefaultGrammarGenerator(rng)
	if r.config.MaxTerminalCount > 0 {
		grammarGenerator.MaxTerminalCount = r.config.MaxTerminalCount
	}
	if r.config.MaxNonterminalCount > 0 {
		grammarGenerator.MaxNonterminalCount = r.config.MaxNonterminalCount
	}
	if r.config.MaxProductionCountPerNonterminal > 0 {
		grammarGenerator.MaxProductionCountPerNonterminal = r.config.MaxProductionCountPerNonterminal
	}
	if r.config.MaxRHSSymbolCount > 0 {
		grammarGenerator.MaxRHSSymbolCount = r.config.MaxRHSSymbolCount
	}
	randomGrammar := grammarGenerator.Generate()

	outcome, err := CompareBehavior(randomGrammar, r.config.SentencesPerGrammar, rng)

	result := Progress{
		Generated:       1,
		Compared:        boolToInt64(outcome.Compared),
		Skipped:         boolToInt64(!outcome.Compared),
		Failed:          boolToInt64(err != nil),
		Discriminating:  boolToInt64(outcome.Discriminating),
		SplittingFired:  boolToInt64(outcome.SplittingFired),
		SentencesParsed: int64(outcome.SentencesParsed),
	}
	if err != nil {
		log.Printf("ERROR: Grammar test failed for seed %d: %v\n", seed, err)
		if dumpErr := r.dumpFailure(seed, randomGrammar, err); dumpErr != nil {
			log.Printf("ERROR: %s", dumpErr)
		}
	}
	return result, err
}

// dumpFailure writes the failing grammar and the failure description into the failure directory and returns the message
// describing what happened and where it went. Dumping is best effort: a run which cannot write the files still reports
// the failure and the seed, which is enough to reconstruct the grammar.
func (r *Runner) dumpFailure(seed int64, grammar frontend.Grammar, cause error) error {
	grammarFilePath := filepath.Join(r.config.FailureDir, fmt.Sprintf("%d.golr", seed))
	logFilePath := filepath.Join(r.config.FailureDir, fmt.Sprintf("%d.log", seed))

	if err := golrfrontend.GrammarToFile(grammarFilePath, nil, grammar); err != nil {
		return fmt.Errorf("writing grammar to file: %w", err)
	}
	log.Printf("Failed grammar written to %q\n", grammarFilePath)

	logContent := fmt.Sprintf("grammar seed %d\n\n%s\n%v\n", seed, grammar.String(), cause)
	//nolint:gosec // The failure report is meant to be read by others, it holds no sensitive information.
	if err := os.WriteFile(logFilePath, []byte(logContent), 0o644); err != nil {
		return fmt.Errorf("writing log to file: %w", err)
	}
	log.Printf("Failed grammar log written to %q\n", logFilePath)
	return nil
}

func boolToInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
