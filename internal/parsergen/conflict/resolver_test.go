package conflict_test

import (
	"errors"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golr "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	lalr1golr "github.com/backbone81/golr/internal/parsergen/core/lalr1/golr"
	lr1golr "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
)

var _ = Describe("Resolve", func() {
	// Resolving the conflicts of the parser tables is phase 5 of IELR(1), and it is what makes parser tables with
	// conflicts usable for a parser. It works on the parser tables alone, so it does not matter which algorithm
	// computed them.
	It("should resolve every conflict of an ambiguous grammar with the policy of GNU Bison", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(conflictedTerminals(parser)).ToNot(
			BeEmpty(),
			"the test grammar is ambiguous, so the parser tables are expected to have conflicts to resolve",
		)

		conflicts, err := conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))

		// The conflicts are resolved in the parser tables we passed in. The policy is total, so it decides every
		// conflict, no conflict is left unresolved to error about, and no terminal is left with more than one action for
		// the parser to choose between.
		Expect(err).ToNot(HaveOccurred())
		Expect(conflicts).ToNot(BeEmpty())
		Expect(conflictedTerminals(parser)).To(BeEmpty())
	})

	// A terminal which does not associate at all leaves the state without any action for it, so the parser reports an
	// error when it sees the terminal there. This is the one case where resolving a conflict removes every action
	// instead of keeping the one which won.
	It("should remove every action for a conflicted terminal which does not associate", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())

		conflicts, err := conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))

		// A terminal which is rejected is a conflict the policy decided, so it is no reason to error.
		Expect(err).ToNot(HaveOccurred())

		var errorConflicts []conflict.Conflict
		for _, c := range conflicts {
			if c.Decision.Kind == conflict.DecisionError {
				errorConflicts = append(errorConflicts, c)
			}
		}
		Expect(errorConflicts).ToNot(
			BeEmpty(),
			"the nonassociative terminal of the test grammar is expected to be rejected somewhere",
		)

		for _, c := range errorConflicts {
			// The conflict names a terminal of the augmented grammar, so it is checked by name rather than against the
			// index constants of PrecedenceTestGrammar, which are the indexes before augmenting shifted them.
			Expect(parser.Grammar.Terminals[c.TerminalIdx].Name).To(Equal("<"))
			Expect(actionCount(parser.States[c.StateIdx], c.TerminalIdx)).To(
				BeZero(),
				"the state is expected to have no action left for the terminal which does not associate",
			)
		}
	})

	// A policy which decides nothing leaves every conflict unresolved. The parser tables are left with more than one
	// action for the conflicted terminals, which no parser can be generated from, so this is an error.
	It("should fail on the conflicts the policy leaves unresolved", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		wantConflictedTerminals := conflictedTerminals(parser)
		Expect(wantConflictedTerminals).ToNot(
			BeEmpty(),
			"the test grammar is ambiguous, so the parser tables are expected to have conflicts",
		)

		conflicts, err := conflict.Resolve(&parser, conflict.NullPolicy(parser.Grammar))

		// The error joins one error per unresolved conflict, so every conflict which was left undecided is reported.
		Expect(err).To(HaveOccurred())
		var unresolvedErrors interface{ Unwrap() []error }
		Expect(errors.As(err, &unresolvedErrors)).To(
			BeTrue(),
			"the error is expected to join the errors of the individual conflicts",
		)
		Expect(unresolvedErrors.Unwrap()).To(HaveLen(len(conflicts)))
		for _, unresolvedError := range unresolvedErrors.Unwrap() {
			var unresolvedConflictError conflict.UnresolvedConflictError
			Expect(errors.As(unresolvedError, &unresolvedConflictError)).To(BeTrue())
			Expect(unresolvedConflictError.Conflict.Decision.Kind).To(Equal(conflict.DecisionUnresolved))
		}

		Expect(conflicts).ToNot(BeEmpty())
		for _, c := range conflicts {
			Expect(c.Decision).To(
				Equal(conflict.NewUnresolvedDecision(c.Contributions)),
				"a policy which decides nothing is expected to leave the conflict with every contribution it "+
					"started with, but the decision was %s",
				c.Decision,
			)
		}
		Expect(conflictedTerminals(parser)).To(
			Equal(wantConflictedTerminals),
			"an unresolved conflict is expected to keep its actions in the parser tables",
		)
	})

	// A reduce/reduce conflict has no shift to fall back on, so the production which was declared first wins it.
	It("should resolve a reduce/reduce conflict in favor of the production which was declared first", func() {
		// The LALR(1) parser tables of this grammar have a reduce/reduce conflict which is an artifact of merging
		// states, see the canonical LR(1) tests. It gives us a reduce/reduce conflict to resolve.
		parser := lalr1golr.GrammarToUnresolvedParser(ielr1golr.ReduceReduceConflictTestGrammar, conflict.DefaultPolicy)

		conflicts, err := conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))

		Expect(err).ToNot(HaveOccurred())
		Expect(conflicts).ToNot(BeEmpty())
		for _, c := range conflicts {
			// The grammar declares no precedence at all, so every contribution of the conflict is a reduction and the
			// earliest production wins.
			wantProductionIdx := -1
			for _, contribution := range c.Contributions.All() {
				Expect(contribution.IsReduceAction()).To(BeTrue())
				if wantProductionIdx == -1 || contribution.ProductionIdx() < wantProductionIdx {
					wantProductionIdx = contribution.ProductionIdx()
				}
			}
			Expect(c.Decision).To(
				Equal(conflict.NewDominantDecision(conflict.NewReduceContribution(wantProductionIdx))),
				"expected the production %d to win the conflict, but the decision was %s",
				wantProductionIdx,
				c.Decision,
			)
		}
		Expect(conflictedTerminals(parser)).To(BeEmpty())
	})
})

