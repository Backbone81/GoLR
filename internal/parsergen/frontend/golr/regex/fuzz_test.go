package regex_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr/regex"
	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	subset "github.com/backbone81/golr/internal/scannergen/core/subset"
	"github.com/backbone81/golr/internal/scannergen/frontend"
)

// seedPatterns are regular expressions without their / delimiters, covering the syntax of the regex parser.
var seedPatterns = []string{
	"a", "ab", "a|b", "(abc)*", "a+b?c*", ".", `a\*`, `\n\r\t\v\f\0`, `\d\w\s\D\W\S`,
	"[ab-d]", "[^ab-d]", "[-ab]", "[ab-]", `[ab\]]`, `[a\-b]`, `[\d\s_]`, "[[:alpha:]]", "[^[:alpha:]]", "[a[:digit:]]",
	"[[:space:][:punct:]]", "a{2}", "a{2,5}", "a{2,}", "a{,5}", "a{0,1}", "(a|bc){1,3}d", "ä|😀+", "[ä-ü]", "[^\n]*",
	"/", "{F}", "{F}b", "a{F}", "{F}{F}", "({F})*", "{F}{2}", "x|{F}c",
}

// FuzzRegexFragment parses a regular expression which refers to a fragment. A reference to the fragment has to give
// the regular expression the fragment gives on its own, however the expression after it uses the fragment, and every
// regular expression the parser accepts has to be valid.
func FuzzRegexFragment(f *testing.F) {
	for _, pattern := range seedPatterns {
		f.Add(pattern, "f", true)
		f.Add(pattern, "[0-9]+", false)
	}
	f.Add("{F}b", "", false)
	f.Add("{F}b", "", true)

	f.Fuzz(func(t *testing.T, pattern string, fragment string, literal bool) {
		// A string fragment is a string token of the GoLR scanner, which ends at the first quote.
		var lexeme string
		switch {
		case literal && !strings.Contains(fragment, `"`):
			lexeme = `"` + fragment + `"`
		case literal:
			return
		case fragment != "":
			lexeme = "/" + fragment + "/"
		}
		fragments := map[string][]byte{"F": []byte(lexeme)}

		alone, err := regex.Parse([]byte("/{F}/"), fragments)
		if err != nil {
			mustBeValid(t, "/"+pattern+"/", fragments)
			return
		}
		if err := alone.Validate(); err != nil {
			t.Fatalf("fragment %q gives the invalid %s: %v", lexeme, alone, err)
		}

		// The first alternative is the node of the fragment, which the parser has completed before it reads the
		// pattern. The pattern can only change it by changing that node in place.
		node := mustBeValid(t, "/{F}|"+pattern+"/", fragments)
		if node == nil {
			return
		}
		if node.Kind != frontend.KindOr {
			t.Fatalf("%q is no alternation: %s", pattern, node)
		}
		if got := node.Or.Children[0].String(); got != alone.String() {
			t.Fatalf("fragment %q is %s in front of %q, but %s on its own", lexeme, got, pattern, alone)
		}
	})
}

// mustBeValid parses the regular expression and checks that the result is valid. It returns nil if the parser
// refuses the regular expression.
func mustBeValid(t *testing.T, pattern string, fragments map[string][]byte) *frontend.Node {
	t.Helper()

	node, err := regex.Parse([]byte(pattern), fragments)
	if err != nil {
		return nil
	}
	if err := node.Validate(); err != nil {
		t.Fatalf("%q gives the invalid %s: %v", pattern, node, err)
	}
	return node
}

