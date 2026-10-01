package counterexample

import (
	"context"
	"fmt"
	"runtime/trace"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/utils"
)

// conflictItemPair is a pair of conflict items: a reduce item, and an item which shifts the conflict terminal or a
// second reduce item.
type conflictItemPair struct {
	reduceItemIdx int
	otherItemIdx  int
}

// Find returns the counterexamples of the conflicts conflict.Resolve returned for the parser tables, one slice per
// conflict in the same order. The tables are those a core returned, with the conflicts resolved or left unresolved, and
// the policy is the one the conflicts were resolved with, bound to the grammar of the parser.
//
// A conflict gets one counterexample per pair of the items of its actions the declarations of the grammar left
// competing: every reduce item with every item which shifts the conflict terminal, and every two reduce items, the
// earlier production first. This is the unit of the paper, which counts two shift items in one state as two conflicts
// (section 5.1, figure 7).
func Find(parser backend.Parser, conflicts []conflict.Conflict, policy conflict.Policy) [][]Counterexample {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Counterexample: Find").End()

	if len(conflicts) == 0 {
		return nil
	}
	tables := NewLookupTables(parser, policy)
	result := make([][]Counterexample, len(conflicts))
	for conflictIdx, c := range conflicts {
		for _, pair := range tables.conflictItemPairs(c) {
			builder := newNonunifyingBuilder(&tables, pair.reduceItemIdx, pair.otherItemIdx, c.TerminalIdx)
			result[conflictIdx] = append(result[conflictIdx], builder.Build())
		}
	}
	return result
}

// conflictItemPairs returns the pairs of conflict items of the conflict, see Find.
func (t *LookupTables) conflictItemPairs(c conflict.Conflict) []conflictItemPair {
	var shiftItemIdxs []int
	var reduceItemIdxs []int
	for _, contribution := range c.Undeclared.All() {
		if contribution.IsShiftAction() {
			shiftItemIdxs = t.ItemsWithNextSymbol(c.StateIdx, frontend.NewTerminalRef(c.TerminalIdx))
			t.assertShiftKept(c)
			continue
		}
		productionIdx := contribution.ProductionIdx()
		reduceCore := backend.NewCore(productionIdx, len(t.grammar.Productions[productionIdx].SymbolRefs))
		reduceItemIdx, found := t.ItemIdx(c.StateIdx, reduceCore)
		utils.DebugAssert(func() error {
			if !found {
				return fmt.Errorf("state %d has no reduce item of production %d", c.StateIdx, productionIdx)
			}
			return nil
		})
		reduceItemIdxs = append(reduceItemIdxs, reduceItemIdx)
	}

	var result []conflictItemPair
	for i, reduceItemIdx := range reduceItemIdxs {
		for _, shiftItemIdx := range shiftItemIdxs {
			result = append(result, conflictItemPair{reduceItemIdx: reduceItemIdx, otherItemIdx: shiftItemIdx})
		}
		for _, otherReduceItemIdx := range reduceItemIdxs[i+1:] {
			result = append(result, conflictItemPair{reduceItemIdx: reduceItemIdx, otherItemIdx: otherReduceItemIdx})
		}
	}
	return result
}

// assertShiftKept checks that the tables still hold the transition of a shift the declarations left competing. The
// search walks the transitions of the resolved tables, which only works as long as a rule of last resort never removes
// a shift.
func (t *LookupTables) assertShiftKept(c conflict.Conflict) {
	utils.DebugAssert(func() error {
		if _, found := t.transitionStateIdx(c.StateIdx, frontend.NewTerminalRef(c.TerminalIdx)); !found {
			return fmt.Errorf("state %d lost its shift on terminal %d to a rule of last resort", c.StateIdx, c.TerminalIdx)
		}
		return nil
	})
}
