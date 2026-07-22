package oracle

import (
	"fmt"
	"math/rand"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// DefaultMaxExpansions is the default soft cap on the number of nonterminal expansions a single derivation performs
// before it switches to always choosing a shortest-terminating production. It keeps generated sentences small, which
// matters because every sentence is parsed by both interpreters in the differential test.
const DefaultMaxExpansions = 12

// InputGenerator produces random terminal sentences that a grammar derives, by a random leftmost derivation. Every
// sentence it returns is in the language of the grammar, so the differential test can feed it to both interpreters
// expecting an accept: an early reject from either side is then itself a signal.
//
// It runs on an *augmented* grammar — the shape frontend.AugmentGrammar produces and every builder in this package
// consumes — so the terminal indexes it returns are already in the augmented alphabet the resolved parser tables speak,
// and no shifting is needed to feed a generated sentence to a ParserInterpreter. Augmentation prepends the EOF terminal
// (index 0) and wraps the grammar in the single production `$accept -> Start EOF` at production index 0. The generator
// derives from Start — the first symbol of that production — rather than from $accept, so the trailing EOF is never
// emitted and generated sentences never contain the EOF terminal; the ParserInterpreter appends the one EOF itself. The
// constructor asserts this augmentation shape.
//
// Termination is guaranteed for any productive grammar. The generator precomputes, for every nonterminal, the minimum
// height of a derivation tree rooted at it and which of its productions achieve that minimum. While a derivation has
// expansions left in its budget it expands nonterminals by a uniformly chosen production, which lets recursion and
// longer right hand sides build up a non-trivial sentence; once the budget is spent it only ever chooses a
// shortest-terminating production, which drives every pending nonterminal to a terminal string in a bounded number of
// further expansions. Because the budget is a single counter shared across the whole derivation, the number of
// unbounded ("free") expansions is capped regardless of how the grammar branches, so the sentence length stays bounded.
type InputGenerator struct {
	// MaxExpansions is the soft cap on the number of nonterminal expansions before the derivation switches to only
	// choosing shortest-terminating productions. It is the main control over the length of the generated sentences.
	MaxExpansions int

	// grammar is the augmented grammar whose sentences are generated.
	grammar frontend.Grammar

	// derivationStartNonterminalIdx is the nonterminal derivations are rooted at: Start, the first symbol of the
	// `$accept -> Start EOF` production, not the $accept start symbol itself. Rooting at Start skips the production which
	// emits the prepended EOF terminal, so a generated sentence carries no EOF.
	derivationStartNonterminalIdx int

	// rand is the source of randomness. Seeding it deterministically makes a stream of generated sentences reproducible.
	rand *rand.Rand

	// productionIdxsByNonterminalIdx groups the productions by their left hand side nonterminal, so a nonterminal's
	// alternatives can be drawn without scanning all productions.
	productionIdxsByNonterminalIdx [][]int

	// shortestProductionIdxsByNonterminalIdx holds, for every nonterminal, the productions whose derivation height
	// equals the nonterminal's minimum derivation height. Choosing among these is what makes a derivation terminate:
	// each such production only depends on nonterminals which are themselves closer to a terminal string.
	shortestProductionIdxsByNonterminalIdx [][]int
}

// NewInputGenerator returns an input generator for the given augmented grammar (see frontend.AugmentGrammar). The
// grammar must be productive (every nonterminal derives some terminal string), which every grammar from GrammarGenerator
// and every hand-written corpus grammar is; a non-productive grammar is a programming error and is caught by a debug
// assertion. The generator precomputes the auxiliary tables the derivation needs, so constructing it once per grammar
// and calling Generate repeatedly is cheaper than reconstructing it per sentence.
func NewInputGenerator(grammar frontend.Grammar, rng *rand.Rand) *InputGenerator {
	utils.DebugAssert(func() error {
		return assertAugmented(grammar)
	})
	generator := &InputGenerator{
		MaxExpansions: DefaultMaxExpansions,
		grammar:       grammar,
		// Start is the first symbol of `$accept -> Start EOF`; rooting derivations here never emits the prepended EOF.
		derivationStartNonterminalIdx: grammar.Productions[acceptProductionIdx].SymbolRefs[0].Idx(),
		rand:                          rng,
	}
	generator.computeDerivationTables()
	return generator
}

// assertAugmented checks that the grammar has the shape frontend.AugmentGrammar produces, on which the generator relies
// to root its derivations at Start while skipping the EOF terminal: the production at acceptProductionIdx must be
// `$accept -> Start EOF`, that is a nonterminal followed by the EOF terminal.
func assertAugmented(grammar frontend.Grammar) error {
	if len(grammar.Productions) <= acceptProductionIdx {
		return fmt.Errorf("augmented grammar is missing the accept production at index %d", acceptProductionIdx)
	}
	acceptProduction := grammar.Productions[acceptProductionIdx]
	if len(acceptProduction.SymbolRefs) != 2 {
		return fmt.Errorf("accept production must have exactly 2 symbols, got %d", len(acceptProduction.SymbolRefs))
	}
	if !acceptProduction.SymbolRefs[0].IsNonterminal() {
		return fmt.Errorf("first symbol of the accept production must be the start nonterminal")
	}
	if !acceptProduction.SymbolRefs[1].IsTerminal() || acceptProduction.SymbolRefs[1].Idx() != eofTerminalIdx {
		return fmt.Errorf("second symbol of the accept production must be the EOF terminal %d", eofTerminalIdx)
	}
	return nil
}

// computeDerivationTables groups the productions by their left hand side and computes, for every nonterminal, the
// productions which achieve its minimum derivation height. The minimum heights are found by a fixpoint iteration: a
// production's height is one plus the largest minimum height among the nonterminals on its right hand side (terminals
// contribute nothing, so an empty or terminals-only right hand side has height one), and a nonterminal's minimum height
// is the smallest height among its productions. The fixpoint is reached in at most as many passes as there are
// nonterminals, because each pass makes at least one more nonterminal reachable in a bounded number of steps.
func (g *InputGenerator) computeDerivationTables() {
	nonterminalCount := len(g.grammar.Nonterminals)

	g.productionIdxsByNonterminalIdx = make([][]int, nonterminalCount)
	for productionIdx, production := range g.grammar.Productions {
		g.productionIdxsByNonterminalIdx[production.NonterminalIdx] = append(
			g.productionIdxsByNonterminalIdx[production.NonterminalIdx],
			productionIdx,
		)
	}

	// known reports whether a finite minimum height has been found for a nonterminal yet, and minHeight holds that
	// height once it has. Keeping a separate known flag avoids representing infinity with a sentinel that could overflow
	// when one is added to it.
	known := make([]bool, nonterminalCount)
	minHeight := make([]int, nonterminalCount)
	for changed := true; changed; {
		changed = false
		for _, production := range g.grammar.Productions {
			height, ok := g.productionHeight(production, known, minHeight)
			if !ok {
				continue
			}
			nonterminalIdx := production.NonterminalIdx
			if !known[nonterminalIdx] || height < minHeight[nonterminalIdx] {
				known[nonterminalIdx] = true
				minHeight[nonterminalIdx] = height
				changed = true
			}
		}
	}

	// A productive grammar makes every nonterminal reachable in the fixpoint above; if one is not, the derivation below
	// would have no shortest production to fall back on and could not terminate.
	utils.DebugAssert(func() error {
		for nonterminalIdx := range nonterminalCount {
			if !known[nonterminalIdx] {
				return fmt.Errorf("nonterminal %d is not productive", nonterminalIdx)
			}
		}
		return nil
	})

	g.shortestProductionIdxsByNonterminalIdx = make([][]int, nonterminalCount)
	for nonterminalIdx := range nonterminalCount {
		for _, productionIdx := range g.productionIdxsByNonterminalIdx[nonterminalIdx] {
			height, _ := g.productionHeight(g.grammar.Productions[productionIdx], known, minHeight)
			if height == minHeight[nonterminalIdx] {
				g.shortestProductionIdxsByNonterminalIdx[nonterminalIdx] = append(
					g.shortestProductionIdxsByNonterminalIdx[nonterminalIdx],
					productionIdx,
				)
			}
		}
	}
}

// productionHeight returns the derivation height of the production given the minimum heights known so far, and whether
// that height is defined yet. The height is one plus the largest minimum height among the right hand side nonterminals,
// and it is undefined as long as any of those nonterminals has no known minimum height.
func (g *InputGenerator) productionHeight(production frontend.Production, known []bool, minHeight []int) (int, bool) {
	maxChildHeight := 0
	for _, symbolRef := range production.SymbolRefs {
		if symbolRef.IsTerminal() {
			continue
		}
		if !known[symbolRef.Idx()] {
			return 0, false
		}
		maxChildHeight = max(maxChildHeight, minHeight[symbolRef.Idx()])
	}
	return maxChildHeight + 1, true
}

// Generate produces one random sentence the grammar derives, as a slice of terminal indexes into the augmented grammar
// (EOF excluded — see the type doc). Consecutive calls with the same generator draw a stream of different sentences; two
// generators with identically seeded randomness produce the same stream. The result may be empty when Start derives the
// empty string, which is a valid sentence the caller feeds as empty input.
func (g *InputGenerator) Generate() []int {
	var sentence []int
	remainingExpansions := g.MaxExpansions
	g.derive(g.derivationStartNonterminalIdx, &remainingExpansions, &sentence)
	return sentence
}

// derive expands the nonterminal in place, appending the terminals of the chosen production to the sentence and
// recursing into its nonterminals left to right, which realizes a leftmost derivation. The shared remaining-expansions
// counter is decremented once per expansion; once it is spent the choice is restricted to shortest-terminating
// productions, so the whole derivation is driven to completion in a bounded number of further expansions.
func (g *InputGenerator) derive(nonterminalIdx int, remainingExpansions *int, sentence *[]int) {
	productionIdx := g.pickProduction(nonterminalIdx, *remainingExpansions)
	*remainingExpansions--

	for _, symbolRef := range g.grammar.Productions[productionIdx].SymbolRefs {
		if symbolRef.IsTerminal() {
			*sentence = append(*sentence, symbolRef.Idx())
		} else {
			g.derive(symbolRef.Idx(), remainingExpansions, sentence)
		}
	}
}

// pickProduction chooses which production of the nonterminal to expand. While the derivation still has expansions left
// it draws uniformly from all of the nonterminal's productions, so recursion and longer right hand sides can build up a
// non-trivial sentence. Once the budget is spent it draws only from the shortest-terminating productions, which is what
// guarantees the derivation terminates.
func (g *InputGenerator) pickProduction(nonterminalIdx int, remainingExpansions int) int {
	if remainingExpansions <= 0 {
		shortest := g.shortestProductionIdxsByNonterminalIdx[nonterminalIdx]
		return shortest[g.rand.Intn(len(shortest))]
	}
	productions := g.productionIdxsByNonterminalIdx[nonterminalIdx]
	return productions[g.rand.Intn(len(productions))]
}
