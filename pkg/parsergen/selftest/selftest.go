package selftest

import intselftest "github.com/backbone81/golr/internal/parsergen/selftest"

type (
	// Config configures a Run. Every field with a sensible default may be left zero.
	Config = intselftest.Config

	// Progress is a cumulative snapshot of a run. Every counter covers the whole run so far, not the interval since
	// the previous snapshot.
	Progress = intselftest.Progress
)

var (
	// DefaultConfig provides the default configuration for the self test.
	DefaultConfig = intselftest.DefaultConfig

	// Run checks randomly generated grammars against the canonical LR(1) oracle until the context is cancelled, the
	// configured grammar count or duration is reached, or a failure ends the run under Config.StopOnFailure. It returns a
	// channel of cumulative Progress snapshots which is closed when the run ends, so a caller reads until the close and
	// reports the last snapshot it received as the summary of the run.
	Run = intselftest.Run
)
