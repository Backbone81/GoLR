package backend

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/backbone81/golr/internal/utils"
)

// Core is the core of an LR(1) item consisting of a production index and a position within that production. The values
// for the production index and the position must be in the range of [0, 65535].
//
// It is implemented as a single unsigned integer to allow for a more compact representation and to enable easy
// sorting when dealing with a slice of cores.
type Core uint32

const (
	coreProductionBits = 16
	coreMaxProduction  = (1 << coreProductionBits) - 1

	corePositionBits = 16
	coreMaxPosition  = (1 << corePositionBits) - 1
	corePositionMask = coreMaxPosition
)

// NewCore creates a new core with the given production index and the position.
func NewCore(productionIdx int, position int) Core {
	utils.AssertValidIndex(productionIdx, coreMaxProduction)
	utils.AssertValidIndex(position, coreMaxPosition)
	// NOTE: We want to have the production index in the upper half of the Core and the position in the lower half.
	// That way we automatically get a sensible order when sorting by the value of the Core (i.e. first by production
	// index and second by position).
	//nolint:gosec // no integer overflow on correct usage
	return Core(productionIdx)<<corePositionBits | Core(position)
}

// ProductionIdx returns the production index of the Core.
func (c Core) ProductionIdx() int {
	return int(c >> corePositionBits)
}

// Position returns the position of the Core.
func (c Core) Position() int {
	return int(c & corePositionMask)
}

// Core implements fmt.Stringer.
var _ fmt.Stringer = (*Core)(nil)

// String returns a string representation.
func (c Core) String() string {
	return fmt.Sprintf("(production %d, position %d)", c.ProductionIdx(), c.Position())
}

type coreMarshal struct {
	ProductionIdx int `json:"productionIdx" yaml:"productionIdx"`
	Position      int `json:"position"      yaml:"position"`
}

// MarshalJSON implements the json.Marshaler interface.
func (c Core) MarshalJSON() ([]byte, error) {
	repr := coreMarshal{
		ProductionIdx: c.ProductionIdx(),
		Position:      c.Position(),
	}
	return json.Marshal(repr)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Core) UnmarshalJSON(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	var repr coreMarshal
	err := json.Unmarshal(b, &repr)
	if err != nil {
		return err
	}
	*c = NewCore(repr.ProductionIdx, repr.Position)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (c Core) MarshalYAML() ([]byte, error) {
	repr := coreMarshal{
		ProductionIdx: c.ProductionIdx(),
		Position:      c.Position(),
	}
	return yaml.Marshal(repr)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *Core) UnmarshalYAML(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	var repr coreMarshal
	err := yaml.Unmarshal(b, &repr)
	if err != nil {
		return err
	}
	*c = NewCore(repr.ProductionIdx, repr.Position)
	return nil
}
