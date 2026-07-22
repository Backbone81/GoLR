package oracle

import (
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strconv"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// ProductionScenario selects the shape of the right hand side generated for a single production. The different
// scenarios are biased toward the grammar structures which exercise the LALR(1) reduction lookahead computation and
// the isocore merging of the oracle, so that a corpus of random grammars is not dominated by trivial grammars which
// never stress the interesting code paths.
type ProductionScenario int

const (
	// ProductionScenarioRandom generates a right hand side of random length where every symbol is independently a
	// terminal or a nonterminal.
	ProductionScenarioRandom ProductionScenario = iota

	// ProductionScenarioEmpty generates an empty right hand side, which makes the left hand side nonterminal nullable.
	// Nullable nonterminals drive the goto follows successor relations and the always follows of the DeRemer-Pennello
	// algorithm.
	ProductionScenarioEmpty

	// ProductionScenarioTerminals generates a right hand side of terminals only, which grounds the nonterminal in a
	// leaf-like production the way real grammars have many of.
	ProductionScenarioTerminals

	// ProductionScenarioNonterminals generates a right hand side of nonterminals only, which builds the chains along
	// which goto follows propagate and creates the internal and predecessor includes relations.
	ProductionScenarioNonterminals

	// ProductionScenarioRecursive generates a random right hand side and forces the left hand side nonterminal to
	// appear in it, which creates cycles and reuses the nonterminal in more than one context.
	ProductionScenarioRecursive

	// ProductionScenarioSharedNonterminal deliberately constructs a nonterminal reached from two distinct reachable
	// contexts with different following terminals. This is the situation where canonical LR(1) and LALR(1) diverge:
	// LALR(1) merges the isocore of the shared nonterminal and takes the union of the two lookahead sets. Unlike the
	// other scenarios it builds a fixed cluster of fresh nonterminals, so it reports that it does not fit when the
	// nonterminal budget has no room left.
	ProductionScenarioSharedNonterminal

	// ProductionScenarioNullableSuffix builds a production of the form B -> alpha A gamma where gamma is a guaranteed
	// nullable, non-empty suffix. Because gamma derives the empty string, whatever can follow B flows onto the
	// transition on A: this is exactly the includes relation of the DeRemer-Pennello algorithm, along which a lookahead
	// set is propagated backwards across a nullable gap. The other scenarios only create includes edges by chance, when a
	// random suffix happens to be nullable, so this scenario makes the relation which is most responsible for the
	// LALR(1) reduction lookahead computation a reliable part of the corpus.
	ProductionScenarioNullableSuffix

	// ProductionScenarioSharedNonterminalNullableGap is the ProductionScenarioSharedNonterminal situation with a nullable
	// nonterminal placed between the shared nonterminal and the terminal which distinguishes the two contexts. In the
	// plain shared scenario that terminal sits directly behind the shared nonterminal, so its lookahead contribution is a
	// direct read; separating them by a nullable symbol forces the contribution to reach the shared nonterminal across
	// the gap through the reads and includes relations instead. This exercises the nullable propagation of the lookahead
	// computation at the very place where canonical LR(1) and LALR(1) diverge, which the adjacent variant never reaches.
	ProductionScenarioSharedNonterminalNullableGap

	// ProductionScenarioReduceReduce constructs the mysterious reduce-reduce conflict: two nonterminals with a common
	// core reduction are reached from two contexts whose following terminals are swapped, so that LALR(1) merges the
	// isocore and unions the lookahead sets into a reduce-reduce conflict which canonical LR(1) does not have. Where
	// ProductionScenarioSharedNonterminal produces a shift-reduce flavoured divergence from a single shared nonterminal,
	// this produces the distinct reduce-reduce inadequacy shape from two competing reductions in one state.
	ProductionScenarioReduceReduce
)

// ProductionScenario implements fmt.Stringer.
var _ fmt.Stringer = ProductionScenario(0)

// String returns the name of the scenario, or a numeric fallback for an unknown value.
func (s ProductionScenario) String() string {
	switch s {
	case ProductionScenarioRandom:
		return "Random"
	case ProductionScenarioEmpty:
		return "Empty"
	case ProductionScenarioTerminals:
		return "Terminals"
	case ProductionScenarioNonterminals:
		return "Nonterminals"
	case ProductionScenarioRecursive:
		return "Recursive"
	case ProductionScenarioSharedNonterminal:
		return "SharedNonterminal"
	case ProductionScenarioNullableSuffix:
		return "NullableSuffix"
	case ProductionScenarioSharedNonterminalNullableGap:
		return "SharedNonterminalNullableGap"
	case ProductionScenarioReduceReduce:
		return "ReduceReduce"
	default:
		return "ProductionScenario(" + strconv.Itoa(int(s)) + ")"
	}
}

// GrammarGenerator produces random context-free grammars for testing purposes. The generated grammars are un-augmented
// frontend.Grammar values which pass frontend.Grammar.Validate, so the caller augments them the same way it augments
// any other grammar before handing them to a builder.
//
// A grammar is grown from a work list seeded with the start nonterminal. When a nonterminal is introduced it is given a
// grounding production straight away, whose right hand side is terminals only, so it is productive from the start. The
// nonterminal is then put on the work list once for each of its remaining productions; pulling it adds one more
// production, whose right hand side is chosen by a weighted scenario and may introduce new terminals and nonterminals
// up to the configured maxima. Introducing nonterminals only while building reachable productions keeps every
// nonterminal reachable, and capping the maxima gives the grammar a finite size. The result is a valid, reachable and
// productive grammar.
type GrammarGenerator struct {
	// MaxTerminalCount is the upper bound on the number of distinct terminals in the grammar.
	MaxTerminalCount int

	// MaxNonterminalCount is the upper bound on the number of distinct nonterminals in the grammar, including the start
	// nonterminal.
	MaxNonterminalCount int

	// MaxProductionCountPerNonterminal is the upper bound on the number of productions a single nonterminal may have.
	// The actual number is drawn uniformly from one to this maximum when the nonterminal is introduced. More than one
	// production per nonterminal is what creates the alternatives which conflicts and non-trivial lookahead sets come
	// from.
	MaxProductionCountPerNonterminal int

	// MaxRHSSymbolCount caps the number of symbols on the right hand side of a production.
	MaxRHSSymbolCount int

	// NewTerminalProbability is the chance to introduce a new terminal, rather than reuse an existing one, whenever a
	// terminal is needed and the maximum has not been reached yet. Value in the range [0.0, 1.0].
	NewTerminalProbability float64

	// NewNonterminalProbability is the chance to introduce a new nonterminal, rather than reuse an existing one,
	// whenever a nonterminal is needed and the maximum has not been reached yet. Value in the range [0.0, 1.0].
	NewNonterminalProbability float64

	// ScenarioWeights gives the relative weight of every scenario. A scenario with weight zero, or one missing from the
	// map, is never chosen.
	ScenarioWeights map[ProductionScenario]int

	// Rand is the source of randomness. It must not be nil. Seeding it deterministically makes a generated grammar
	// reproducible from its seed.
	Rand *rand.Rand

	// ObservedScenarioCounts records how many productions each scenario has successfully built. It accumulates across
	// Generate calls and is never reset by the generator, so after generating a corpus it holds the scenario
	// distribution of the whole corpus; a caller which wants per-grammar or per-run figures clears it itself. It is an
	// observability output, not a configuration input: the scenario weights only set how often a scenario is rolled,
	// while a cluster scenario is dropped for the rest of a grammar when it does not fit the remaining nonterminal
	// budget, so the actual firing distribution can differ markedly from the weights. Watching these counts is how a
	// scenario being starved by the budget - which would silently erode the discriminating coverage of the corpus - is
	// caught.
	ObservedScenarioCounts map[ProductionScenario]int

	// grammar is the grammar under construction. Terminals, nonterminals and productions are appended to it as they are
	// generated, so their counts are the lengths of the respective slices and the finished grammar is simply this field
	// once the work list is drained.
	grammar frontend.Grammar

	// nonterminalIdxWorklist holds the nonterminals which still need a production. A nonterminal appears once for each
	// production it is to have.
	nonterminalIdxWorklist utils.DynamicRingBuffer[int]

	// remainingScenarioWeights is the working copy of ScenarioWeights for a single Generate call. When a scenario
	// reports that it does not fit into the remaining budget it is deleted from this copy, so it is never rolled again
	// for the rest of the grammar. This is correct because the budget only ever shrinks during a Generate call, so a
	// scenario which does not fit now can never fit later either.
	remainingScenarioWeights map[ProductionScenario]int
}

// DefaultGrammarGenerator returns a generator with small, general-purpose limits and scenario weights biased toward the
// grammar structures which exercise the LALR(1) machinery. The grammars stay small on purpose: the cost of the
// differential test is the canonical LR(1) construction of the oracle, which can blow up in the number of states.
//
// The three heavy discriminating clusters carry roughly equal weight so the corpus reaches the shift-reduce merge, the
// reduce-reduce conflict and the shared-nonterminal-across-a-nullable-gap shapes about equally often. Their weights are
// not exactly equal because a cluster only fits while the nonterminal budget still has room for it, and once it does not
// fit it is dropped for the rest of the grammar: the gap variant needs four fresh nonterminals against the three of the
// others, so it only fits on an early roll and is under-represented at equal weight. Its weight is raised to bring the
// three shapes back to parity in ObservedScenarioCounts (measured over a corpus of two thousand grammars).
func DefaultGrammarGenerator(rng *rand.Rand) *GrammarGenerator {
	return &GrammarGenerator{
		MaxTerminalCount:                 5,
		MaxNonterminalCount:              6,
		MaxProductionCountPerNonterminal: 6,
		MaxRHSSymbolCount:                4,
		NewTerminalProbability:           0.5,
		NewNonterminalProbability:        0.5,
		ScenarioWeights: map[ProductionScenario]int{
			ProductionScenarioRandom:                       40,
			ProductionScenarioEmpty:                        10,
			ProductionScenarioTerminals:                    15,
			ProductionScenarioNonterminals:                 20,
			ProductionScenarioRecursive:                    15,
			ProductionScenarioSharedNonterminal:            20,
			ProductionScenarioNullableSuffix:               15,
			ProductionScenarioSharedNonterminalNullableGap: 26,
			ProductionScenarioReduceReduce:                 20,
		},
		Rand: rng,
	}
}

// Generate produces a random grammar according to the configuration of the generator. Consecutive calls with the same
// Rand produce a stream of different grammars, while two generators with identically seeded Rand produce the same
// stream. The working state is reset on every call, so a generator can be reused for more than one grammar.
func (g *GrammarGenerator) Generate() frontend.Grammar {
	// Clamp the limits which would otherwise make the random draws below ill-defined. This only ever raises an invalid
	// value to its minimum, so a valid configuration is left untouched.
	g.MaxNonterminalCount = max(g.MaxNonterminalCount, 1)
	g.MaxTerminalCount = max(g.MaxTerminalCount, 1)
	g.MaxProductionCountPerNonterminal = max(g.MaxProductionCountPerNonterminal, 1)
	g.MaxRHSSymbolCount = max(g.MaxRHSSymbolCount, 1)

	// Reset the working state, so a previous call does not leak into this one. The zero value of frontend.Grammar has
	// no symbols and no productions, and a start nonterminal index of zero, which is exactly what the start nonterminal
	// created below ends up with.
	g.grammar = frontend.Grammar{}
	g.nonterminalIdxWorklist = utils.NewDynamicRingBuffer[int]()
	g.remainingScenarioWeights = maps.Clone(g.ScenarioWeights)

	// The observed scenario counts accumulate across Generate calls, so they are only created once and never reset here.
	// Everything downstream may therefore assume the map exists.
	if g.ObservedScenarioCounts == nil {
		g.ObservedScenarioCounts = map[ProductionScenario]int{}
	}

	// Create the start nonterminal and seed the work list with it. It keeps index 0, so it stays the start nonterminal
	// of the generated grammar.
	g.newNonterminal()

	g.run()

	utils.DebugAssert(g.grammar.Validate)
	return g.grammar
}

// run drains the work list, adding one production per pulled nonterminal. Every nonterminal already carries its
// grounding production from newNonterminal, so the productions generated here are free to reference any symbol without
// putting productivity at risk.
func (g *GrammarGenerator) run() {
	for !g.nonterminalIdxWorklist.IsEmpty() {
		nonterminalIdx := g.nonterminalIdxWorklist.Remove()

		scenario := g.pickScenario()
		symbolRefs, success := g.buildRHS(nonterminalIdx, scenario)
		if !success {
			// The scenario could not build a right hand side within the remaining budget. Drop it from the working
			// weights so it is not rolled again, and put the nonterminal back to roll a different scenario. This
			// terminates because there are finitely many scenarios to drop and the plain scenarios always succeed.
			delete(g.remainingScenarioWeights, scenario)
			g.nonterminalIdxWorklist.Add(nonterminalIdx)
			continue
		}
		g.ObservedScenarioCounts[scenario]++
		g.grammar.Productions = append(g.grammar.Productions, frontend.Production{
			NonterminalIdx: nonterminalIdx,
			SymbolRefs:     symbolRefs,
		})
	}
}

// newNonterminal introduces a new nonterminal and returns its index. The number of productions it is to have is drawn
// uniformly from one to the configured maximum. Its first production is created right away as a grounding production,
// so the nonterminal is productive from the moment it exists; the remaining productions are put on the work list to be
// filled in by a scenario later.
func (g *GrammarGenerator) newNonterminal() int {
	// The grounding production has a right hand side of terminals only, which derives a terminal string on its own.
	// This makes the nonterminal productive without depending on any other nonterminal, so no grammar-wide repair is
	// needed.
	nonterminalIdx := g.addNonterminal(g.buildGroundingRHS())

	// One of the productions is the grounding production above, so the work list only needs the remaining ones.
	for range g.Rand.Intn(g.MaxProductionCountPerNonterminal) {
		g.nonterminalIdxWorklist.Add(nonterminalIdx)
	}
	return nonterminalIdx
}

// addNonterminal appends a new nonterminal with the given single production and returns its index. In contrast to
// newNonterminal it does not put the nonterminal on the work list, so the nonterminal ends up with exactly the one
// production passed in. It is used both to bootstrap a nonterminal with its grounding production and by scenarios which
// construct a self-contained cluster of fully specified nonterminals.
func (g *GrammarGenerator) addNonterminal(symbolRefs []frontend.SymbolRef) int {
	nonterminalIdx := len(g.grammar.Nonterminals)
	g.grammar.Nonterminals = append(g.grammar.Nonterminals, frontend.Symbol{
		Name: "N" + strconv.Itoa(nonterminalIdx),
	})
	g.grammar.Productions = append(g.grammar.Productions, frontend.Production{
		NonterminalIdx: nonterminalIdx,
		SymbolRefs:     symbolRefs,
	})
	return nonterminalIdx
}

// buildGroundingRHS builds the right hand side of a grounding production: a possibly empty sequence of terminals. An
// empty right hand side derives the empty string and a non-empty one derives its terminals, so either way the
// nonterminal it belongs to can derive a terminal string.
func (g *GrammarGenerator) buildGroundingRHS() []frontend.SymbolRef {
	length := g.Rand.Intn(g.MaxRHSSymbolCount + 1)
	symbolRefs := make([]frontend.SymbolRef, 0, length)
	for range length {
		symbolRefs = append(symbolRefs, frontend.NewTerminalRef(g.pickTerminal()))
	}
	return symbolRefs
}

// pickNonterminal returns the index of a nonterminal for a right hand side, introducing a new one when the maximum
// allows it and the random draw calls for it, and reusing an existing one otherwise. There is always at least the start
// nonterminal to reuse.
func (g *GrammarGenerator) pickNonterminal() int {
	if len(g.grammar.Nonterminals) < g.MaxNonterminalCount && g.Rand.Float64() < g.NewNonterminalProbability {
		return g.newNonterminal()
	}
	return g.Rand.Intn(len(g.grammar.Nonterminals))
}

// pickTerminal returns the index of a terminal for a right hand side, introducing a new one when the maximum allows it
// and the random draw calls for it, and reusing an existing one otherwise. The first terminal is always introduced,
// because there is nothing to reuse yet.
func (g *GrammarGenerator) pickTerminal() int {
	if len(g.grammar.Terminals) == 0 ||
		(len(g.grammar.Terminals) < g.MaxTerminalCount && g.Rand.Float64() < g.NewTerminalProbability) {
		terminalIdx := len(g.grammar.Terminals)
		g.grammar.Terminals = append(g.grammar.Terminals, frontend.Symbol{
			Name: "t" + strconv.Itoa(terminalIdx),
		})
		return terminalIdx
	}
	return g.Rand.Intn(len(g.grammar.Terminals))
}

// buildRHS builds the right hand side for a production of the nonterminal according to the scenario. The second return
// value reports whether the scenario succeeded; a scenario which cannot fit its structure into the remaining budget
// returns false and leaves the grammar untouched, so the caller can roll a different scenario.
func (g *GrammarGenerator) buildRHS(nonterminalIdx int, scenario ProductionScenario) ([]frontend.SymbolRef, bool) {
	switch scenario {
	case ProductionScenarioEmpty:
		return nil, true
	case ProductionScenarioTerminals:
		symbolRefs := make([]frontend.SymbolRef, 1+g.Rand.Intn(g.MaxRHSSymbolCount))
		for i := range symbolRefs {
			symbolRefs[i] = frontend.NewTerminalRef(g.pickTerminal())
		}
		return symbolRefs, true
	case ProductionScenarioNonterminals:
		symbolRefs := make([]frontend.SymbolRef, 1+g.Rand.Intn(g.MaxRHSSymbolCount))
		for i := range symbolRefs {
			symbolRefs[i] = frontend.NewNonterminalRef(g.pickNonterminal())
		}
		return symbolRefs, true
	case ProductionScenarioRecursive:
		symbolRefs := g.buildRandomRHS()
		// Force the left hand side to appear somewhere on the right hand side, which turns the production recursive.
		position := g.Rand.Intn(len(symbolRefs) + 1)
		return slices.Insert(symbolRefs, position, frontend.NewNonterminalRef(nonterminalIdx)), true
	case ProductionScenarioSharedNonterminal:
		return g.buildSharedNonterminalRHS()
	case ProductionScenarioNullableSuffix:
		return g.buildNullableSuffixRHS()
	case ProductionScenarioSharedNonterminalNullableGap:
		return g.buildSharedNonterminalNullableGapRHS()
	case ProductionScenarioReduceReduce:
		return g.buildReduceReduceRHS()
	case ProductionScenarioRandom:
		fallthrough
	default:
		return g.buildRandomRHS(), true
	}
}

// buildSharedNonterminalRHS builds the shared-nonterminal situation deliberately: a fresh nonterminal reached from two
// fresh call sites whose terminals surrounding it differ. Canonical LR(1) keeps the two contexts apart, while LALR(1)
// merges the isocore of the shared nonterminal and takes the union of the two lookahead sets, which is exactly where a
// wrong lookahead computation would show up. The two call sites become the right hand side of the pulled nonterminal,
// which keeps the whole cluster reachable.
//
// It needs three fresh nonterminals and reports failure when the nonterminal budget has no room for them.
func (g *GrammarGenerator) buildSharedNonterminalRHS() ([]frontend.SymbolRef, bool) {
	const neededNonterminalCount = 3
	if len(g.grammar.Nonterminals)+neededNonterminalCount > g.MaxNonterminalCount {
		return nil, false
	}

	// The shared nonterminal grounds in a single terminal, so it is productive, and its reduction is what the two
	// contexts disagree on.
	sharedIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
	})

	// Each call site surrounds the shared nonterminal with terminals. The terminal following the shared nonterminal is
	// what each context contributes to its lookahead set.
	firstCallSiteIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
		frontend.NewNonterminalRef(sharedIdx),
		frontend.NewTerminalRef(g.pickTerminal()),
	})
	secondCallSiteIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
		frontend.NewNonterminalRef(sharedIdx),
		frontend.NewTerminalRef(g.pickTerminal()),
	})

	return []frontend.SymbolRef{
		frontend.NewNonterminalRef(firstCallSiteIdx),
		frontend.NewNonterminalRef(secondCallSiteIdx),
	}, true
}

