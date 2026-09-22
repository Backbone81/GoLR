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
// Both tables are built with the policy factory, which is how a run checks a policy that leaves conflicts unresolved,
// see conflict.SelectPolicy. Under such a policy a grammar can fail to generate, which is not a failure of the
// comparison: the oracle then is that canonical LR(1) fails on the same grammar with the same conflicts, see
// compareUnresolvedConflicts. Only a grammar whose tables both build is driven through sentences.
//
// The grammar is un-augmented, the same as any grammar handed to a core.
func CompareBehavior(
	grammar frontend.Grammar,
	inputsPerGrammar int,
	rng *rand.Rand,
	policyFactory conflict.PolicyFactory,
) (GrammarOutcome, error) {
	tables, built, err := buildComparisonTables(grammar, policyFactory)
	if err != nil || !built {
		// A grammar whose conflicts were compared counts as compared even when that comparison is what failed. Only a
		// grammar which could not be built at all is a skipped one.
		return GrammarOutcome{Compared: built}, err
	}
	if tables.unresolved {
		// Neither table exists, because both cores gave up on the conflicts they were left with. The conflicts were
		// compared while building, so there is nothing left to do for this grammar.
		return GrammarOutcome{Compared: true, Discriminating: tables.discriminating}, nil
	}

	// Every table is built now, so the coverage flags are known and the outcome is filled in before anything else can
	// fail. From here on it travels out with whatever error comes back, so a failing grammar still reports what it
	// contributed instead of vanishing from the corpus statistics.
	outcome := GrammarOutcome{
		Compared:       true,
		Discriminating: tables.discriminating,
		SplittingFired: len(tables.sutParser.States) > tables.lalrStateCount,
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

	// lalrParser is the resolved LALR(1) table, which is the lower bound of the size invariant. It is only there when
	// hasLalrParser is set: under a policy which leaves conflicts unresolved, LALR(1) fails on the grammars whose
	// mysterious conflicts IELR(1) removes, which is expected rather than a failure.
	lalrParser    backend.Parser
	hasLalrParser bool

	// lalrStateCount is the number of states of the LALR(1) automaton, which is what tells whether phase 3 split a
	// state. It is taken from the unresolved table when the resolved one does not exist, because those grammars are
	// where splitting matters most.
	lalrStateCount int

	// unresolved reports that both cores failed on unresolved conflicts, so there are no tables to compare any further.
	unresolved bool

	// discriminating is true when LALR(1) reports more conflicts than canonical LR(1): the surplus are the mysterious
	// LALR conflicts LR(1) removes, the shapes where phase 3 splitting matters. Comparing conflict counts is a
	// conservative proxy - it never over-counts a discriminating grammar - which is all a coverage metric needs.
	discriminating bool
}

// buildComparisonTables constructs the resolved parser tables for the grammar under the policy. It reports built as
// false, with a nil error, when the grammar has to be skipped because its canonical LR(1) automaton exceeds the
// addressable state limit and the oracle can therefore not be built at all.
//
// Under a policy which leaves conflicts unresolved, a core fails on a grammar it cannot resolve instead of returning a
// table. The oracle for that is canonical LR(1): IELR(1) has to fail on exactly the grammars canonical LR(1) fails on,
// with the same conflicts, which is what compareUnresolvedConflicts checks.
func buildComparisonTables(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (comparisonTables, bool, error) {
	// The oracle: canonical LR(1), resolved with the same policy IELR(1) is built with (both go through their core's
	// GrammarToParser, which resolves conflicts under the hood). A grammar whose canonical LR(1) automaton is too large
	// to address is skipped, not a failure of the builder under test.
	// Both the oracle and the system under test are built without the default-reduction compaction: the comparison is
	// action for action, and a default reduction reduces where canonical LR(1) would report an error, on a lookahead
	// partition that differs between the two automata. That is a correct optimization (same language, same parses, only
	// the error is reported one or more reductions later), but it is not what this comparison is checking, so it is
	// switched off on both sides to keep the comparison on the canonical resolved tables.
	oracleParser, lr1Conflicts, oracleErr := lr1golrcore.GrammarToParser(
		grammar, policyFactory, core.WithoutDefaultReductions(),
	)
	if oracleErr != nil && errors.Is(oracleErr, backend.ErrStateLimitExceeded) {
		return comparisonTables{}, false, nil
	}
	if oracleErr != nil && !isUnresolvedConflictError(oracleErr) {
		return comparisonTables{}, false, fmt.Errorf("building the canonical LR(1) oracle: %w", oracleErr)
	}

	// The system under test: the IELR(1) table, resolved with the same policy by its GrammarToParser and, like the
	// oracle above, without the default-reduction compaction so the two are compared as canonical resolved tables.
	sutParser, _, sutErr := ielr1golrcore.GrammarToParser(
		grammar, policyFactory, core.WithoutDefaultReductions(),
	)
	if sutErr != nil && !isUnresolvedConflictError(sutErr) {
		return comparisonTables{}, false, fmt.Errorf("building the IELR(1) parser under test: %w", sutErr)
	}

	if err := compareUnresolvedConflicts(oracleErr, sutErr); err != nil {
		return comparisonTables{}, true, err
	}

	// The LALR(1) table, built the same way, is the lower bound of the size invariant and the source of the
	// discriminating signal. It is always no larger than canonical LR(1), so if the oracle built without hitting the
	// state limit this one does too. Under a policy which leaves conflicts unresolved it fails on the grammars whose
	// mysterious conflicts only IELR(1) and canonical LR(1) get rid of, which is what makes such a grammar
	// discriminating rather than a failure.
	lalrParser, lalrConflicts, lalrErr := lalr1golrcore.GrammarToParser(grammar, policyFactory)
	if lalrErr != nil && !isUnresolvedConflictError(lalrErr) {
		return comparisonTables{}, false, fmt.Errorf("building the LALR(1) parser: %w", lalrErr)
	}
	lalrStateCount := len(lalrParser.States)
	if lalrErr != nil {
		// The resolved table does not exist, but the automaton it would have been built from does, and its state count
		// is all the split signal needs. Resolving the conflicts is also what removes the unreachable states, so this
		// count is not the lower bound of the size invariant.
		unresolvedLalrParser, err := lalr1golrcore.GrammarToUnresolvedParser(grammar, policyFactory)
		if err != nil {
			return comparisonTables{}, false, fmt.Errorf("building the unresolved LALR(1) parser: %w", err)
		}
		lalrStateCount = len(unresolvedLalrParser.States)
	}

	return comparisonTables{
		oracleParser:   oracleParser,
		sutParser:      sutParser,
		lalrParser:     lalrParser,
		hasLalrParser:  lalrErr == nil,
		lalrStateCount: lalrStateCount,
		unresolved:     oracleErr != nil,
		// The conflicts a core reports come back with the error as well, so this reads the same for a grammar which
		// failed to generate as for one which did not.
		discriminating: len(lalrConflicts) > len(lr1Conflicts),
	}, true, nil
}

// isUnresolvedConflictError reports whether the error is a core giving up on the conflicts a policy left unresolved,
// rather than something which went wrong.
func isUnresolvedConflictError(err error) bool {
	return len(conflict.UnresolvedConflictErrors(err)) > 0
}

// compareUnresolvedConflicts checks that IELR(1) gave up on the grammar exactly when canonical LR(1) did, and on the
// same conflicts. A conflict which only one of them reports means that IELR(1) either lost a conflict canonical LR(1)
// has, or invented one it does not have.
func compareUnresolvedConflicts(oracleErr error, sutErr error) error {
	if (oracleErr == nil) != (sutErr == nil) {
		return fmt.Errorf(
			"IELR(1) and canonical LR(1) disagree on whether the grammar can be generated:"+
				" canonical LR(1) error is %v, IELR(1) error is %v",
			oracleErr, sutErr,
		)
	}
	oracleConflicts := unresolvedConflictKeys(oracleErr)
	sutConflicts := unresolvedConflictKeys(sutErr)
	if !slices.Equal(oracleConflicts, sutConflicts) {
		return fmt.Errorf(
			"IELR(1) and canonical LR(1) report different unresolved conflicts:"+
				"\n=== canonical LR(1) ===\n%s\n=== IELR(1) ===\n%s",
			strings.Join(oracleConflicts, "\n"), strings.Join(sutConflicts, "\n"),
		)
	}
	return nil
}

// unresolvedConflictKeys describes every unresolved conflict of the error by its kind, its terminal and the actions the
// parser is left undecided between, sorted and without duplicates.
//
// Two things are deliberately not part of the description. The state is not, because the two automatons number their
// states differently. And the actions which competed for the terminal are not, because IELR(1) merges two isocores
// whenever they decide the conflict the same way, which unions the actions they contribute: the merged state is left
// undecided between the same actions as each isocore, while more actions competed for the terminal than in either of
// them. That merging is what IELR(1) is for, so what has to agree is which conflicts are left and what each of them is
// undecided between.
func unresolvedConflictKeys(err error) []string {
	var result []string
	for _, unresolvedConflictError := range conflict.UnresolvedConflictErrors(err) {
		for _, entry := range unresolvedConflictError.Report.Entries {
			result = append(result, fmt.Sprintf(
				"%s on terminal %s: %s",
				entry.Kind, entry.Terminal, strings.Join(entry.DecisionContributions, ", "),
			))
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}

// checkSizeInvariant verifies |LALR(1)| <= |IELR(1)| <= |canonical LR(1)|. Conflict resolution never adds or removes
// states, so comparing the resolved tables is valid. An IELR(1) table larger than canonical LR(1) or smaller than
// LALR(1) is a correctness-preserving quality bug - splitting too eagerly or losing a required split.
func (t comparisonTables) checkSizeInvariant() error {
	if t.hasLalrParser && len(t.sutParser.States) < len(t.lalrParser.States) {
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
