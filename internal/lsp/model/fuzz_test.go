package model_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/backbone81/golr/internal/lsp/model"
)

// fuzzSeeds returns the grammar files of the repository which are small enough for fuzzing, and broken texts which
// lead the pass through its recovery.
func fuzzSeeds() ([][]byte, error) {
	paths, err := filepath.Glob("../../backendtest/golden/*/spec.golr")
	if err != nil {
		return nil, err
	}
	paths = append(paths,
		"../../../examples/calculator/calculator.golr",
		"../../parsergen/frontend/golr/spec/golr.golr",
	)
	seeds := [][]byte{[]byte(grammar)}
	for _, path := range paths {
		text, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		seeds = append(seeds, text)
	}
	for _, seed := range []string{
		"",
		`@scanner { A: "a" } @parser { s: A; }`,
		`@scanner { A: "a"; @parser { s: A; }`,
		`@scanner { A: "a" B: "b"; } @parser { s: A t: s A; }`,
		`@scanner { A: "a"; } @parser { @precedence { @left: A } s: A @precedence(A) @name(n) | @empty; }`,
		`@scanner { A: "a"; } @parser { @start: s; s: A"a" @ B @na`,
		"@scanner {\n    A: \"a;\n    B: /b\n}\n@parser { s: A // x\n /* y */ ; }",
		"@parser { s: A /* unterminated",
		"\xff\x00@scanner{}",
	} {
		seeds = append(seeds, []byte(seed))
	}
	return seeds, nil
}

func FuzzBuild(f *testing.F) {
	seeds, err := fuzzSeeds()
	if err != nil {
		f.Fatal(err)
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text []byte) {
		m := model.Build(text)
		if err := checkLists(text, m); err != nil {
			t.Fatal(err)
		}
		if err := checkOccurrences(text, m); err != nil {
			t.Fatal(err)
		}
		if err := checkLookups(m); err != nil {
			t.Fatal(err)
		}
		if err := checkOffsetQueries(text, m); err != nil {
			t.Fatal(err)
		}
		if err := checkRenames(m); err != nil {
			t.Fatal(err)
		}
	})
}

// checkSpan returns an error if the span is empty, does not lie within the text or does not hold the name.
func checkSpan(text []byte, name string, span model.Span) error {
	if span.Length <= 0 || span.Offset < 0 || len(text) < span.End() {
		return fmt.Errorf("span %+v of %q outside of the text of length %d", span, name, len(text))
	}
	if string(text[span.Offset:span.End()]) != name {
		return fmt.Errorf("span %+v holds %q instead of %q", span, text[span.Offset:span.End()], name)
	}
	return nil
}

// checkLists returns an error if a definition, alias or reference does not hold its name at its span, an alias is no
// string, or a list is not in the order of the text.
func checkLists(text []byte, m *model.Model) error {
	end := 0
	for _, definition := range m.Definitions() {
		if err := checkSpan(text, definition.Name, definition.Span); err != nil {
			return fmt.Errorf("definition: %w", err)
		}
		if definition.Span.Offset < end {
			return fmt.Errorf("definition %+v before the one ending at %d", definition, end)
		}
		end = definition.Span.End()
	}

	end = 0
	for _, alias := range m.Aliases() {
		if err := checkSpan(text, alias.Name, alias.Span); err != nil {
			return fmt.Errorf("alias: %w", err)
		}
		if !m.IsAlias(alias.Name) || len(m.DefinitionsNamed(alias.Terminal)) == 0 {
			return fmt.Errorf("alias %+v is no string or has no terminal", alias)
		}
		if alias.Span.Offset < end {
			return fmt.Errorf("alias %+v before the one ending at %d", alias, end)
		}
		end = alias.Span.End()
	}

	end = 0
	for _, reference := range m.References() {
		if err := checkSpan(text, reference.Name, reference.Span); err != nil {
			return fmt.Errorf("reference: %w", err)
		}
		if reference.Span.Offset < end {
			return fmt.Errorf("reference %+v before the one ending at %d", reference, end)
		}
		end = reference.Span.End()
	}
	return nil
}