// buildNullableSuffixRHS builds a production of the form B -> alpha A gamma where gamma is a nullable, non-empty suffix.
// Because gamma derives the empty string, whatever can follow B flows onto the transition on A, which is the includes
// relation of the DeRemer-Pennello algorithm. The suffix is a fresh nonterminal whose only production is empty, so it is
// guaranteed nullable and productive. The nonterminal A before it is any existing or fresh nonterminal, so it carries
// reductions whose lookahead the includes edge propagates. The prefix alpha is a random right hand side which may be
// empty, so the edge is exercised both with and without symbols in front of A.
//
// It needs one fresh nonterminal for the nullable suffix and reports failure when the nonterminal budget has no room for
// it.
func (g *GrammarGenerator) buildNullableSuffixRHS() ([]frontend.SymbolRef, bool) {
	const neededNonterminalCount = 1
	if len(g.grammar.Nonterminals)+neededNonterminalCount > g.MaxNonterminalCount {
		return nil, false
	}

	// The suffix grounds in a single empty production, so it derives the empty string and nothing else, which makes it
	// unconditionally nullable.
	nullableSuffixIdx := g.addNonterminal(nil)

	// The prefix may be empty, so the production ranges over both B -> A gamma and B -> alpha A gamma. The nonterminal A
	// is placed directly before the nullable suffix, so the includes edge runs from its transition to the follow of the
	// left hand side.
	symbolRefs := g.buildRandomRHS()
	symbolRefs = append(symbolRefs, frontend.NewNonterminalRef(g.pickNonterminal()))
	symbolRefs = append(symbolRefs, frontend.NewNonterminalRef(nullableSuffixIdx))
	return symbolRefs, true
}

