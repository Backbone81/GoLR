package golr_test

import (
	"fmt"
	"math/rand"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/core/ielr1/golr/oracle"
	lr1golrcore "github.com/backbone81/golr/internal/parsergen/core/lr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

var _ = Describe("IELR(1) phase 2: compute annotations", func() {
	// Section 3.4 of IELR(1) walks the grammar of figure 5 through phase 2 in full detail, state by state and
	// contribution by contribution. That walk-through is the gold standard for phase 2, so we replay it here.
	//
	// The grammar has a single inadequacy: the state whose kernel item is "A -> a C D . E" has a shift/reduce conflict
	// on "a" between the shift of "E -> a" and the reduction of the empty production "E -> ".
	Context("the grammar of figure 5", func() {
		const (
			// The terminal "a" of the grammar. The augmented grammar puts the end of input marker in front of the
			// terminals of the grammar.
			terminalIdxA = 1

			// The productions of the augmented grammar. The augmented grammar puts the start production in front of the
			// productions of the grammar, so the productions of the grammar are moved back by one.
			productionIdxStart  = 0 // $accept -> S $end
			productionIdxSaABa  = 1 // S -> a A B a
			productionIdxSbABb  = 2 // S -> b A B b
			productionIdxAaCDE  = 3 // A -> a C D E
			productionIdxEmptyE = 9 // E ->

			// The states of our LALR(1) parser tables which the paper numbers 18, 17, 16, 2 and 5 in table 4. Our
			// LALR(1) construction creates the states in a different order than the paper depicts them in, so the state
			// indexes do not agree with the paper. The kernel items of every one of these states are pinned down in the
			// BeforeEach below, so a change to the state numbering fails on the state it went wrong on rather than
			// somewhere deep inside an expectation.
			stateIdxStart             = 0  // $accept -> . S $end
			stateIdxAlways            = 1  // S -> a . A B a, state 2 of the paper
			stateIdxNever             = 2  // S -> b . A B b, state 5 of the paper
			stateIdxSecondPredecessor = 4  // A -> a . C D E, state 16 of the paper
			stateIdxFirstPredecessor  = 9  // A -> a C . D E, state 17 of the paper
			stateIdxConflicted        = 14 // A -> a C D . E, state 18 of the paper
		)

		// kernelItemsByStateIdx are the kernel items the states of the walk-through are expected to have.
		kernelItemsByStateIdx := map[int]backend.CoreSet{
			stateIdxStart:             backend.NewCoreSet(backend.NewCore(productionIdxStart, 0)),
			stateIdxAlways:            backend.NewCoreSet(backend.NewCore(productionIdxSaABa, 1)),
			stateIdxNever:             backend.NewCoreSet(backend.NewCore(productionIdxSbABb, 1)),
			stateIdxSecondPredecessor: backend.NewCoreSet(backend.NewCore(productionIdxAaCDE, 1)),
			stateIdxFirstPredecessor:  backend.NewCoreSet(backend.NewCore(productionIdxAaCDE, 2)),
			stateIdxConflicted:        backend.NewCoreSet(backend.NewCore(productionIdxAaCDE, 3)),
		}

		var annotationsBuilder *ielr1golrcore.AnnotationsBuilder

		BeforeEach(func() {
			// The paper walks the grammar of figure 5 through phase 2 with a conflict-preserving policy: it presents the
			// annotations as computed, and only section 3.4.3 points out that a shift-over-reduce policy would make them
			// split-stable and discard them. The null policy resolves nothing, so it keeps every annotation the
			// annotation computation produces, which is what this walk-through checks.
			annotationsBuilder = newAnnotationsBuilder(ielr1golrcore.GotoFollowsTestGrammarFig5, conflict.NullPolicy)

			// The walk-through of the paper only makes sense when our LALR(1) tables have the states it talks about at
			// the state indexes the expectations below use.
			parser := annotationsBuilder.Parser()
			for stateIdx, wantKernelItems := range kernelItemsByStateIdx {
				gotKernelItems := parser.States[stateIdx].KernelItems
				Expect(gotKernelItems.Equal(&wantKernelItems)).To(
					BeTrue(),
					"state %d has the kernel items %s instead of %s",
					stateIdx, gotKernelItems.String(), wantKernelItems.String(),
				)
			}
		})

		// This is definition 3.26 of IELR(1). The lookahead set of the single kernel item of the conflicted state has to
		// contain the conflicted terminal, otherwise the conflict could not manifest there at all. The paper spells this
		// out for the conflicted state and for its predecessor.
		It("should compute the item lookahead sets", func() {
			conflictedLookaheadSet := annotationsBuilder.ItemLookaheadSet(stateIdxConflicted, 0)
			Expect(conflictedLookaheadSet.Contains(terminalIdxA)).To(BeTrue())

			predecessorLookaheadSet := annotationsBuilder.ItemLookaheadSet(stateIdxFirstPredecessor, 0)
			Expect(predecessorLookaheadSet.Contains(terminalIdxA)).To(BeTrue())

			// The kernel item of the start state is followed by the end of input marker, so nothing can ever follow it
			// and its lookahead set stays empty. This is point 3 of the definition.
			startLookaheadSet := annotationsBuilder.ItemLookaheadSet(stateIdxStart, 0)
			Expect(startLookaheadSet.IsEmpty()).To(BeTrue())
		})

		// This is definition 3.27 of IELR(1).
		It("should compute the inadequacies", func() {
			Expect(annotationsBuilder.Inadequacies()).To(HaveLen(1))

			inadequacies := annotationsBuilder.Inadequacies()[stateIdxConflicted]
			Expect(inadequacies).To(HaveLen(1))
			Expect(inadequacies[0].TerminalIdx).To(Equal(terminalIdxA))

			// The conflict has two contributions: the shift of "E -> a" and the reduction of the empty production
			// "E -> ". The paper calls them contribution 1 and contribution 2.
			var contributions []conflict.Contribution
			for _, contribution := range inadequacies[0].Contributions.All() {
				contributions = append(contributions, contribution)
			}
			Expect(contributions).To(Equal([]conflict.Contribution{
				conflict.NewShiftContribution(),
				conflict.NewReduceContribution(productionIdxEmptyE),
			}))
		})

		// This is definition 3.30 and definition 3.31 of IELR(1), which section 3.4.2 of the paper walks through for the
		// conflicted state: the shift is an always contribution, so its contribution row stays undefined, and the
		// reduction of the empty production is a potential contribution which the single kernel item of the state
		// carries.
		//
		// Definition 3.32 then carries that annotation backwards to the two predecessors on the lane. The paper points
		// out that the contribution matrix does not change on the way, because the kernel item of each of those states
		// feeds the kernel item of its successor and sees the conflicted terminal.
		DescribeTable("should annotate the states on the lane of the conflict",
			func(stateIdx int) {
				annotations := annotationsBuilder.AnnotationLists()[stateIdx]
				Expect(annotations).To(HaveLen(1))
				Expect(annotations[0].Inadequacy.StateIdx).To(Equal(stateIdxConflicted))
				Expect(annotations[0].Inadequacy.TerminalIdx).To(Equal(terminalIdxA))

				Expect(annotations[0].ContributionMatrix).To(HaveLen(2))
				Expect(annotations[0].ContributionMatrix[0].IsAlways()).To(BeTrue())
				Expect(annotations[0].ContributionMatrix[1].IsPotential()).To(BeTrue())
				Expect(slices.Collect(annotations[0].ContributionMatrix[1].KernelItems.All())).To(Equal([]int{0}))
			},
			// A -> a C D . E
			Entry("the conflicted state itself", stateIdxConflicted),

			// A -> a C . D E
			Entry("the predecessor of the conflicted state", stateIdxFirstPredecessor),

			// A -> a . C D E
			Entry("the predecessor of that predecessor", stateIdxSecondPredecessor),
		)

		// This is observation 3.33 and observation 3.34 of IELR(1), which section 3.4.3 of the paper walks through.
		//
		// The two remaining predecessors on the lane are where phase 2 stops. In the state "S -> a . A B a" the
		// conflicted terminal is an always follow of the goto on A, so the reduction becomes an always contribution. In
		// the state "S -> b . A B b" the rest of the kernel item after A can not be empty, so the goto on A does not
		// depend on the kernel item at all and the reduction becomes a never contribution. In both states all
		// contributions are always or never contributions, which means every isocore which can be split from them makes
		// exactly the same contributions. Splitting them can therefore not change which contribution dominates the
		// conflict, so both annotations are useless and phase 2 discards them.
		//
		// Discarding them also terminates the iteration along the lane, so the start state, which is the predecessor of
		// both of them, never gets annotated either.
		It("should discard the useless annotations and stop iterating along the lane", func() {
			Expect(annotationsBuilder.AnnotationLists()).ToNot(HaveKey(stateIdxAlways))
			Expect(annotationsBuilder.AnnotationLists()).ToNot(HaveKey(stateIdxNever))
			Expect(annotationsBuilder.AnnotationLists()).ToNot(HaveKey(stateIdxStart))

			// The annotated states are exactly the three states of the lane which the paper annotates.
			Expect(annotationsBuilder.AnnotationLists()).To(HaveLen(3))
		})
	})

	// This is the general case of definition 3.35, which section 3.4.3 of the paper spells out for the grammar of figure
	// 5 on page 24: if the shift/reduce conflict on "a" is resolved by shift over reduce, then the shift dominates the
	// conflict in every isocore which phase 3 could split from the conflicted state, because the shift is an always
	// contribution and no reduction can beat it. The dominant contribution is therefore split-stable, so the annotation
	// on the conflicted state is useless, as are all the annotations phase 2 computes from it along the lane. The paper
	// concludes that "phase 3 would have no reason to split any LALR(1) states", so phase 2 annotates nothing.
	It("should discard every annotation of figure 5 when shift over reduce resolves the conflict", func() {
		grammar := ielr1golrcore.GotoFollowsTestGrammarFig5

		annotationsBuilder := newAnnotationsBuilder(grammar, conflict.DefaultPolicy)

		// The conflict is still found, it is only its annotation which the split stability makes useless.
		Expect(annotationsBuilder.Inadequacies()).ToNot(BeEmpty())
		Expect(annotationsBuilder.AnnotationLists()).To(BeEmpty())
	})

	// The lanes of a conflicted state can be cyclic, so the reverse iteration of definition 3.29 only terminates because
	// it stops as soon as it computes an annotation which the state carries already. A left recursive grammar is what
	// puts a cycle into the lanes, so this makes sure we do not iterate forever on one.
	DescribeTable("should terminate the iteration along cyclic lanes",
		func(grammar frontend.Grammar) {
			// A conflict-preserving policy keeps every annotation which is not split-stable on its own, so that the
			// termination of the reverse iteration is what this test exercises rather than the policy discarding
			// annotations.
			conflictPolicy := conflict.NullPolicy(frontend.AugmentGrammar(grammar))
			annotationsBuilder := newAnnotationsBuilder(grammar, conflict.NullPolicy)

			Expect(annotationsBuilder.Inadequacies()).ToNot(BeEmpty())
			for _, annotations := range annotationsBuilder.AnnotationLists() {
				for _, annotation := range annotations {
					// Useless annotations are discarded, so no annotation which survives may be split-stable.
					Expect(annotation.IsSplitStable(conflictPolicy)).To(BeFalse())
				}
			}
		},
		Entry("the ambiguous grammar of figure 2", ielr1golrcore.AmbiguousTestGrammarFig2),
		Entry("a grammar with a reduce/reduce conflict", ielr1golrcore.ReduceReduceConflictTestGrammar),
	)

	// This is the property which makes phase 2 worth anything: when a conflict of the LALR(1) parser tables is not a
	// conflict of the canonical LR(1) parser tables, then it is an LR(1)-relative inadequacy, which phase 3 can only
	// remove by splitting a state. Phase 2 has to hand phase 3 an annotation for it, otherwise phase 3 has nothing to
	// work with and IELR(1) silently degrades into LALR(1).
	//
	// Failing to annotate is invisible in an end to end test which only checks that the parser accepts the right
	// language, because LALR(1) parser tables accept the same language. So we check it here, with canonical LR(1) as the
	// oracle.
	//
	// The oracle is asked about every inadequacy on its own, rather than about the grammar as a whole. A random grammar
	// regularly has a genuine conflict somewhere, and a genuine conflict is one which no state splitting can remove: any
	// isocore which phase 3 could split the conflicted state into still has it, so phase 2 is free to leave it
	// unannotated. Discarding a grammar because it has one would throw away all its other conflicts as well, and those
	// are perfectly good LR(1)-relative inadequacies which we do want to check.
	//
	// A conflict is genuine exactly when canonical LR(1) has it too. Canonical LR(1) keeps apart the contexts which
	// LALR(1) merges, so the isocores of a conflicted LALR(1) state are the canonical LR(1) states with the same kernel
	// items, and they are the finest split phase 3 could ever produce. When one of them has a conflict on the conflicted
	// terminal, splitting cannot help.
	It("should annotate every inadequacy which canonical LR(1) does not have", func() {
		const grammarCount = 2000

		// The corpus is seeded from the Ginkgo random seed, so every run explores a different set of grammars and a rare
		// edge case surfaces eventually rather than being masked by a fixed corpus, while a failing run stays
		// reproducible with `ginkgo --seed=...`. A master RNG derives a distinct seed for every grammar, so the corpus is
		// grammarCount different grammars instead of the same grammar repeated. The derived seed is reported with a
		// failure so a single failing grammar can be reconstructed directly, which is what shrinking it into a regression
		// fixture needs.
		masterRng := rand.New(rand.NewSource(GinkgoRandomSeed()))

		var discriminatingGrammarCount, lr1RelativeInadequacyCount, genuineConflictCount int
		for range grammarCount {
			grammarSeed := masterRng.Int63()
			grammar := oracle.DefaultGrammarGenerator(rand.New(rand.NewSource(grammarSeed))).Generate()

			// This oracle needs the raw canonical LR(1) table with its conflicts intact — the property tested below only
			// holds under a conflict-preserving policy — so it builds it directly instead of through GrammarToParser,
			// which resolves conflicts as a public parser interface should.
			lr1Builder := lr1golrcore.NewLR1Builder(frontend.AugmentGrammar(grammar))
			if err := lr1Builder.Build(); err != nil {
				// A grammar whose canonical LR(1) automaton exceeds the addressable state limit cannot be handled by the
				// oracle. It is skipped, not a failure of phase 2.
				Expect(err).To(
					MatchError(lr1golrcore.ErrStateLimitExceeded),
					"grammar seed %d:\n%s", grammarSeed, grammar.String(),
				)
				continue
			}
			lr1Parser := lr1Builder.Parser()
			lr1StateIdxsByKernelItemsHash := stateIdxsByKernelItemsHash(lr1Parser)

			// The property that every LR(1)-relative inadequacy is annotated only holds under a policy which resolves no
			// conflict: a policy which resolves a conflict split-stably would rightly discard its annotation, which is the
			// separate concern of the general case of definition 3.35. So this test uses the conflict-preserving policy.
			annotationsBuilder := newAnnotationsBuilder(grammar, conflict.NullPolicy)

			// An annotation is attached to every state along the lane of the conflict, not only to the conflicted state
			// itself, and a state may carry annotations of conflicts it is not the conflicted state of. So we collect the
			// inadequacies which are annotated anywhere, rather than looking at the annotations of the conflicted state.
			annotatedInadequacies := make(map[*ielr1golrcore.Inadequacy]bool)
			for _, annotations := range annotationsBuilder.AnnotationLists() {
				for _, annotation := range annotations {
					annotatedInadequacies[annotation.Inadequacy] = true
				}
			}

			var isDiscriminating bool
			for _, inadequacies := range annotationsBuilder.Inadequacies() {
				for _, inadequacy := range inadequacies {
					if isGenuineConflict(
						lr1Parser,
						lr1StateIdxsByKernelItemsHash,
						annotationsBuilder.Parser(),
						inadequacy,
					) {
						genuineConflictCount++
						continue
					}
					lr1RelativeInadequacyCount++
					isDiscriminating = true

					Expect(annotatedInadequacies).To(
						HaveKey(inadequacy),
						"the inadequacy on terminal %d of state %d is not annotated, but no canonical LR(1) isocore of "+
							"that state has a conflict on that terminal, for grammar seed %d:\n%s",
						inadequacy.TerminalIdx,
						inadequacy.StateIdx,
						grammarSeed,
						grammar.String(),
					)
				}
			}
			if isDiscriminating {
				discriminatingGrammarCount++
			}
		}

		// A run which never hits an LR(1)-relative inadequacy proves nothing at all, so we track how many of the
		// generated grammars actually discriminate.
		Expect(discriminatingGrammarCount).To(
			BeNumerically(">", 0),
			"none of the generated grammars was discriminating, so the property was never exercised",
		)
		AddReportEntry("discriminating grammars", fmt.Sprintf("%d of %d", discriminatingGrammarCount, grammarCount))
		AddReportEntry("LR(1)-relative inadequacies", lr1RelativeInadequacyCount)
		AddReportEntry("genuine conflicts skipped", genuineConflictCount)
	})
})

// stateIdxsByKernelItemsHash groups the states of the parser tables by the hash of their kernel items. Canonical LR(1)
// keeps the contexts of a state apart which LALR(1) merges, so a single LALR(1) state generally corresponds to several
// canonical LR(1) states with the same kernel items. Those are its isocores, and they are the finest split which phase 3
// could ever produce from that LALR(1) state.
//
// The hash is only what buckets the states. Two different sets of kernel items can end up in the same bucket, so the
// lookup below still compares the kernel items themselves.
func stateIdxsByKernelItemsHash(parser backend.Parser) map[uint64][]int {
	result := make(map[uint64][]int, len(parser.States))
	for stateIdx := range parser.States {
		kernelItemsHash := parser.States[stateIdx].KernelItems.Hash()
		result[kernelItemsHash] = append(result[kernelItemsHash], stateIdx)
	}
	return result
}

// isGenuineConflict reports if the conflict the inadequacy describes is a conflict of the grammar rather than an
// artifact of the LALR(1) state merging. That is the case when any canonical LR(1) isocore of the conflicted state has
// more than one action on the conflicted terminal, because then no state splitting can remove the conflict and phase 2
// is free to leave the inadequacy unannotated.
func isGenuineConflict(
	lr1Parser backend.Parser,
	lr1StateIdxsByKernelItemsHash map[uint64][]int,
	lalr1Parser backend.Parser,
	inadequacy *ielr1golrcore.Inadequacy,
) bool {
	kernelItems := lalr1Parser.States[inadequacy.StateIdx].KernelItems

	var isocoreStateIdxs []int
	for _, lr1StateIdx := range lr1StateIdxsByKernelItemsHash[kernelItems.Hash()] {
		lr1KernelItems := lr1Parser.States[lr1StateIdx].KernelItems
		if !lr1KernelItems.Equal(&kernelItems) {
			// The kernel items only share the hash bucket, so this state is not an isocore of the conflicted state.
			continue
		}
		isocoreStateIdxs = append(isocoreStateIdxs, lr1StateIdx)
	}
	Expect(isocoreStateIdxs).ToNot(
		BeEmpty(),
		"the conflicted LALR(1) state %d has no isocore in the canonical LR(1) parser tables",
		inadequacy.StateIdx,
	)

	for _, isocoreStateIdx := range isocoreStateIdxs {
		if actionCountOnTerminal(lr1Parser.States[isocoreStateIdx], inadequacy.TerminalIdx) > 1 {
			return true
		}
	}
	return false
}

// actionCountOnTerminal returns the number of actions the state has on the terminal. More than one action is a conflict.
func actionCountOnTerminal(state backend.State, terminalIdx int) int {
	var actionCount int
	for _, transitionAction := range state.TransitionActions.All() {
		if transitionAction.SymbolRef().IsTerminal() && transitionAction.SymbolRef().Idx() == terminalIdx {
			actionCount++
		}
	}
	for _, reduceAction := range state.ReduceActions.All() {
		if reduceAction.LookaheadSet.Contains(terminalIdx) {
			actionCount++
		}
	}
	return actionCount
}

// newAnnotationsBuilder runs the phases which phase 2 depends on and returns the annotations builder after it has
// computed the annotations. The conflict policy is what phase 2 decides split-stable dominant contributions with, so a
// caller passes the conflict-preserving empty compound policy to keep every annotation, or a resolving policy to have
// the split-stable ones discarded.
//
// The LALR(1) states come from the LALR(1) builder rather than from the IELR(1) builder, because phase 3 of IELR(1)
// splits those states and replaces them in the IELR(1) parser tables. The phase 0 and phase 1 auxiliary tables the
// annotations builder needs on top of them are relative to the LALR(1) automaton and stay valid, so those are taken from
// the IELR(1) builder. Both builders construct the same LALR(1) automaton from the same grammar, so their state indexes
// agree.
func newAnnotationsBuilder(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) *ielr1golrcore.AnnotationsBuilder {
	augmentedGrammar := frontend.AugmentGrammar(grammar)
	conflictPolicy := policyFactory(augmentedGrammar)

	lalr1Builder := ielr1golrcore.NewLALR1Builder(augmentedGrammar)
	lalr1Builder.Build()

	ielr1 := ielr1golrcore.NewIELR1(augmentedGrammar, conflictPolicy)
	ielr1.BuildParser()

	annotationsBuilder := ielr1golrcore.NewAnnotationsBuilder(
		lalr1Builder.Parser(),
		conflictPolicy,
		ielr1.GotoRecords(),
		ielr1.GotoIdxsByStateIdx(),
		ielr1.GotoFollows(),
		ielr1.AlwaysFollows(),
		ielr1.Predecessors(),
		ielr1.FollowKernelItems(),
	)
	annotationsBuilder.Execute()
	return annotationsBuilder
}
