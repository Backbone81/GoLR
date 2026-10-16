package golr_test

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	golrfmt "github.com/backbone81/golr/internal/fmt"
	parsergenfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
	scannerfrontend "github.com/backbone81/golr/internal/scannergen/frontend"
)

// FuzzGoLRFrontend reads arbitrary grammar files. A grammar which reads has to read the same after formatting it,
// formatting it again has to change nothing, and writing it has to give a grammar file which reads back to the same
// grammar.
func FuzzGoLRFrontend(f *testing.F) {
	paths, err := filepath.Glob("../../../backendtest/golden/*/spec.golr")
	if err != nil {
		f.Fatal(err)
	}
	paths = append(paths, "../../../../examples/calculator/calculator.golr", "spec/golr.golr")
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(source))
	}
	f.Add("@scanner{A:\"a\";B:/[0-9]+/@skip;C:\"c\"@fragment;D:/{C}d/;}@parser{@start:s;s:A s|@empty|@error A @name(e);}")
	f.Add("@scanner{A:\"\\\"\";B:\"\\\\\";N:/\\n/;}@parser{s:A B N;}")
	f.Add("@scanner{A:\"a\";B:\"b\";}@parser{@precedence{@left:A;@right:B;@none:\"a\";}s:s A s|s B s @prec(A)|A;}")

	f.Fuzz(func(t *testing.T, source string) {
		rules, grammar, err := golr.GrammarFromString(source)
		if err != nil {
			return
		}

		formatted, err := golrfmt.GoLRString(source)
		if err != nil {
			t.Fatalf("grammar does not format: %v\n%s", err, source)
		}
		formattedRules, formattedGrammar, err := golr.GrammarFromString(formatted)
		if err != nil {
			t.Fatalf("formatted grammar does not read: %v\n%s", err, formatted)
		}
		if !reflect.DeepEqual(formattedRules, rules) || !reflect.DeepEqual(formattedGrammar, grammar) {
			t.Fatalf("formatting changes the grammar\n%s\nto\n%s", source, formatted)
		}
		reformatted, err := golrfmt.GoLRString(formatted)
		if err != nil {
			t.Fatalf("formatted grammar does not format: %v\n%s", err, formatted)
		}
		if reformatted != formatted {
			t.Fatalf("formatting is not idempotent\n%s\nformats to\n%s", formatted, reformatted)
		}

		written, err := golr.GrammarToString(rules, grammar)
		if err != nil {
			t.Fatalf("grammar does not write: %v\n%s", err, source)
		}
		writtenRules, writtenGrammar, err := golr.GrammarFromString(written)
		if err != nil {
			t.Fatalf("written grammar does not read: %v\n%s\nwritten from\n%s", err, written, source)
		}
		if !reflect.DeepEqual(rankPrecedences(writtenGrammar), rankPrecedences(grammar)) ||
			describeRules(writtenRules) != describeRules(rules) {
			t.Fatalf("writing changes the grammar\n%s\nto\n%s", source, written)
		}
	})
}

// rankPrecedences returns the grammar with the precedence of every terminal replaced by its rank, 0 for no precedence
// and counting from 1 for the levels. Writing keeps the order of the levels but not their numbers, which change when a
// terminal declared twice leaves a level empty.
func rankPrecedences(grammar parsergenfrontend.Grammar) parsergenfrontend.Grammar {
	var precedences []int
	for _, terminal := range grammar.Terminals {
		if terminal.Precedence != 0 {
			precedences = append(precedences, terminal.Precedence)
		}
	}
	slices.Sort(precedences)
	precedences = slices.Compact(precedences)

	grammar.Terminals = slices.Clone(grammar.Terminals)
	for idx, terminal := range grammar.Terminals {
		if terminal.Precedence != 0 {
			rank, _ := slices.BinarySearch(precedences, terminal.Precedence)
			grammar.Terminals[idx].Precedence = rank + 1
		}
	}
	return grammar
}

// describeRules describes the rules by what the writer keeps of them: the name, the regular expression and whether
// the scanner skips them.
func describeRules(rules []scannerfrontend.Rule) string {
	var description strings.Builder
	for _, rule := range rules {
		description.WriteString(rule.Name + " " + rule.Regex.String())
		if rule.Skip {
			description.WriteString(" skip")
		}
		description.WriteString("\n")
	}
	return description.String()
}