// buildSharedNonterminalNullableGapRHS builds the ProductionScenarioSharedNonterminal cluster with a nullable nonterminal
// inserted between the shared nonterminal and the distinguishing terminal of each call site. In the plain shared
// scenario the distinguishing terminal sits directly behind the shared nonterminal, so its lookahead contribution is a
// direct read; the nullable gap forces the contribution to reach the shared nonterminal across the gap through the reads
// and includes relations, which is where a wrong nullable propagation of the lookahead computation would show up. A
// single nullable gap nonterminal is shared by both call sites; being empty-only it just bridges the gap and does not
// merge the two contexts of the shared nonterminal, which stay apart through their differing following terminals.
//
// It needs four fresh nonterminals (the shared nonterminal, the nullable gap and the two call sites) and reports failure
// when the nonterminal budget has no room for them.
func (g *GrammarGenerator) buildSharedNonterminalNullableGapRHS() ([]frontend.SymbolRef, bool) {
	const neededNonterminalCount = 4
	if len(g.grammar.Nonterminals)+neededNonterminalCount > g.MaxNonterminalCount {
		return nil, false
	}

	// The shared nonterminal grounds in a single terminal, so it is productive, and its reduction is what the two
	// contexts disagree on.
	sharedIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
	})

	// The gap grounds in a single empty production, so it is unconditionally nullable and lets the distinguishing
	// terminal of each call site reach the shared nonterminal across it instead of as a direct read.
	gapIdx := g.addNonterminal(nil)

	// Each call site surrounds the shared nonterminal with terminals, but with the nullable gap between the shared
	// nonterminal and the terminal which the context contributes to the lookahead set.
	firstCallSiteIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
		frontend.NewNonterminalRef(sharedIdx),
		frontend.NewNonterminalRef(gapIdx),
		frontend.NewTerminalRef(g.pickTerminal()),
	})
	secondCallSiteIdx := g.addNonterminal([]frontend.SymbolRef{
		frontend.NewTerminalRef(g.pickTerminal()),
		frontend.NewNonterminalRef(sharedIdx),
		frontend.NewNonterminalRef(gapIdx),
		frontend.NewTerminalRef(g.pickTerminal()),
	})

	return []frontend.SymbolRef{
		frontend.NewNonterminalRef(firstCallSiteIdx),
		frontend.NewNonterminalRef(secondCallSiteIdx),
	}, true
}