// FuzzDFA builds the DFA of two rules and runs it on an input. The longest prefix it accepts and the rule it accepts
// it for have to be the ones Go's regexp finds for the same regular expressions, and the compressed tables the table
// driven backends generate from the DFA have to find the same.
func FuzzDFA(f *testing.F) {
	for idx, pattern := range seedPatterns {
		f.Add(pattern, seedPatterns[(idx+1)%len(seedPatterns)], "abc123 äöü\n😀 a+b-c ab")
	}
	f.Add("[a-z]+", "if", "if")
	f.Add("if", "[a-z]+", "iffy")
	f.Add("a*", "b", "")

	f.Fuzz(func(t *testing.T, first string, second string, input string) {
		if !utf8.ValidString(input) || len(input) > 256 {
			return
		}

		var rules []frontend.Rule
		var expressions []*regexp.Regexp
		for idx, pattern := range []string{first, second} {
			if !utf8.ValidString(pattern) {
				return
			}
			node := mustBeValid(t, "/"+pattern+"/", nil)
			if node == nil {
				if idx == 0 {
					return
				}
				continue
			}
			if automatonSize(node) > maxAutomatonSize {
				return
			}
			expression, err := regexp.Compile(`^(?:` + goSyntax(node) + `)$`)
			if err != nil {
				t.Fatalf("%s translates to invalid Go syntax %s: %v", node, goSyntax(node), err)
			}
			rules = append(rules, frontend.Rule{Name: fmt.Sprintf("R%d", idx), Regex: *node})
			expressions = append(expressions, expression)
		}

		dfa := subset.RulesToDFA(rules)
		length, ruleIdx := longestMatch(dfa.States, []byte(input))

		wantLength, wantRuleIdx := -1, -1
		for prefixLength := len(input); prefixLength >= 0 && wantLength < 0; prefixLength-- {
			if prefixLength < len(input) && !utf8.RuneStart(input[prefixLength]) {
				continue
			}
			for idx, expression := range expressions {
				if expression.MatchString(input[:prefixLength]) {
					wantLength, wantRuleIdx = prefixLength, idx
					break
				}
			}
		}

		if length != wantLength || ruleIdx != wantRuleIdx {
			t.Fatalf("rules %s on %q: DFA matches %d bytes for rule %d, regexp %d bytes for rule %d",
				describeRules(rules), input, length, ruleIdx, wantLength, wantRuleIdx)
		}

		compressed := table.NewCompressedDFA(dfa)
		if compressed.StateCount() != len(dfa.States) {
			t.Fatalf("rules %s: compressed DFA has %d states instead of %d",
				describeRules(rules), compressed.StateCount(), len(dfa.States))
		}
		compressedLength, compressedRuleIdx := compressedLongestMatch(&compressed, []byte(input))
		if compressedLength != length || compressedRuleIdx != ruleIdx {
			t.Fatalf("rules %s on %q: compressed DFA matches %d bytes for rule %d, DFA %d bytes for rule %d",
				describeRules(rules), input, compressedLength, compressedRuleIdx, length, ruleIdx)
		}
	})
}

// maxAutomatonSize bounds the automata FuzzDFA builds, because a repetition copies its child as often as it repeats.
const maxAutomatonSize = 64

// automatonSize estimates the size of the automaton for the node, with repetitions counted by their copies.
func automatonSize(node *frontend.Node) int {
	switch node.Kind {
	case frontend.KindAny, frontend.KindCharClass:
		return 1
	case frontend.KindLiteral:
		return utf8.RuneCountInString(node.Literal.Text)
	case frontend.KindConcat:
		return sumAutomatonSizes(node.Concat.Children)
	case frontend.KindOr:
		return sumAutomatonSizes(node.Or.Children)
	case frontend.KindOneOrMore:
		return automatonSize(node.OneOrMore.Child)
	case frontend.KindOptional:
		return automatonSize(node.Optional.Child)
	case frontend.KindZeroOrMore:
		return automatonSize(node.ZeroOrMore.Child)
	case frontend.KindRepetition:
		return max(1, min(node.Repetition.Maximum, maxAutomatonSize+1)) * automatonSize(node.Repetition.Child)
	default:
		panic(fmt.Sprintf("unknown node kind %v", node.Kind))
	}
}

func sumAutomatonSizes(nodes []*frontend.Node) int {
	size := 0
	for _, node := range nodes {
		size += automatonSize(node)
	}
	return size
}

