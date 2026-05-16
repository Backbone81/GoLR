package frontend

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"

	"github.com/backbone81/golr/internal/utils"
)

// SymbolRef is either a terminal index or a nonterminal index. The most significant bit is used to signal a
// nonterminal. The maximum terminal or nonterminal index which can be stored is 32767.
type SymbolRef uint16

const (
	// symbolRefNonterminalBit is the most significant bit which is set when the SymbolRef stores a nonterminal index.
	symbolRefNonterminalBit = 1 << 15

	// symbolRefMask is the bitmask for selecting only the index from SymbolRef.
	symbolRefMask = symbolRefNonterminalBit - 1

	// symbolRefMaxIdx is the maximum index possible to store in SymbolRef.
	symbolRefMaxIdx = symbolRefMask
)

// NewTerminalRef creates a new SymbolRef for a terminal index.
func NewTerminalRef(idx int) SymbolRef {
	utils.AssertValidIndex(idx, symbolRefMaxIdx)
	return SymbolRef(idx) //nolint:gosec // We already have an assertion in place to test for the correct range.
}

// NewNonterminalRef creates a new SymbolRef for a nonterminal index.
func NewNonterminalRef(idx int) SymbolRef {
	utils.AssertValidIndex(idx, symbolRefMaxIdx)

	//nolint:gosec // We already have an assertion in place to test for the correct range.
	return SymbolRef(symbolRefNonterminalBit | idx)
}

// IsTerminal reports if the SymbolRef is holding a terminal index.
func (s SymbolRef) IsTerminal() bool {
	return !s.IsNonterminal()
}

// IsNonterminal reports if the SymbolRef is holding a nonterminal index.
func (s SymbolRef) IsNonterminal() bool {
	return s&symbolRefNonterminalBit != 0
}

// Idx returns the index for the terminal or nonterminal.
func (s SymbolRef) Idx() int {
	return int(s & symbolRefMask)
}

// SymbolRef implements fmt.Stringer.
var _ fmt.Stringer = (*SymbolRef)(nil)

// String returns a string representation.
func (s SymbolRef) String() string {
	if s.IsNonterminal() {
		return fmt.Sprintf("nonterminal %d", s.Idx())
	}
	return fmt.Sprintf("terminal %d", s.Idx())
}

// symbolIdxMarshal is a helper struct which is only used for marshaling.
type symbolIdxMarshal struct {
	Nonterminal bool `json:"nonterminal" yaml:"nonterminal"`
	Index       int  `json:"index"       yaml:"index"`
}

// MarshalJSON implements the json.Marshaler interface.
func (s SymbolRef) MarshalJSON() ([]byte, error) {
	repr := symbolIdxMarshal{
		Nonterminal: s.IsNonterminal(),
		Index:       s.Idx(),
	}
	return json.Marshal(repr)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SymbolRef) UnmarshalJSON(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	var repr symbolIdxMarshal
	err := json.Unmarshal(b, &repr)
	if err != nil {
		return err
	}
	if repr.Nonterminal {
		*s = NewNonterminalRef(repr.Index)
	} else {
		*s = NewTerminalRef(repr.Index)
	}
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (s SymbolRef) MarshalYAML() ([]byte, error) {
	repr := symbolIdxMarshal{
		Nonterminal: s.IsNonterminal(),
		Index:       s.Idx(),
	}
	return yaml.Marshal(repr)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *SymbolRef) UnmarshalYAML(b []byte) error {
	if slices.Equal(b, []byte("null")) {
		return nil
	}
	var repr symbolIdxMarshal
	err := yaml.Unmarshal(b, &repr)
	if err != nil {
		return err
	}
	if repr.Nonterminal {
		*s = NewNonterminalRef(repr.Index)
	} else {
		*s = NewTerminalRef(repr.Index)
	}
	return nil
}
