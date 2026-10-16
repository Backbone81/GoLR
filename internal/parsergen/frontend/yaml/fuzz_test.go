package yaml_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend/yaml"
)

// FuzzYAMLGrammar reads arbitrary YAML documents as grammars. A grammar which reads has to read the same after
// writing it, with an empty list the same as none, and building its parser has to return a parser or an error. YAML
// writes a missing list as an empty one, which is why the comparison sets empty lists to nil first, see
// emptyListsToNil.
func FuzzYAMLGrammar(f *testing.F) {
	paths, err := filepath.Glob("../../../backendtest/golden/*/spec.golr")
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		_, grammar, err := golrfrontend.GrammarFromString(string(source))
		if err != nil {
			f.Fatal(err)
		}
		document, err := yaml.GrammarToString(grammar)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(document)
	}
	f.Add("terminals: [{name: a}]\nnonterminals: [{name: s}]\nproductions:\n" +
		"- {nonterminalIdx: 0, symbolRefs: [{nonterminal: false, index: -1}]}\nstartNonterminalIdx: 0\n")
	f.Add("terminals: [{name: a}]\nnonterminals: [{name: s}]\nproductions:\n" +
		"- {nonterminalIdx: 0, symbolRefs: [{nonterminal: true, index: -1}]}\nstartNonterminalIdx: 0\n")
	f.Add("terminals: [{name: a}]\nnonterminals: [{name: s}]\nproductions:\n" +
		"- {nonterminalIdx: 0, symbolRefs: [{nonterminal: false, index: 0}], precedenceTerminalIdx: 5}\n" +
		"startNonterminalIdx: 0\n")

	f.Fuzz(func(t *testing.T, document string) {
		grammar, err := yaml.GrammarFromString(document)
		if err != nil {
			return
		}

		written, err := yaml.GrammarToString(grammar)
		if err != nil {
			t.Fatalf("grammar does not write: %v\n%s", err, document)
		}
		writtenGrammar, err := yaml.GrammarFromString(written)
		if err != nil {
			t.Fatalf("written grammar does not read: %v\n%s\nwritten from\n%s", err, written, document)
		}

		// The parser is built before the comparison sets the empty lists to nil, so that it sees the grammar as read.
		_, _, _, _ = ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)

		emptyListsToNil(&grammar)
		emptyListsToNil(&writtenGrammar)
		if !reflect.DeepEqual(writtenGrammar, grammar) {
			t.Fatalf("writing changes the grammar\n%s\nto\n%s", document, written)
		}
	})
}

// emptyListsToNil sets every empty list of the grammar to nil, which makes a missing list and an empty one compare
// equal.
func emptyListsToNil(grammar *frontend.Grammar) {
	grammar.Terminals = emptyToNil(grammar.Terminals)
	grammar.Nonterminals = emptyToNil(grammar.Nonterminals)
	grammar.Productions = emptyToNil(grammar.Productions)
	for productionIdx := range grammar.Productions {
		production := &grammar.Productions[productionIdx]
		production.SymbolRefs = emptyToNil(production.SymbolRefs)
	}
}

// emptyToNil returns nil for an empty slice and the slice itself otherwise.
func emptyToNil[T any](slice []T) []T {
	if len(slice) == 0 {
		return nil
	}
	return slice
}
