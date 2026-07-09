package golr

import (
	"errors"

	"github.com/backbone81/golr/internal/utils"
)

type ConflictContribution uint32

func NewShiftConflictContribution() ConflictContribution {
	return ConflictContribution(0)
}

func NewReduceConflictContribution(productionIdx int) ConflictContribution {
	utils.DebugAssert(func() error {
		if productionIdx < 0 || conflictContributionMaxProductionIdx < productionIdx {
			return errors.New("production index out of bounds")
		}
		return nil
	})
	//nolint:gosec // The integer overflow conversion is required here.
	return conflictContributionReduceActionFlag | ConflictContribution(productionIdx)
}

const (
	conflictContributionProductionBits                         = 16
	conflictContributionMaxProductionIdx                       = (1 << conflictContributionProductionBits) - 1
	conflictContributionProductionIdxMask                      = conflictContributionMaxProductionIdx
	conflictContributionReduceActionFlag  ConflictContribution = 1 << conflictContributionProductionBits
)

func (c ConflictContribution) IsShiftAction() bool {
	return !c.IsReduceAction()
}

func (c ConflictContribution) IsReduceAction() bool {
	return c&conflictContributionReduceActionFlag != 0
}

func (c ConflictContribution) ProductionIdx() int {
	return int(c & conflictContributionProductionIdxMask)
}
