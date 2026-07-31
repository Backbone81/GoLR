package selftest

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lalr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// GrammarOutcome reports what a single grammar contributed to a corpus. Compared is false when the three tables could
// not all be built, either because the canonical LR(1) automaton exceeded the addressable state limit or because a
// construction failed outright; the other fields are then meaningless. Discriminating marks a grammar where LALR(1) has
// a conflict canonical LR(1) does not - the non-LALR shapes the corpus exists to find. SplittingFired marks a grammar
// where the IELR(1) table has more states than the LALR(1) table, i.e. phase 3 actually split a state. SentencesParsed
// counts the sentences the two tables were driven through, including the one which diverged when the comparison fails
// on a sentence.
//
// CompareBehavior fills the outcome in as far as the facts are known, on a failure as well, so a caller aggregating a
// corpus counts the work a failing grammar did rather than dropping it. Every field is valid whenever Compared is true,
// no matter whether an error came back with it.
type GrammarOutcome struct {
	Compared        bool
	Discriminating  bool
	SplittingFired  bool
	SentencesParsed int
}

// CompareBehavior is the correct oracle for IELR(1): the parser table it produces is intentionally not isomorphic to
// LALR(1) or canonical LR(1) (different state count, numbering and splitting granularity), so a structural diff is the
// wrong tool. What IELR(1) does guarantee is behavioral - an IELR(1) parser accepts the same language and produces the
// same parses as canonical LR(1) under the same conflict-resolution policy. So the oracle is canonical LR(1), which is
// much simpler to build correctly (full LR(1) items, no splitting cleverness), and both resolved tables are driven
// through the same generated sentences in lockstep, asserting they take the identical sequence of LR actions.
//
// It builds the resolved canonical LR(1) oracle table and the IELR(1) table under test for the grammar and drives both
// through inputsPerGrammar generated sentences, with the sentences drawn from rng. It also builds the LALR(1) table so
// it can check the state-count size invariant |LALR(1)| <= |IELR(1)| <= |canonical LR(1)| and report the corpus
// coverage flags in the returned GrammarOutcome.
//
// A grammar whose canonical LR(1) automaton exceeds the addressable state limit is skipped - reported as a zero
// GrammarOutcome and a nil error - rather than failed, because the oracle cannot be built then. Every other problem is
// returned as an error: a divergence between the two tables, a violated size invariant, or a table which failed to
// build. The divergence error carries the offending sentence and the full action trace of both tables, which is what
// pins down the state and the lookahead where they first parted ways.
//
// The returned GrammarOutcome is populated as far as the comparison got, an error notwithstanding, so that a caller
// aggregating statistics over a corpus does not lose the sentences a failing grammar was driven through.
//
// The grammar is un-augmented, the same as any grammar handed to a core.
func CompareBehavior(grammar frontend.Grammar, inputsPerGrammar int, rng *rand.Rand) (GrammarOutcome, error) {
	tables, built, err := buildComparisonTables(grammar)
	if err != nil || !built {
		return GrammarOutcome{}, err
	}

	// Every table is built now, so the coverage flags are known and the outcome is filled in before anything else can
	// fail. From here on it travels out with whatever error comes back, so a failing grammar still reports what it
	// contributed instead of vanishing from the corpus statistics.
	outcome := GrammarOutcome{
		Compared:       true,
		Discriminating: tables.discriminating,
		SplittingFired: len(tables.sutParser.States) > len(tables.lalrParser.States),
	}
	if err := tables.checkSizeInvariant(); err != nil {
		return outcome, err
	}

	// The input generator speaks the augmented alphabet, so augment for it here; the constructions above augment the
	// grammar the same way internally.
	generator := oracle.NewInputGenerator(frontend.AugmentGrammar(grammar), rng)
	for range inputsPerGrammar {
		if err := tables.compareOnSentence(generator.Generate()); err != nil {
			outcome.SentencesParsed++
			return outcome, err
		}
		outcome.SentencesParsed++
	}
	return outcome, nil
}

// comparisonTables holds the three resolved parser tables a comparison is built from, together with the conflict-count
// verdict which only the construction sees.
type comparisonTables struct {
	// oracleParser is the canonical LR(1) table the comparison trusts.
	oracleParser backend.Parser

	// sutParser is the IELR(1) table under test.
	sutParser backend.Parser

	// lalrParser is the LALR(1) table, which is the lower bound of the size invariant.
	lalrParser backend.Parser

	// discriminating is true when LALR(1) reports more conflicts than canonical LR(1): the surplus are the mysterious
	// LALR conflicts LR(1) removes, the shapes where phase 3 splitting matters. Comparing conflict counts is a
	// conservative proxy - it never over-counts a discriminating grammar - which is all a coverage metric needs.
	discriminating bool
}

