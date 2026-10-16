package convert_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/backbone81/golr/internal/convert"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/internal/parsergen/frontend/bison"
	"github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// FuzzBisonToGoLR converts arbitrary Bison grammars. Every valid grammar the Bison frontend reads and the converter
// accepts has to convert to a GoLR grammar file which reads back to the same grammar.
//
// Writing and reading the GoLR file is the oracle for the converted grammar being valid GoLR: Bison allows names,
// aliases and string tokens which GoLR does not, and a conversion which lets one through shows as a file which does
// not read or reads as a different grammar. FuzzGoLRFrontend can not find these, as the GoLR reader never produces
// such a grammar.
func FuzzBisonToGoLR(f *testing.F) {
	paths, err := filepath.Glob("../parsergen/core/ielr1/golr/testdata/*.y")
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(source)
	}
	for _, seed := range []string{
		"%token NUM \"number\" PLUS '+' STR _(\"string\")\n%left '+'\n%right '^'\n%%\n" +
			"e: e '+' e | e '^' e | '-' e %prec '^' | NUM | STR | '\\n' | '\"' | '\\'';\n",
		"%start top\n%%\nlist.opt: %empty | list;\nlist: item | list item;\nitem: \"x\" | error ';';\ntop: list.opt;\n",
		"%token a-b c.d\n%%\ns: a-b c.d | s.1;\ns.1: s_1 ;\ns_1: c.d;\n",
		"%%\ns: a; t: b; s: c;\na: 'a'; b: 'b'; c: 'c';\n",
		"%precedence LOW\n%nonassoc '<' '>'\n%%\ns: s '<' s | s '>' s | 'x' %prec LOW;\n%%\nint main(void) {}\n",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		grammar, err := bison.ToGrammar(bytes.NewReader(source), "fuzz.y")
		if err != nil {
			return
		}
		// The conversion changes the symbols in place.
		usedErrorSymbol := usesErrorSymbol(grammar)
		converted, err := convert.BisonGrammar2GoLR(grammar)
		if err != nil {
			return
		}
		if _, err := converted.Validate(); err != nil {
			return
		}

		written, err := golr.GrammarToString(nil, converted)
		if err != nil {
			t.Fatalf("converted grammar does not write: %v\n%s", err, source)
		}
		_, reread, err := golr.GrammarFromString(written)
		if err != nil {
			t.Fatalf("converted grammar does not read: %v\n%s\nconverted from\n%s", err, written, source)
		}
		if usesErrorSymbol(reread) != usedErrorSymbol {
			t.Fatalf("converted grammar does not keep the use of the error symbol\n%s\nconverted from\n%s", written,
				source)
		}
		if got, want := describeGrammar(reread), describeGrammar(converted); got != want {
			t.Fatalf("converted grammar reads as\n%s\ninstead of\n%s\nfrom\n%s\nconverted from\n%s", got, want, written,
				source)
		}
	})
}

// describeGrammar describes the grammar independent of the order of its symbols, the order of the nonterminals of
// its productions, and the numbers of its precedence levels, which the GoLR format does not keep. A terminal without
// precedence keeps rank 0, the levels count from 1. The error symbol is not described and does not count.
func describeGrammar(grammar frontend.Grammar) string {
	precedenceRanks := map[int]int{0: 0}
	var precedences []int
	for _, terminal := range grammar.Terminals {
		if terminal.Name != frontend.SymbolError.Name && terminal.Precedence != 0 {
			precedences = append(precedences, terminal.Precedence)
		}
	}
	slices.Sort(precedences)
	for _, precedence := range slices.Compact(precedences) {
		precedenceRanks[precedence] = len(precedenceRanks)
	}

	var terminals []string
	for _, terminal := range grammar.Terminals {
		if terminal.Name == frontend.SymbolError.Name {
			continue
		}
		terminals = append(terminals, fmt.Sprintf("%s %s %v %d", terminal.Name, terminal.Alias,
			terminal.Associativity, precedenceRanks[terminal.Precedence]))
	}
	slices.Sort(terminals)

	productionsByNonterminal := make(map[string][]string)
	for _, production := range grammar.Productions {
		var description strings.Builder
		for _, symbolRef := range production.SymbolRefs {
			description.WriteString(" " + symbolName(grammar, symbolRef))
		}
		if production.PrecedenceTerminalIdx != nil {
			description.WriteString(" %prec " + grammar.Terminals[*production.PrecedenceTerminalIdx].Name)
		}
		if production.Name != nil {
			description.WriteString(" @name " + *production.Name)
		}
		nonterminal := grammar.Nonterminals[production.NonterminalIdx].Name
		productionsByNonterminal[nonterminal] = append(productionsByNonterminal[nonterminal], description.String())
	}
	var productions []string
	for nonterminal, descriptions := range productionsByNonterminal {
		productions = append(productions, nonterminal+":"+strings.Join(descriptions, " |"))
	}
	slices.Sort(productions)

	return fmt.Sprintf("start %s\n%s\n%s", grammar.Nonterminals[grammar.StartNonterminalIdx].Name,
		strings.Join(terminals, "\n"), strings.Join(productions, "\n"))
}

// symbolName returns the name of the terminal or nonterminal the symbol reference refers to.
func symbolName(grammar frontend.Grammar, symbolRef frontend.SymbolRef) string {
	if symbolRef.IsTerminal() {
		return grammar.Terminals[symbolRef.Idx()].Name
	}
	return grammar.Nonterminals[symbolRef.Idx()].Name
}

// usesErrorSymbol reports whether a production of the grammar refers to the error symbol.
func usesErrorSymbol(grammar frontend.Grammar) bool {
	errorRef, ok := frontend.ErrorTerminalRef(grammar)
	if !ok {
		return false
	}
	for _, production := range grammar.Productions {
		if slices.Contains(production.SymbolRefs, errorRef) {
			return true
		}
	}
	return false
}
