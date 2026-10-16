package json_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
	subset "github.com/backbone81/golr/internal/scannergen/core/subset"
	"github.com/backbone81/golr/internal/scannergen/frontend/json"
)

// FuzzJSONRules reads arbitrary JSON documents as scanner rules. Rules which read have to read the same after
// writing them, and building their DFA has to work.
func FuzzJSONRules(f *testing.F) {
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
		document, err := json.RulesToString(rules)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(document)
	}
	f.Add(`[{"name":"A","regex":{"kind":"Literal"}}]`)
	f.Add(`[{"name":"A","regex":{"kind":"Literal","literal":{"text":""}}}]`)
	f.Add(`[{"name":"A","regex":{"kind":"Repetition","repetition":{"minimum":2,"maximum":1,"child":null}}}]`)
	f.Add(`[{"name":"A","regex":{"kind":"CharClass","charClass":{"ranges":[{"low":98,"high":97}]}}}]`)

	f.Fuzz(func(t *testing.T, document string) {
		rules, err := json.RulesFromString(document)
		if err != nil {
			return
		}

		written, err := json.RulesToString(rules)
		if err != nil {
			t.Fatalf("rules do not write: %v\n%s", err, document)
		}
		writtenRules, err := json.RulesFromString(written)
		if err != nil {
			t.Fatalf("written rules do not read: %v\n%s\nwritten from\n%s", err, written, document)
		}
		if !reflect.DeepEqual(writtenRules, rules) {
			t.Fatalf("writing changes the rules\n%s\nto\n%s", document, written)
		}

		subset.RulesToDFA(rules)
	})
}
