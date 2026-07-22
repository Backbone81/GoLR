package golr

import (
	"context"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// GrammarToParser calculates a parser from the context free grammar.
//
// The grammar is taken as a frontend produces it. What comes back are minimal LR(1) parser tables a backend can
// serialize, with every conflict of the grammar decided, so that no state is left with more than one action for a
// terminal. The tables accept the same language and produce the same parses as canonical LR(1) ones while staying close
// to the size of LALR(1) ones, which is what IELR(1) is for.
//
// The policy factory decides the conflicts, see conflict.PolicyFactory. Pass conflict.DefaultPolicy to decide them
// the way GNU Bison and Yacc do. The policy is not only what phase 5 resolves the conflicts with, it is also what
// phase 3 splits the states with, see GrammarToUnresolvedParser.
//
// Every conflict which was found is returned, whether it was decided or not, because a parser generator reports the
// conflicts of a grammar to the user even when it decided them on its own. The error reports the conflicts which were
// left undecided, one conflict.UnresolvedConflictError each; no parser can be generated from such a grammar, so the
// parser tables come back empty then and the conflicts are all there is left to report.
func GrammarToParser(
	grammar frontend.Grammar,
	policyFactory conflict.PolicyFactory,
) (backend.Parser, []conflict.Conflict, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: IELR1: GoLR: GrammarToParser").End()

	// Phases 0 to 4 of IELR(1) build the parser.
	parser := GrammarToUnresolvedParser(grammar, policyFactory)

	// Phase 5 of IELR(1) (section 3.7 of the paper).
	conflicts, err := conflict.Resolve(&parser, policyFactory(parser.Grammar))
	if err != nil {
		return backend.Parser{}, conflicts, err
	}

	// Resolving a conflict can delete the only shift into a state, which strands that state and everything behind it.
	// This is the unreachable state removal of section 3.8.2, the optional phase 6 of the paper.
	parser, conflicts = conflict.RemoveUnreachableStates(parser, conflicts)
	return parser, conflicts, nil
}

// GrammarToUnresolvedParser runs phases 0 to 4 of IELR(1) and stops there, so the minimal LR(1) parser tables come back
// with their conflicts intact and their unreachable states in place, before phase 5 decides anything.
//
// This is what the oracle and differential testing work is after: IELR(1) has a conflict exactly where canonical LR(1)
// has one, which is the invariant phase 3 is judged by, and a table whose conflicts are already resolved has none left
// to compare. It saves those callers from augmenting the grammar and driving the builder themselves, and it gives the
// three GoLR cores one shape to be called with. Reach for GrammarToParser whenever the tables are meant for a backend,
// because a table with conflicts in it is not a parser.
//
// Unlike the LALR(1) and canonical LR(1) cores, IELR(1) needs the policy to build at all: phase 3 decides the dominant
// contribution of definition 3.42 with it, and the compatibility test of definition 3.43 is defined in terms of that
// decision. So these tables are not policy free, they are only unresolved - which policy is passed in decides which
// states phase 3 declines to split.
func GrammarToUnresolvedParser(grammar frontend.Grammar, policyFactory conflict.PolicyFactory) backend.Parser {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Core: IELR1: GoLR: GrammarToUnresolvedParser").End()

	// The whole algorithm works on the augmented grammar, where a new start symbol derives the old one followed by the
	// end of input marker, so the caller hands us the grammar as the frontend produced it and we augment it here.
	augmentedGrammar := frontend.AugmentGrammar(grammar)

	builder := NewIELR1(augmentedGrammar, policyFactory(augmentedGrammar))
	return builder.BuildParser()
}

// IELR1 provides an implementation of the IELR(1) algorithm as described by Denny and Malloy in "The IELR(1) algorithm
// for generating minimal LR(1) parser tables for non-LR(1) grammars with conflict resolution" at
// https://doi.org/10.1016/j.scico.2009.08.001.
type IELR1 struct {
	grammar frontend.Grammar
	parser  backend.Parser

	// conflictPolicy is the conflict resolution which phase 3 uses to decide the dominant contribution of definition 3.42.
	// It is the Δ function of the paper, and phase 3's compatibility test of definition 3.43 is defined in terms of it:
	// two isocores are merged only when they agree on the dominant contribution the policy decides. This is what makes
	// the result IELR(1) rather than minimal LR(1) - phase 3 declines to split a state whose lookahead distinctions the
	// policy resolves away, keeping the tables close to LALR(1) size.
	conflictPolicy conflict.Policy

	// TODO: We should not store the LALR(1) builder. Instead we should copy over what we need and be done with it.
	lalr1Builder LALR1Builder

	// predecessorStateIdxsByStateIdx provides the state indexes for the states which are predecessors when accessed by
	// a state index. This is definition 3.15 of IELR(1) and simply named "predecessors" there.
	// TODO: This table might be valuable to calculate and use during LALR(1) construction already.
	predecessorStateIdxsByStateIdx [][]int

	// followKernelItemsByGotoIdx reports if the goto does depend on the kernel item's lookahead set. It holds the kernel
	// item indexes of the state the goto is coming from. This is definition 3.16 of IELR(1) and named
	// "follow_kernel_items" there.
	followKernelItemsByGotoIdx []utils.Bitset

	// inadequaciesByStateIdx holds the inadequacies of the LALR(1) parser tables, keyed by the state index of the
	// conflicted state. This is definition 3.27 of IELR(1) and named "inadequacy_lists" there.
	inadequaciesByStateIdx map[int][]*Inadequacy

	// annotationListsByStateIdx holds the annotations of a state, keyed by state index. This is definition 3.29 of
	// IELR(1) and named "annotation_lists" there.
	annotationListsByStateIdx map[int][]Annotation
}

// NewIELR1 returns a new IELR(1) builder for the augmented grammar. The split policy is the conflict resolution phase 3
// uses to decide the dominant contribution of definition 3.42, so that it merges isocores whose only difference is a
// conflict the policy resolves, which keeps the split automaton close to LALR(1) size. The caller passes it in rather
// than the builder constructing its own, because it has to be the same conflict resolution which phase 5 applies to the
// finished tables, or phase 3 would decline to split a state over a decision phase 5 then makes differently. Both come
// from the same conflict.PolicyFactory over the same augmented grammar, see GrammarToParser.
func NewIELR1(augmentedGrammar frontend.Grammar, conflictPolicy conflict.Policy) IELR1 {
	result := IELR1{
		grammar: augmentedGrammar,
		parser: backend.Parser{
			Grammar: augmentedGrammar,
		},
		conflictPolicy: conflictPolicy,
	}
	return result
}

func (i *IELR1) BuildParser() backend.Parser {
	defer trace.StartRegion(context.TODO(), "IELR(1): BuildParser").End()

	i.phase0ComputeLALR1ParserTables()
	i.phase1ComputeAuxiliaryTables()
	i.phase2ComputeAnnotations()
	i.phase3SplitStates()
	i.phase4ComputeReductionLookaheads()
	// Phase 5 of IELR(1) (section 3.7 of the paper), resolving the remaining conflicts, is deliberately not a step of the
	// builder. It runs outside, in GrammarToParser, through conflict.Resolve, so that BuildParser returns the minimal
	// LR(1) tables with their conflicts intact for oracle and differential testing.
	return i.parser
}

func (i *IELR1) Parser() backend.Parser {
	return i.parser
}

func (i *IELR1) phase0ComputeLALR1ParserTables() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 0: LALR(1)").End()
	i.lalr1Builder = NewLALR1Builder(i.grammar)
	i.lalr1Builder.Build()
	i.parser = i.lalr1Builder.Parser()
}

