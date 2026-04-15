package bison

import (
	"encoding/xml"
	"os"
)

func LoadBisonXMLReportFromFile(filePath string) (BisonXMLReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return BisonXMLReport{}, err
	}

	var report BisonXMLReport
	if err := xml.Unmarshal(data, &report); err != nil {
		return BisonXMLReport{}, err
	}
	return report, nil
}

type BisonXMLReport struct {
	Version   string
	BugReport string
	URL       string
	Filename  string
}

type Grammar struct {
	Rules        []Rule
	Terminals    []Terminal
	Nonterminals []Nonterminal
	Automaton    Automaton
}

type Rule struct {
	Number     int
	Usefulness string
	Lhs        string
	Rhs        []string
}

type Terminal struct {
	SymbolNumber int
	TokenNumber  int
	Name         string
	Type         string
	Usefulness   string
}

type Nonterminal struct {
	SymbolNumber int
	Name         string
	Type         string
	Usefulness   string
}

type Automaton struct {
	States []State
}

type State struct {
	Number      int
	ItemSet     []Item
	Transitions []Transition
	Reductions  []Reduction
}

type Item struct {
	RuleNumber int
	Dot        int
}

type Transition struct {
	Type   string
	Symbol string
	State  int
}

type Reduction struct {
	Symbol  string
	Rule    int
	Enabled bool
}
