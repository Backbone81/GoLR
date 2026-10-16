package yaml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	subset "github.com/backbone81/golr/internal/scannergen/core/subset"
	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/scannergen/frontend/yaml"
)

// FuzzYAMLRules reads arbitrary YAML documents as scanner rules. Rules which read have to read the same after
// writing them, with an empty list the same as none, and building their DFA has to work. YAML writes a missing list
// as an empty one, which is why the comparison sets empty lists to nil first, see rulesEmptyListsToNil.
func FuzzYAMLRules(f *testing.F) {
	paths, err := filepath.Glob("../../../backendtest/golden/*/spec.golr")
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		rules, _, err := golrfrontend.GrammarFromString(string(source))
		if err != nil {
			f.Fatal(err)
		}
		document, err := yaml.RulesToString(rules)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(document)
	}
	f.Add("- name: A\n  regex:\n    kind: Literal\n")
	f.Add("- name: A\n  regex:\n    kind: Literal\n    literal:\n      text: \"\"\n")
	f.Add("- name: A\n  regex:\n    kind: Repetition\n    repetition:\n      minimum: 2\n      maximum: 1\n")

	f.Fuzz(func(t *testing.T, document string) {
		rules, err := yaml.RulesFromString(document)
		if err != nil {
			return
		}

		written, err := yaml.RulesToString(rules)
		if err != nil {
			t.Fatalf("rules do not write: %v\n%s", err, document)
		}
		writtenRules, err := yaml.RulesFromString(written)
		if err != nil {
			t.Fatalf("written rules do not read: %v\n%s\nwritten from\n%s", err, written, document)
		}

		// The DFA is built before the comparison sets the empty lists to nil, so that it sees the rules as read.
		subset.RulesToDFA(rules)

		rules = rulesEmptyListsToNil(rules)
		writtenRules = rulesEmptyListsToNil(writtenRules)
		if !reflect.DeepEqual(writtenRules, rules) {
			t.Fatalf("writing changes the rules\n%s\nto\n%s", document, written)
		}
	})
}

// rulesEmptyListsToNil sets every empty list of the rules and their regular expressions to nil, which makes a missing
// list and an empty one compare equal.
func rulesEmptyListsToNil(rules []frontend.Rule) []frontend.Rule {
	for ruleIdx := range rules {
		nodeEmptyListsToNil(&rules[ruleIdx].Regex)
	}
	return emptyToNil(rules)
}

// nodeEmptyListsToNil sets every empty list of the node and the nodes below it to nil. The fields of the kinds the node
// is not of are zero, which makes going through all of them the same as going through the one of its kind.
func nodeEmptyListsToNil(node *frontend.Node) {
	if node == nil {
		return
	}
	node.CharClass.Ranges = emptyToNil(node.CharClass.Ranges)
	node.Concat.Children = emptyToNil(node.Concat.Children)
	node.Or.Children = emptyToNil(node.Or.Children)

	for _, child := range node.Concat.Children {
		nodeEmptyListsToNil(child)
	}
	for _, child := range node.Or.Children {
		nodeEmptyListsToNil(child)
	}
	nodeEmptyListsToNil(node.OneOrMore.Child)
	nodeEmptyListsToNil(node.Optional.Child)
	nodeEmptyListsToNil(node.Repetition.Child)
	nodeEmptyListsToNil(node.ZeroOrMore.Child)
}

// emptyToNil returns nil for an empty slice and the slice itself otherwise.
func emptyToNil[T any](slice []T) []T {
	if len(slice) == 0 {
		return nil
	}
	return slice
}