// phase1ComputeAuxiliaryTables computes the auxiliary tables of section 3.3 of IELR(1).
//
// Section 3.3 asks for three tables: predecessors, follow_kernel_items and always_follows. Only the first two are
// computed here. The always_follows of definition 3.20 are already computed by the LALR(1) builder, because we follow
// implementation 2 of section 3.3.5, which the paper recommends for a parser generator without an existing LALR(1)
// implementation: the always follows are computed between the two steps of phase 0 and phase 0 derives its goto follows
// from them with definition 3.24. That saves a closure computation and means successor follows are never computed at
// all.
func (i *IELR1) phase1ComputeAuxiliaryTables() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 1: Compute auxiliary tables").End()

	i.initPredecessorStateIdxsByStateIdx()
	i.initFollowKernelItems()
}

// initPredecessorStateIdxsByStateIdx initializes predecessorStateIdxsByStateIdx.
func (i *IELR1) initPredecessorStateIdxsByStateIdx() {
	i.predecessorStateIdxsByStateIdx = make([][]int, len(i.parser.States))
	for stateIdx := range i.parser.States {
		state := i.parser.States[stateIdx]
		for _, transition := range state.TransitionActions.All() {
			i.predecessorStateIdxsByStateIdx[transition.StateIdx()] = append(
				i.predecessorStateIdxsByStateIdx[transition.StateIdx()],
				stateIdx,
			)
		}
	}
}

