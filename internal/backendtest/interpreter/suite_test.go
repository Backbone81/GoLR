package interpreter_test

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/backendtest/interpreter"
	parserbackend "github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	parsergolrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/scannergen/backend"
	subsetcore "github.com/backbone81/golr/internal/scannergen/core/subset"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/scannergen/frontend/golr"
)

func TestInterpreter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interpreter Suite")
}

// golrSpecPath is the scanner specification the GoLR frontend is generated from. It is used as a real-world scanner of
// non-trivial size, in addition to the rules built by hand.
const golrSpecPath = "../../parsergen/frontend/golr/spec/golr.golr"

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

// expectScannerTrace asserts that scanning the given source produces exactly the given trace lines. The whole trace is
// compared as one string, because the point of the trace is that a backend is judged by a plain text diff.
func expectScannerTrace(dfa backend.DFA, source string, lines ...string) {
	Expect(interpreter.ScanTrace(dfa, []byte(source)).String()).To(Equal(strings.Join(append(lines, ""), "\n")))
}

// specToParser builds the scanner DFA and the parser tables of a GoLR specification the same way the generators do. The
// specification carries the scanner rules and the grammar in one document, which is what makes a case in the corpus a
// single file.
//
// The core is named explicitly rather than left to the default, for the same reason scripts/generate.sh pins one: a
// change of the default core would otherwise renumber symbols and productions underneath every expectation here.
func specToParser(spec string) (parserbackend.Parser, backend.DFA) {
	rules, grammar, err := parsergolrfrontend.GrammarFromString(spec)
	Expect(err).ToNot(HaveOccurred())

	parser, _, err := ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)
	Expect(err).ToNot(HaveOccurred())
	Expect(parser.States).ToNot(BeEmpty())

	return parser, rulesToDFA(rules...)
}

// expectParserTrace asserts that scanning and parsing the given source with the given specification produces exactly
// the given trace lines.
func expectParserTrace(spec string, source string, lines ...string) {
	parser, dfa := specToParser(spec)
	Expect(interpreter.ParseTrace(parser, dfa, []byte(source)).String()).
		To(Equal(strings.Join(append(lines, ""), "\n")))
}