var _ = Describe("Detect", func() {
	// Detect is what a caller of GrammarToUnresolvedParser uses to learn which conflicts its tables are left with. It
	// has to see exactly the conflicts Resolve sees, because Resolve is the only other thing which reports them, and it
	// has to leave the tables alone, because the caller still wants the conflicting actions.
	It("should report the conflicts Resolve reports, undecided and without touching the parser tables", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		wantConflictedTerminals := conflictedTerminals(parser)
		Expect(wantConflictedTerminals).ToNot(
			BeEmpty(),
			"the test grammar is ambiguous, so the parser tables are expected to have conflicts",
		)

		detectedConflicts := conflict.Detect(parser)

		Expect(detectedConflicts).ToNot(BeEmpty())
		for _, c := range detectedConflicts {
			Expect(c.Decision.Kind).To(
				Equal(conflict.DecisionUndefined),
				"no policy was applied, so the conflict is expected to carry no decision",
			)
		}
		Expect(conflictedTerminals(parser)).To(
			Equal(wantConflictedTerminals),
			"detecting the conflicts is expected to leave the parser tables alone",
		)

		resolvedConflicts, err := conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))
		Expect(err).ToNot(HaveOccurred())

		// The decision is the one thing which sets the two apart, so it is dropped before comparing what is left: the
		// same conflicted terminals, in the same order, with the same contributions competing for them.
		for i := range resolvedConflicts {
			resolvedConflicts[i].Decision = conflict.Decision{}
		}
		Expect(detectedConflicts).To(Equal(resolvedConflicts))
	})

	It("should report nothing for parser tables whose conflicts were resolved", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())

		_, err = conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))
		Expect(err).ToNot(HaveOccurred())

		Expect(conflict.Detect(parser)).To(BeEmpty())
	})

	It("should answer HasConflict for parser tables with and without conflicts", func() {
		parser, err := lr1golr.GrammarToUnresolvedParser(conflict.PrecedenceTestGrammar, conflict.DefaultPolicy)
		Expect(err).ToNot(HaveOccurred())
		Expect(conflict.HasConflict(parser)).To(BeTrue())

		_, err = conflict.Resolve(&parser, conflict.DefaultPolicy(parser.Grammar))
		Expect(err).ToNot(HaveOccurred())
		Expect(conflict.HasConflict(parser)).To(BeFalse())
	})
})

// conflictedTerminals returns the terminals the parser tables have more than one action for, keyed by state index.
func conflictedTerminals(parser backend.Parser) map[int][]int {
	result := make(map[int][]int)
	for stateIdx := range parser.States {
		for terminalIdx, contributions := range conflict.ContributionsByTerminalIdx(parser.States[stateIdx]) {
			if contributions.Length() <= 1 {
				continue
			}
			result[stateIdx] = append(result[stateIdx], terminalIdx)
		}
		slices.Sort(result[stateIdx])
	}
	return result
}

// actionCount returns the number of actions the state has on the terminal.
func actionCount(state backend.State, terminalIdx int) int {
	contributions := conflict.ContributionsByTerminalIdx(state)[terminalIdx]
	return contributions.Length()
}