// buildReduceReduceRHS builds the mysterious reduce-reduce conflict. Two reducer nonterminals derive the same single
// terminal, so their completed items share a core and land in the same LALR(1) state. A driver nonterminal reaches the
// reducers from two prefixes and follows them with two terminals in swapped combinations, so that both reducers become
// reducible on both following terminals once the isocore is merged and the lookahead sets are unioned. This is a
// reduce-reduce conflict which canonical LR(1) does not have, because it keeps the two prefixes apart and each reducer
// stays reducible on a single terminal.
//
// The prefixes and following terminals are drawn from the shared terminal pool, so they are only distinct with some
// probability; when they collide the two contexts collapse and the conflict is no longer LALR(1) specific, which the
// corpus selection then discards as non-discriminating. It needs three fresh nonterminals (two reducers and the driver)
// and reports failure when the nonterminal budget has no room for them.
func (g *GrammarGenerator) buildReduceReduceRHS() ([]frontend.SymbolRef, bool) {
	const neededNonterminalCount = 3
	if len(g.grammar.Nonterminals)+neededNonterminalCount > g.MaxNonterminalCount {
		return nil, false
	}

	// The two reducers derive the same single terminal, so their completed items share a core and are reduced in the
	// same state, which is what turns the merged lookahead sets into a reduce-reduce conflict.
	reduced := frontend.NewTerminalRef(g.pickTerminal())
	firstReducerIdx := g.addNonterminal([]frontend.SymbolRef{reduced})
	secondReducerIdx := g.addNonterminal([]frontend.SymbolRef{reduced})

	// The two contexts reach the reducers behind different prefixes and are followed by swapped terminals: the driver
	// has the four productions prefix1 R1 follow1, prefix1 R2 follow2, prefix2 R1 follow2 and prefix2 R2 follow1. Under
	// LALR(1) each reducer's lookahead set becomes the union of the two following terminals.
	firstPrefix := frontend.NewTerminalRef(g.pickTerminal())
	secondPrefix := frontend.NewTerminalRef(g.pickTerminal())
	firstFollow := frontend.NewTerminalRef(g.pickTerminal())
	secondFollow := frontend.NewTerminalRef(g.pickTerminal())

	driverIdx := g.addNonterminal([]frontend.SymbolRef{firstPrefix, frontend.NewNonterminalRef(firstReducerIdx), firstFollow})
	g.grammar.Productions = append(g.grammar.Productions,
		frontend.Production{
			NonterminalIdx: driverIdx,
			SymbolRefs:     []frontend.SymbolRef{firstPrefix, frontend.NewNonterminalRef(secondReducerIdx), secondFollow},
		},
		frontend.Production{
			NonterminalIdx: driverIdx,
			SymbolRefs:     []frontend.SymbolRef{secondPrefix, frontend.NewNonterminalRef(firstReducerIdx), secondFollow},
		},
		frontend.Production{
			NonterminalIdx: driverIdx,
			SymbolRefs:     []frontend.SymbolRef{secondPrefix, frontend.NewNonterminalRef(secondReducerIdx), firstFollow},
		},
	)

	return []frontend.SymbolRef{frontend.NewNonterminalRef(driverIdx)}, true
}

