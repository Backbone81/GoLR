package backend

import "golr/internal/scannergen/frontend"

// DFA is a deterministic finite automata.
type DFA struct {
	Rules  []frontend.Rule `json:"rules" yaml:"rules"`
	States []State         `json:"states" yaml:"states"`
}
