package json_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend/json"
)

// FuzzJSONGrammar reads arbitrary JSON documents as grammars. A grammar which reads has to read the same after
// writing it, and building its parser has to return a parser or an error.
func FuzzJSONGrammar(f *testing.F) {
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
		document, err := json.GrammarToString(grammar)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(document)
	}
	f.Add(`{"terminals":[{"name":"a"}],"nonterminals":[{"name":"s"}],"productions":[` +
		`{"nonterminalIdx":0,"symbolRefs":[{"nonterminal":false,"index":-1}]}],"startNonterminalIdx":0}`)
	f.Add(`{"terminals":[{"name":"a"}],"nonterminals":[{"name":"s"}],"productions":[` +
		`{"nonterminalIdx":0,"symbolRefs":[{"nonterminal":true,"index":-1}]}],"startNonterminalIdx":0}`)
	f.Add(`{"terminals":[{"name":"a"}],"nonterminals":[{"name":"s"}],"productions":[` +
		`{"nonterminalIdx":0,"symbolRefs":[{"nonterminal":false,"index":0}],"precedenceTerminalIdx":5}],` +
		`"startNonterminalIdx":0}`)
	f.Add(`{"terminals":[{"name":"a","associativity":"left","precedence":-1}],"nonterminals":[{"name":"s"}],` +
		`"productions":[{"nonterminalIdx":0,"symbolRefs":[null]}],"startNonterminalIdx":0}`)

	f.Fuzz(func(t *testing.T, document string) {
		grammar, err := json.GrammarFromString(document)
		if err != nil {
			return
		}

		written, err := json.GrammarToString(grammar)
		if err != nil {
			t.Fatalf("grammar does not write: %v\n%s", err, document)
		}
		writtenGrammar, err := json.GrammarFromString(written)
		if err != nil {
			t.Fatalf("written grammar does not read: %v\n%s\nwritten from\n%s", err, written, document)
		}
		if !reflect.DeepEqual(writtenGrammar, grammar) {
			t.Fatalf("writing changes the grammar\n%s\nto\n%s", document, written)
		}

		_, _, _, _ = ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)
	})
}