// buildRandomRHS builds a right hand side of random length where every symbol is independently a terminal or a
// nonterminal. The length may be zero, which produces an empty right hand side just like ProductionScenarioEmpty does.
func (g *GrammarGenerator) buildRandomRHS() []frontend.SymbolRef {
	length := g.Rand.Intn(g.MaxRHSSymbolCount + 1)
	symbolRefs := make([]frontend.SymbolRef, 0, length)
	for range length {
		if g.Rand.Intn(2) == 0 {
			symbolRefs = append(symbolRefs, frontend.NewTerminalRef(g.pickTerminal()))
		} else {
			symbolRefs = append(symbolRefs, frontend.NewNonterminalRef(g.pickNonterminal()))
		}
	}
	return symbolRefs
}

// pickScenario chooses a scenario according to the configured weights. When no weight is positive it falls back to a
// completely random right hand side, so the generator always makes progress.
func (g *GrammarGenerator) pickScenario() ProductionScenario {
	totalWeight := 0
	for _, weight := range g.remainingScenarioWeights {
		if weight > 0 {
			totalWeight += weight
		}
	}
	if totalWeight == 0 {
		return ProductionScenarioRandom
	}

	choice := g.Rand.Intn(totalWeight)
	for _, scenario := range slices.Sorted(maps.Keys(g.remainingScenarioWeights)) {
		weight := g.remainingScenarioWeights[scenario]
		if weight <= 0 {
			continue
		}
		if choice < weight {
			return scenario
		}
		choice -= weight
	}
	return ProductionScenarioRandom
}