// initFollowKernelItems initializes followKernelItemsByGotoIdx as specified in definition 3.16 of IELR(1).
//
// A goto follow set depends on the lookahead set of a kernel item of the state the goto is coming from, when the kernel
// item has its position in front of the nonterminal of that goto and the rest of the production after that nonterminal
// can be empty. Definition 3.16 does not stop at the goto itself, but follows the goto follows internal relation: the
// follow set of a goto contains the follow sets of all the gotos it depends on internally, so it also depends on the
// kernel items those gotos depend on.
//
// We therefore seed every goto with the kernel items the goto itself depends on, and let the digraph algorithm add the
// kernel items of the gotos which are reachable through the internal relation. This gives us the reflexive transitive
// closure of the internal relation which the definition asks for. Kernel item indexes are only meaningful within a
// single state, but the internal relation only ever relates gotos which come from the same state, so the propagation
// never mixes kernel item indexes of different states.
func (i *IELR1) initFollowKernelItems() {
	gotoRecords := i.lalr1Builder.lookaheads.GotoRecords()
	i.followKernelItemsByGotoIdx = make([]utils.Bitset, len(gotoRecords))
	for gotoIdx, gotoRecord := range gotoRecords {
		state := i.parser.States[gotoRecord.FromStateIdx]
		for kernelItemIdx, kernelItem := range state.KernelItems.All() {
			production := i.grammar.Productions[kernelItem.ProductionIdx()]
			if kernelItem.Position() == len(production.SymbolRefs) {
				// The kernel item is at the end of the production, so it does not take the goto.
				continue
			}
			symbolRef := production.SymbolRefs[kernelItem.Position()]
			if symbolRef.IsTerminal() || symbolRef.Idx() != gotoRecord.NonterminalIdx {
				// The kernel item does not move over the nonterminal of the goto, so it does not take the goto.
				continue
			}
			if !i.lalr1Builder.lookaheads.IsCoreTailEmpty(backend.NewCore(kernelItem.ProductionIdx(), kernelItem.Position()+1)) {
				// The rest of the production after the nonterminal transition can not be empty, so the lookahead set of
				// the kernel item can never follow the nonterminal of the goto.
				continue
			}
			i.followKernelItemsByGotoIdx[gotoIdx].Add(kernelItemIdx)
		}
	}
	propagation := NewDigraphAlgorithm(i.followKernelItemsByGotoIdx, i.lalr1Builder.lookaheads.GotoFollowsInternalRelation())
	propagation.Execute()
}

// FollowKernelItems returns the kernel items whose lookahead sets a goto follow set depends on, indexed by goto index.
// The kernel item indexes are indexes into the kernel items of the state the goto is coming from. This is definition
// 3.16 of IELR(1) and named "follow_kernel_items" there.
//
// The table is only valid after phase 1 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) FollowKernelItems() []utils.Bitset {
	return i.followKernelItemsByGotoIdx
}

// GotoRecords returns the details about every nonterminal transition of the LALR(1) automaton, indexed by goto index.
// This is what the goto indexes of FollowKernelItems refer to.
//
// The table is only valid after phase 0 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) GotoRecords() []GotoRecord {
	return i.lalr1Builder.lookaheads.GotoRecords()
}

// Predecessors returns the state indexes of the states which have a transition into the state, indexed by state index.
// This is definition 3.15 of IELR(1) and named "predecessors" there.
//
// The table is only valid after phase 1 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) Predecessors() [][]int {
	return i.predecessorStateIdxsByStateIdx
}

// GotoIdxsByStateIdx returns the goto indexes of the gotos which come from a state, keyed by state index.
//
// The table is only valid after phase 0 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) GotoIdxsByStateIdx() map[int][]int {
	return i.lalr1Builder.lookaheads.GotoIdxsByStateIdx()
}

// GotoFollows returns the goto follow set of every goto, indexed by goto index. This is "goto_follows" from definition
// 3.4 of IELR(1).
//
// The table is only valid after phase 0 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) GotoFollows() []backend.LookaheadSet {
	return i.lalr1Builder.lookaheads.GotoFollows()
}

