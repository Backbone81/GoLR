package bison

import (
	"encoding/xml"
	"os"
)

func LoadBisonXMLReportFromFile(filePath string) (BisonXMLReport, error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
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
	XMLName   xml.Name  `xml:"bison-xml-report"`
	Version   string    `xml:"version,attr"`
	BugReport string    `xml:"bug-report,attr"`
	URL       string    `xml:"url,attr"`
	Filename  string    `xml:"filename"`
	Grammar   Grammar   `xml:"grammar"`
	Automaton Automaton `xml:"automaton"`
}

type Grammar struct {
	Rules        []Rule        `xml:"rules>rule"`
	Terminals    []Terminal    `xml:"terminals>terminal"`
	Nonterminals []Nonterminal `xml:"nonterminals>nonterminal"`
}

type Rule struct {
	Number     int      `xml:"number,attr"`
	Usefulness string   `xml:"usefulness,attr"`
	Lhs        string   `xml:"lhs"`
	Rhs        []string `xml:"rhs>symbol"`
}

type Terminal struct {
	SymbolNumber int    `xml:"symbol-number,attr"`
	TokenNumber  int    `xml:"token-number,attr"`
	Name         string `xml:"name,attr"`
	Type         string `xml:"type,attr"`
	Usefulness   string `xml:"usefulness,attr"`
}

type Nonterminal struct {
	SymbolNumber int    `xml:"symbol-number,attr"`
	Name         string `xml:"name,attr"`
	Type         string `xml:"type,attr"`
	Usefulness   string `xml:"usefulness,attr"`
}

type Automaton struct {
	States []State `xml:"state"`
}

type State struct {
	Number      int          `xml:"number,attr"`
	ItemSet     []Item       `xml:"itemset>item"`
	Transitions []Transition `xml:"actions>transitions>transition"`
	Reductions  []Reduction  `xml:"actions>reductions>reduction"`
}

type Item struct {
	RuleNumber int `xml:"rule-number,attr"`
	Dot        int `xml:"dot,attr"`
}

func (i Item) IsKernelItem() bool {
	return i.Dot > 0 || i.RuleNumber == 0
}

type Transition struct {
	Type   string `xml:"type,attr"`
	Symbol string `xml:"symbol,attr"`
	State  int    `xml:"state,attr"`
}

type Reduction struct {
	Symbol  string `xml:"symbol,attr"`
	Rule    string `xml:"rule,attr"`
	Enabled bool   `xml:"enabled,attr"`
}
