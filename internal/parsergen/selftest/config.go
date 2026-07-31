package selftest

import "time"

// Config configures a self-test run.
type Config struct {
	// Workers is the number of grammars checked concurrently. Defaults to runtime.NumCPU(), which is what saturates a
	// machine; the check is CPU bound, so more workers than cores only adds scheduling overhead.
	Workers int

	// GrammarCount is how many grammars to check in total. Zero means unlimited, so the run only ends on Duration, on
	// cancellation of the context, or on a failure with StopOnFailure set.
	GrammarCount int

	// Duration is how long to keep checking grammars. Zero means unlimited.
	Duration time.Duration

	// SentencesPerGrammar is how many generated sentences each grammar is driven through. Defaults to the value of
	// DefaultConfig.
	SentencesPerGrammar int

	// FailureDir is the directory a failing grammar is dumped into, as a <seed>.golr grammar file next to a <seed>.log
	// holding the failure and the action traces. It is created if it does not exist. Defaults to the value of
	// DefaultConfig, which is the current directory.
	FailureDir string

	// StopOnFailure ends the whole run as soon as one grammar fails, rather than counting the failure and carrying on.
	StopOnFailure bool

	// MemoryLimitMiB is how much heap a run may fill before the garbage collector has to run. It is the garbage
	// collector, and not the parser core, which limits the throughput of a run: the parser tables of a grammar are
	// garbage the moment its comparison ends, so the live heap stays at a few megabytes, and the default GOGC of 100
	// therefore triggers a collection once the heap has grown to twice that - about a thousand times per second. Each
	// of those stops every core, which is what keeps a run from saturating the machine. Trading half a gigabyte of
	// memory for it gives roughly six times the throughput.
	//
	// Zero leaves the garbage collector at whatever the process is configured with, which is what an embedded run
	// wants. Set it to DefaultConfig.MemoryLimitMiB to get the tuning a long run benefits from; the CLI does that
	// through the default of its --memory-limit flag.
	MemoryLimitMiB int

	// The generator limits below shape the random grammars. Each one defaults to the value of
	// oracle.DefaultGrammarGenerator when left zero, which is tuned to produce small grammars whose canonical LR(1)
	// oracle is cheap to build. Raising them explores larger grammars at a steeply rising cost per grammar, because the
	// canonical LR(1) construction is what the runtime and the memory of a check are spent on.
	MaxTerminalCount                 int
	MaxNonterminalCount              int
	MaxProductionCountPerNonterminal int
	MaxRHSSymbolCount                int
}

// DefaultConfig provides the default config for the self-test.
var DefaultConfig = Config{
	Workers:                          0,
	GrammarCount:                     0,
	Duration:                         0,
	SentencesPerGrammar:              16,
	FailureDir:                       ".",
	StopOnFailure:                    false,
	MemoryLimitMiB:                   512,
	MaxTerminalCount:                 5,
	MaxNonterminalCount:              8,
	MaxProductionCountPerNonterminal: 6,
	MaxRHSSymbolCount:                4,
}
