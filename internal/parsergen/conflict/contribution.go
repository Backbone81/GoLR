package conflict

import (
	"errors"
	"fmt"

	"github.com/backbone81/golr/internal/utils"
)

// Contribution is a single action which a state can take on a terminal. When a state has more than one contribution for
// the same terminal, those contributions are in conflict with each other. This is the value of the contributions
// function of definition 2.17 of IELR(1).
//
// A contribution is either the shift of the terminal, or the reduction of a production. It is implemented as a single
// unsigned integer to allow for a compact representation and to give a stable order when sorting a slice of
// contributions.
type Contribution uint32

// NewShiftContribution creates the contribution which shifts the conflicted terminal.
func NewShiftContribution() Contribution {
	return Contribution(0)
}

// NewReduceContribution creates the contribution which reduces the production on the conflicted terminal.
func NewReduceContribution(productionIdx int) Contribution {
	utils.DebugAssert(func() error {
		if productionIdx < 0 || contributionMaxProductionIdx < productionIdx {
			return errors.New("production index out of bounds")
		}
		return nil
	})
	//nolint:gosec // The integer overflow conversion is required here.
	return contributionReduceActionFlag | Contribution(productionIdx)
}

const (
	contributionProductionBits                 = 16
	contributionMaxProductionIdx               = (1 << contributionProductionBits) - 1
	contributionProductionIdxMask              = contributionMaxProductionIdx
	contributionReduceActionFlag  Contribution = 1 << contributionProductionBits
)

// IsShiftAction reports if the contribution shifts the conflicted terminal.
func (c Contribution) IsShiftAction() bool {
	return !c.IsReduceAction()
}

// IsReduceAction reports if the contribution reduces a production on the conflicted terminal.
func (c Contribution) IsReduceAction() bool {
	return c&contributionReduceActionFlag != 0
}

// ProductionIdx returns the production index the contribution reduces. It is only meaningful when the contribution is a
// reduce action.
func (c Contribution) ProductionIdx() int {
	return int(c & contributionProductionIdxMask)
}

// Contribution implements fmt.Stringer.
var _ fmt.Stringer = (*Contribution)(nil)

// String returns a string representation.
func (c Contribution) String() string {
	if c.IsShiftAction() {
		return "shift"
	}
	return fmt.Sprintf("reduce production %d", c.ProductionIdx())
}