// buildComparisonTables constructs the three resolved parser tables for the grammar. It reports built as false, with a
// nil error, when the grammar has to be skipped because its canonical LR(1) automaton exceeds the addressable state
// limit and the oracle can therefore not be built at all.
func buildComparisonTables(grammar frontend.Grammar) (comparisonTables, bool, error) {
	// The oracle: canonical LR(1), resolved with the same default policy IELR(1) uses (both go through their core's
	// GrammarToParser, which resolves conflicts under the hood). A grammar whose canonical LR(1) automaton is too large
	// to address is skipped, not a failure of the builder under test; any other error means conflict resolution failed,
	// which the default policy never should for a generated grammar (no precedence declarations).
	// Both the oracle and the system under test are built without the default-reduction compaction: the comparison is
	// action for action, and a default reduction reduces where canonical LR(1) would report an error, on a lookahead
	// partition that differs between the two automata. That is a correct optimization (same language, same parses, only
	// the error is reported one or more reductions later), but it is not what this comparison is checking, so it is
	// switched off on both sides to keep the comparison on the canonical resolved tables.
	oracleParser, lr1Conflicts, err := lr1golrcore.GrammarToParser(
		grammar, conflict.DefaultPolicy, core.WithoutDefaultReductions(),
	)
	if err != nil {
		if errors.Is(err, backend.ErrStateLimitExceeded) {
			return comparisonTables{}, false, nil
		}
		return comparisonTables{}, false, fmt.Errorf("building the canonical LR(1) oracle: %w", err)
	}

	// The system under test: the IELR(1) table, resolved with the same policy by its GrammarToParser and, like the
	// oracle above, without the default-reduction compaction so the two are compared as canonical resolved tables.
	sutParser, _, err := ielr1golrcore.GrammarToParser(
		grammar, conflict.DefaultPolicy, core.WithoutDefaultReductions(),
	)
	if err != nil {
		return comparisonTables{}, false, fmt.Errorf("building the IELR(1) parser under test: %w", err)
	}

	// The LALR(1) table, built the same way, is the lower bound of the size invariant and the source of the
	// discriminating signal. It is always no larger than canonical LR(1), so if the oracle built without hitting the
	// state limit this one does too; the default policy resolves every conflict of a generated grammar, so any error is
	// a real failure.
	lalrParser, lalrConflicts, err := lalr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)
	if err != nil {
		return comparisonTables{}, false, fmt.Errorf("building the LALR(1) parser: %w", err)
	}

	return comparisonTables{
		oracleParser:   oracleParser,
		sutParser:      sutParser,
		lalrParser:     lalrParser,
		discriminating: len(lalrConflicts) > len(lr1Conflicts),
	}, true, nil
}

// checkSizeInvariant verifies |LALR(1)| <= |IELR(1)| <= |canonical LR(1)|. Conflict resolution never adds or removes
// states, so comparing the resolved tables is valid. An IELR(1) table larger than canonical LR(1) or smaller than
// LALR(1) is a correctness-preserving quality bug - splitting too eagerly or losing a required split.
func (t comparisonTables) checkSizeInvariant() error {
	if len(t.sutParser.States) < len(t.lalrParser.States) {
		return fmt.Errorf(
			"IELR(1) has fewer states than LALR(1): %d < %d",
			len(t.sutParser.States), len(t.lalrParser.States),
		)
	}
	if len(t.sutParser.States) > len(t.oracleParser.States) {
		return fmt.Errorf(
			"IELR(1) has more states than canonical LR(1): %d > %d",
			len(t.sutParser.States), len(t.oracleParser.States),
		)
	}
	return nil
}

// compareOnSentence drives the table under test and the oracle through the sentence in lockstep and reports a
// divergence as an error carrying the sentence and the full action trace of both tables.
func (t comparisonTables) compareOnSentence(input []int) error {
	// Both interpreters get the same runaway step bound, sized off the larger of the two tables. A cyclic grammar (the
	// generator can produce one, e.g. N -> N) makes both tables reduce forever; with a shared bound they cut that
	// identical loop off at the same step and read as the agreement it is, rather than diverging only because the
	// smaller IELR(1) table's default bound fires earlier. The input length includes the EOF each interpreter appends.
	maxSteps := oracle.DefaultMaxSteps(len(input)+1, max(len(t.sutParser.States), len(t.oracleParser.States)))

	// Each interpreter appends its own EOF and mutates its own input cursor, so hand each a private copy of the
	// sentence to keep them fully independent.
	sutInterpreter := oracle.NewParserInterpreter(t.sutParser, slices.Clone(input), oracle.WithMaxSteps(maxSteps))
	oracleInterpreter := oracle.NewParserInterpreter(t.oracleParser, slices.Clone(input), oracle.WithMaxSteps(maxSteps))

	// a is the IELR(1) table under test, b is the canonical LR(1) oracle, matching the "a=" / "b=" labels of the
	// divergence message.
	err := oracle.RunInLockstep(sutInterpreter, oracleInterpreter)
	if err == nil {
		return nil
	}

	// On a divergence, replay both tables with tracing on so the failure carries the two full action traces: reading
	// them against each other is what pins down the state and lookahead where the IELR(1) table and the canonical LR(1)
	// oracle first parted ways. Tracing is only paid for on a failure, so a passing corpus carries no cost for it.
	return fmt.Errorf(
		"input %v\n%w\n\n=== IELR(1) trace ===\n%s\n=== canonical LR(1) trace ===\n%s",
		input, err,
		traceParse(t.sutParser, input, maxSteps),
		traceParse(t.oracleParser, input, maxSteps),
	)
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