// goSyntax translates the node into the syntax of Go's regexp.
func goSyntax(node *frontend.Node) string {
	switch node.Kind {
	case frontend.KindAny:
		// Neither matches a line feed.
		return "."
	case frontend.KindCharClass:
		if len(node.CharClass.Ranges) == 0 {
			if node.CharClass.Negate {
				return `[\x00-\x{10FFFF}]`
			}
			return `[^\x00-\x{10FFFF}]`
		}
		var class strings.Builder
		class.WriteString("[")
		if node.CharClass.Negate {
			class.WriteString("^")
		}
		for _, charRange := range node.CharClass.Ranges {
			fmt.Fprintf(&class, `\x{%x}-\x{%x}`, charRange.Low, min(charRange.High, unicode.MaxRune))
		}
		class.WriteString("]")
		return class.String()
	case frontend.KindLiteral:
		return regexp.QuoteMeta(node.Literal.Text)
	case frontend.KindConcat:
		var concat strings.Builder
		for _, child := range node.Concat.Children {
			concat.WriteString("(?:" + goSyntax(child) + ")")
		}
		return concat.String()
	case frontend.KindOr:
		alternatives := make([]string, 0, len(node.Or.Children))
		for _, child := range node.Or.Children {
			alternatives = append(alternatives, "(?:"+goSyntax(child)+")")
		}
		return strings.Join(alternatives, "|")
	case frontend.KindOneOrMore:
		return "(?:" + goSyntax(node.OneOrMore.Child) + ")+"
	case frontend.KindOptional:
		return "(?:" + goSyntax(node.Optional.Child) + ")?"
	case frontend.KindZeroOrMore:
		return "(?:" + goSyntax(node.ZeroOrMore.Child) + ")*"
	case frontend.KindRepetition:
		return fmt.Sprintf("(?:%s){%d,%d}", goSyntax(node.Repetition.Child), node.Repetition.Minimum,
			node.Repetition.Maximum)
	default:
		panic(fmt.Sprintf("unknown node kind %v", node.Kind))
	}
}

// longestMatch runs the DFA from the start of the input and returns the length of the longest prefix it accepts and
// the rule it accepts it for, or -1 for both if it accepts no prefix.
func longestMatch(states []backend.State, input []byte) (int, int) {
	length, ruleIdx := -1, -1
	stateIdx := 0
	if states[stateIdx].Accept {
		length, ruleIdx = 0, states[stateIdx].RuleIdx
	}
	for idx, value := range input {
		next := -1
		for _, transition := range states[stateIdx].Transitions {
			if transition.ByteRange.Low <= value && value <= transition.ByteRange.High {
				next = transition.StateIdx
				break
			}
		}
		if next < 0 {
			break
		}
		stateIdx = next
		if states[stateIdx].Accept {
			length, ruleIdx = idx+1, states[stateIdx].RuleIdx
		}
	}
	return length, ruleIdx
}

// compressedLongestMatch does what longestMatch does on the tables of a compressed DFA, the way a table driven scanner
// reads them.
func compressedLongestMatch(compressed *table.CompressedDFA, input []byte) (int, int) {
	length, ruleIdx := -1, -1
	stateIdx := 0
	if accept := compressed.AcceptRuleIdxByStateIdx[stateIdx]; accept != table.NoRule {
		length, ruleIdx = 0, accept
	}
	for idx, value := range input {
		stateIdx = compressed.Transition(stateIdx, value)
		if stateIdx == table.NoTransition {
			break
		}
		if accept := compressed.AcceptRuleIdxByStateIdx[stateIdx]; accept != table.NoRule {
			length, ruleIdx = idx+1, accept
		}
	}
	return length, ruleIdx
}

func describeRules(rules []frontend.Rule) string {
	descriptions := make([]string, 0, len(rules))
	for _, rule := range rules {
		descriptions = append(descriptions, fmt.Sprintf("%s: /%s/", rule.Name, rule.Regex.String()))
	}
	return strings.Join(descriptions, ", ")
}
