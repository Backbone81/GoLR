package backend

import (
	"cmp"
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// ReduceAction is a reduce action of an LR(1) item consisting of a lookahead set of terminals and a production index
// to reduce for. The values for the production must be in the range of [0, 65535].
type ReduceAction struct {
	LookaheadSet  LookaheadSet `json:"lookaheadSet"  yaml:"lookaheadSet"`
	ProductionIdx int          `json:"productionIdx" yaml:"productionIdx"`
}

const (
	reduceActionProductionIdxMax = (1 << 16) - 1
)

// NewReduceAction creates a new reduce action with the given lookahead set and the production index.
func NewReduceAction(lookaheadSet LookaheadSet, productionIdx int) ReduceAction {
	utils.AssertValidIndex(productionIdx, reduceActionProductionIdxMax)

	return ReduceAction{
		LookaheadSet:  lookaheadSet,
		ProductionIdx: productionIdx,
	}
}

func (a ReduceAction) Equal(other ReduceAction) bool {
	return a.ProductionIdx == other.ProductionIdx && a.LookaheadSet.Equal(other.LookaheadSet)
}

func CompareReduceAction(x, y ReduceAction) int {
	if result := cmp.Compare(x.ProductionIdx, y.ProductionIdx); result != 0 {
		return result
	}
	return x.LookaheadSet.Compare(y.LookaheadSet)
}

func ReduceActionEqual(x, y ReduceAction) bool {
	return CompareReduceAction(x, y) == 0
}

// ReduceAction implements fmt.Stringer.
var _ fmt.Stringer = (*ReduceAction)(nil)

// String returns a string representation.
func (a ReduceAction) String() string {
	return fmt.Sprintf("(%s, production %d)", a.LookaheadSet.String(), a.ProductionIdx)
}