// AlwaysFollows returns the terminals which follow a goto no matter what the lookahead sets of the kernel items of the
// state the goto comes from are, indexed by goto index. This is definition 3.20 of IELR(1) and named "always_follows"
// there.
//
// The table is only valid after phase 0 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) AlwaysFollows() []backend.LookaheadSet {
	return i.lalr1Builder.lookaheads.AlwaysFollows()
}

func (i *IELR1) phase2ComputeAnnotations() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 2: Compute annotations").End()

	annotationsBuilder := NewAnnotationsBuilder(
		i.parser,
		i.conflictPolicy,
		i.lalr1Builder.lookaheads.GotoRecords(),
		i.lalr1Builder.lookaheads.GotoIdxsByStateIdx(),
		i.lalr1Builder.lookaheads.GotoFollows(),
		i.lalr1Builder.lookaheads.AlwaysFollows(),
		i.predecessorStateIdxsByStateIdx,
		i.followKernelItemsByGotoIdx,
	)
	annotationsBuilder.Execute()
	i.inadequaciesByStateIdx = annotationsBuilder.Inadequacies()
	i.annotationListsByStateIdx = annotationsBuilder.AnnotationLists()
}

// Inadequacies returns the inadequacies of the LALR(1) parser tables, keyed by the state index of the conflicted
// state. This is definition 3.27 of IELR(1) and named "inadequacy_lists" there.
//
// The table is only valid after phase 2 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) Inadequacies() map[int][]*Inadequacy {
	return i.inadequaciesByStateIdx
}

// AnnotationLists returns the annotations of the LALR(1) states, keyed by state index. An annotation describes whether
// and how any isocore which phase 3 might split from the state can contribute to an inadequacy. This is definition 3.29
// of IELR(1) and named "annotation_lists" there.
//
// The table is only valid after phase 2 has run. It describes the LALR(1) automaton of phase 0 and is not updated for
// the split automaton of phases 3 and 4, so it must not be indexed with the states of the final parser.
func (i *IELR1) AnnotationLists() map[int][]Annotation {
	return i.annotationListsByStateIdx
}

// phase3SplitStates splits the LALR(1) states into the isocores of the minimal LR(1) parser tables, as described in
// section 3.5 of IELR(1). The split automaton replaces the LALR(1) automaton in the parser tables. Its reduce actions
// still carry the LALR(1) reduction lookahead sets, which phase 4 recomputes for the split automaton.
func (i *IELR1) phase3SplitStates() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 3: Split states").End()

	splitStatesBuilder := NewSplitStatesBuilder(
		i.grammar,
		i.parser.States,
		i.conflictPolicy,
		i.annotationListsByStateIdx,
		i.lalr1Builder.lookaheads.GotoRecords(),
		i.lalr1Builder.lookaheads.GotoIdxsByStateIdx(),
		i.lalr1Builder.lookaheads.AlwaysFollows(),
		i.followKernelItemsByGotoIdx,
	)
	splitStatesBuilder.Build()
	i.parser.States = splitStatesBuilder.States()
}

// phase4ComputeReductionLookaheads recomputes the reduction lookahead sets for the split automaton and writes them back
// into its reduce actions. This is section 3.6 of IELR(1), which runs step 2 of phase 0 again on the states phase 3
// produced. The reduction lookahead builder derives the reduction lookaheads from the states alone, so it serves both
// phase 0 and phase 4.
//
// The full recomputation is mandatory: the item lookahead sets phase 3 recomputed cannot be reused, because they are
// filtered down to the annotation-relevant terminals of definition 3.38 and may carry lookaheads from predecessors
// which were later redirected away.
//
// A state phase 3 left without any predecessor still gets its lookaheads recomputed, and the backward traces of
// reachable states may pass through such a state. Both are phase 3 orphans of section 3.8.1 of IELR(1), which the paper
// accepts as suboptimum state merging: correctness is unaffected, only table minimality can suffer.
func (i *IELR1) phase4ComputeReductionLookaheads() {
	defer trace.StartRegion(context.TODO(), "IELR(1): Phase 4: Compute reduction lookaheads").End()

	reductionLookaheadBuilder := NewReductionLookaheadBuilder(i.grammar, i.parser.States)
	reductionLookaheadBuilder.Build()
	applyReductionLookaheads(i.parser.States, reductionLookaheadBuilder.ReduceActions())
}
