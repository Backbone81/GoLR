package table_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	subsetcore "github.com/backbone81/golr/internal/scannergen/core/subset"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/scannergen/frontend/golr"
)

func TestTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table Suite")
}

// golrSpecPath is the scanner specification the GoLR frontend is generated from. It is used as a real-world scanner of
// non-trivial size, in addition to the rules built by hand.
const golrSpecPath = "../../../parsergen/frontend/golr/spec/golr.golr"

// rulesToDFA builds the DFA for the given rules the same way the scanner generator does.
func rulesToDFA(rules ...frontend.Rule) backend.DFA {
	return subsetcore.RulesToDFA(rules)
}

// golrSpecDFA builds the DFA for the scanner of the GoLR specification.
func golrSpecDFA() backend.DFA {
	rules, _, err := golrfrontend.RulesFromFile(golrSpecPath)
	Expect(err).ToNot(HaveOccurred())
	Expect(rules).ToNot(BeEmpty())

	return rulesToDFA(rules...)
}
