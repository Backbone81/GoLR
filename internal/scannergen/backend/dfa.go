package backend

import "github.com/backbone81/golr/internal/scannergen/frontend"

// DFA is a deterministic finite automata.
type DFA struct {
	Rules  []frontend.Rule `json:"rules"  yaml:"rules"`
	States []State         `json:"states" yaml:"states"`
}