// checkOccurrences returns an error if the occurrences are not the definitions, aliases and references sorted by
// offset without overlapping.
func checkOccurrences(text []byte, m *model.Model) error {
	occurrences := m.Occurrences()
	if len(occurrences) != len(m.Definitions())+len(m.Aliases())+len(m.References()) {
		return fmt.Errorf(
			"%d occurrences for %d definitions, %d aliases and %d references",
			len(occurrences), len(m.Definitions()), len(m.Aliases()), len(m.References()),
		)
	}
	end := 0
	for _, occurrence := range occurrences {
		if err := checkSpan(text, occurrence.Name, occurrence.Span); err != nil {
			return fmt.Errorf("occurrence: %w", err)
		}
		if occurrence.Span.Offset < end {
			return fmt.Errorf("occurrence %+v overlaps the one ending at %d", occurrence, end)
		}
		end = occurrence.Span.End()
	}
	return nil
}

// checkLookups returns an error if a lookup by name does not hold what the lists hold for that name, or the usages of
// a name are not in the order of the text.
func checkLookups(m *model.Model) error {
	definitionCounts := map[string]int{}
	for _, definition := range m.Definitions() {
		definitionCounts[definition.Name]++
	}
	for name, count := range definitionCounts {
		if len(m.DefinitionsNamed(name)) != count || len(m.DeclarationsNamed(name)) != count {
			return fmt.Errorf("%d definitions named %q, but the lookups find others", count, name)
		}
	}

	aliasCounts := map[string]int{}
	for _, alias := range m.Aliases() {
		aliasCounts[alias.Name]++
	}
	for name, count := range aliasCounts {
		if len(m.AliasesNamed(name)) != count || len(m.DeclarationsNamed(name)) != count {
			return fmt.Errorf("%d aliases named %q, but the lookups find others", count, name)
		}
	}

	referenceCounts := map[string]int{}
	for _, reference := range m.References() {
		referenceCounts[reference.Name]++
	}
	for name, count := range referenceCounts {
		if len(m.ReferencesNamed(name)) != count {
			return fmt.Errorf("%d references named %q, but the lookup finds %d", count, name, len(m.ReferencesNamed(name)))
		}
	}

	for name := range definitionCounts {
		usages := m.UsagesNamed(name)
		for idx := 1; idx < len(usages); idx++ {
			if usages[idx].Span.Offset <= usages[idx-1].Span.Offset {
				return fmt.Errorf("usages of %q out of order: %+v", name, usages)
			}
		}
	}
	return nil
}

// checkOffsetQueries returns an error if the occurrence or the completion at an offset of the text, or just outside
// of it, does not fit the offset.
func checkOffsetQueries(text []byte, m *model.Model) error {
	for offset := -1; offset <= len(text)+1; offset++ {
		if occurrence, found := m.OccurrenceAt(offset); found {
			if offset < occurrence.Span.Offset || occurrence.Span.End() < offset {
				return fmt.Errorf("occurrence %+v at %d does not contain it", occurrence, offset)
			}
		}

		completion := m.CompletionAt(offset)
		caret := min(max(offset, 0), len(text))
		if completion.Offset < 0 || caret < completion.Offset {
			return fmt.Errorf("completion at %d starts its word at %d", offset, completion.Offset)
		}
		if completion.Offset < caret && text[completion.Offset] == '@' && completion.Symbols {
			return fmt.Errorf("completion at %d offers symbols for a word starting with @", offset)
		}
	}
	return nil
}

// checkRenames returns an error if renaming an occurrence to its own name is rejected.
func checkRenames(m *model.Model) error {
	for _, occurrence := range m.Occurrences() {
		if err := m.ValidateRename(occurrence, occurrence.Name); err != nil {
			return fmt.Errorf("renaming %+v to its own name: %w", occurrence, err)
		}
		// The result depends on the text, the call must only return.
		_ = m.ValidateRename(occurrence, occurrence.Name+"x")
	}
	return nil
}
