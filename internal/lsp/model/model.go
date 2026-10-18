package model

import (
	"fmt"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
)

// Model is the symbol model of a text. It is immutable, and the slices it returns must not be modified.
type Model struct {
	text []byte

	definitions []Definition
	aliases     []Alias
	references  []Reference

	// occurrences holds all definitions, aliases and references sorted by offset. Their spans do not overlap.
	occurrences []Occurrence

	// names holds the distinct names by their name ID, nameIDs the name ID of every name.
	names   []string
	nameIDs map[string]nameID

	// definitionsByID, aliasesByID and referencesByID hold the definitions, aliases and references of every name ID.
	definitionsByID [][]Definition
	aliasesByID     [][]Alias
	referencesByID  [][]Reference

	// aliasNamesByTerminal holds the names of the aliases of a terminal without duplicates, as a terminal declared
	// twice with the same string has the alias twice.
	aliasNamesByTerminal map[string][]string

	// completions holds the completion contexts sorted by offset, the first at offset zero. The offsets are distinct.
	completions []CompletionContext
}

// nameID identifies a distinct name of the text, numbered from zero in the order of the first occurrence.
type nameID int32

// index builds the lookups by name from the lists and the name IDs of their elements.
func (m *Model) index(definitionIDs []nameID, aliasIDs []nameID, referenceIDs []nameID) {
	m.definitionsByID = groupByID(m.definitions, definitionIDs, len(m.names))
	m.aliasesByID = groupByID(m.aliases, aliasIDs, len(m.names))
	m.referencesByID = groupByID(m.references, referenceIDs, len(m.names))

	m.aliasNamesByTerminal = make(map[string][]string, len(m.aliases))
	for _, alias := range m.aliases {
		if !slices.Contains(m.aliasNamesByTerminal[alias.Terminal], alias.Name) {
			m.aliasNamesByTerminal[alias.Terminal] = append(m.aliasNamesByTerminal[alias.Terminal], alias.Name)
		}
	}
}

// groupByID groups the values by their name IDs, keeping their order. The group of a name ID without values is nil.
// All groups share one array.
//
// Appending to a slice per name allocates every time a group outgrows its capacity. For a big grammar this is
// thousands of allocations, as most names have only one definition and a few references. Counting the name IDs first
// gives every group its exact size, so the groups are carved out of one array and never grow. Indexing by name ID
// needs no map.
func groupByID[T any](values []T, ids []nameID, nameCount int) [][]T {
	counts := make([]int, nameCount)
	for _, id := range ids {
		counts[id]++
	}
	backing := make([]T, len(values))
	groups := make([][]T, nameCount)
	start := 0
	for id, count := range counts {
		if count == 0 {
			continue
		}
		// The capacity ends at the group, so appending to it cannot overwrite the next group.
		groups[id] = backing[start : start : start+count]
		start += count
	}
	for i, value := range values {
		groups[ids[i]] = append(groups[ids[i]], value)
	}
	return groups
}

// IsAlias reports whether the name is a string, which references a terminal by its alias.
func (m *Model) IsAlias(name string) bool {
	return strings.HasPrefix(name, `"`)
}

// Definitions returns the definitions in the order of the text.
func (m *Model) Definitions() []Definition {
	return m.definitions
}

// Aliases returns the aliases in the order of the text.
func (m *Model) Aliases() []Alias {
	return m.aliases
}

// References returns the references in the order of the text.
func (m *Model) References() []Reference {
	return m.references
}

// Occurrences returns the definitions, aliases and references in the order of the text.
func (m *Model) Occurrences() []Occurrence {
	return m.occurrences
}

// OccurrenceAt returns the occurrence which contains the offset. An offset at the end of an occurrence counts as
// inside it, so a caret right after a name finds the name, unless another occurrence starts there.
func (m *Model) OccurrenceAt(offset int) (Occurrence, bool) {
	// The spans do not overlap, so their ends are sorted as well as their starts.
	idx, _ := slices.BinarySearchFunc(m.occurrences, offset, Occurrence.compareEnd)
	if idx < len(m.occurrences) && m.occurrences[idx].Span.End() == offset {
		// The occurrence ends at the offset. One starting there wins over it.
		if idx+1 < len(m.occurrences) && m.occurrences[idx+1].Span.Offset == offset {
			return m.occurrences[idx+1], true
		}
		return m.occurrences[idx], true
	}
	if idx < len(m.occurrences) && m.occurrences[idx].Span.Offset <= offset {
		return m.occurrences[idx], true
	}
	return Occurrence{}, false
}

// DefinitionsNamed returns the definitions with the name.
func (m *Model) DefinitionsNamed(name string) []Definition {
	id, ok := m.nameIDs[name]
	if !ok {
		return nil
	}
	return m.definitionsByID[id]
}

// AliasesNamed returns the aliases with the name.
func (m *Model) AliasesNamed(name string) []Alias {
	id, ok := m.nameIDs[name]
	if !ok {
		return nil
	}
	return m.aliasesByID[id]
}

// DeclarationsNamed returns what a reference with the name resolves to: the aliases for a string, the definitions
// otherwise. A name declared twice resolves to both declarations.
func (m *Model) DeclarationsNamed(name string) []Occurrence {
	if m.IsAlias(name) {
		aliases := m.AliasesNamed(name)
		declarations := make([]Occurrence, 0, len(aliases))
		for _, alias := range aliases {
			declarations = append(declarations, Occurrence{
				Name: alias.Name,
				Role: RoleAlias,
				Span: alias.Span,
			})
		}
		return declarations
	}

	definitions := m.DefinitionsNamed(name)
	declarations := make([]Occurrence, 0, len(definitions))
	for _, definition := range definitions {
		declarations = append(declarations, Occurrence{
			Name: definition.Name,
			Role: RoleDefinition,
			Span: definition.Span,
		})
	}
	return declarations
}

// ReferencesNamed returns the references with the name.
func (m *Model) ReferencesNamed(name string) []Reference {
	id, ok := m.nameIDs[name]
	if !ok {
		return nil
	}
	return m.referencesByID[id]
}

// UsagesNamed returns the references of the symbol with the name in the order of the text: those by the name and,
// for a terminal, those by its aliases. An alias has only the references by its string.
func (m *Model) UsagesNamed(name string) []Reference {
	aliasNames := m.aliasNamesByTerminal[name]
	if len(aliasNames) == 0 {
		return m.ReferencesNamed(name)
	}

	usages := slices.Clone(m.ReferencesNamed(name))
	for _, aliasName := range aliasNames {
		usages = append(usages, m.ReferencesNamed(aliasName)...)
	}
	slices.SortFunc(usages, Reference.compareOffset)
	return usages
}

// ValidateRename returns why renaming the occurrence to the new name is not possible, or nil if it is. The new name
// of a symbol has to be exactly one IDENTIFIER and the new name of an alias exactly one STRING, so a rename keeps a
// terminal and its alias apart. A new name which is defined or referenced already is rejected, as it would merge
// symbols. Keeping the old name is fine.
func (m *Model) ValidateRename(occurrence Occurrence, newName string) error {
	if newName == occurrence.Name {
		return nil
	}

	if m.IsAlias(occurrence.Name) {
		if err := m.validateAliasRename(newName); err != nil {
			return err
		}
	} else if err := m.validateSymbolRename(newName); err != nil {
		return err
	}
	if len(m.ReferencesNamed(newName)) > 0 {
		return fmt.Errorf("'%s' is %w", newName, ErrNameReferenced)
	}
	return nil
}

// validateAliasRename returns why renaming an alias to the new name is not possible: it is not exactly one STRING or
// is defined already.
func (m *Model) validateAliasRename(newName string) error {
	if !m.isSingleToken(newName, parser.TokenString) {
		return fmt.Errorf("'%s' is %w", newName, ErrNotString)
	}
	if len(m.AliasesNamed(newName)) > 0 {
		return fmt.Errorf("'%s' is %w", newName, ErrNameDefined)
	}
	return nil
}

// validateSymbolRename returns why renaming a symbol to the new name is not possible: it is not exactly one
// IDENTIFIER or is defined already.
func (m *Model) validateSymbolRename(newName string) error {
	if !m.isSingleToken(newName, parser.TokenIdentifier) {
		return fmt.Errorf("'%s' is %w", newName, ErrNotIdentifier)
	}
	if len(m.DefinitionsNamed(newName)) > 0 {
		return fmt.Errorf("'%s' is %w", newName, ErrNameDefined)
	}
	return nil
}

// isSingleToken reports whether the golr scanner reads the text as exactly one token of the type.
func (m *Model) isSingleToken(text string, token parser.Token) bool {
	scanner := parser.NewScanner([]byte(text), "")
	return scanner.Next() && scanner.Token() == token && !scanner.Next()
}

// CompletionAt returns what completion offers at the caret. The place is the one in front of the word being
// completed, which is the identifier characters before the caret with a leading "@".
func (m *Model) CompletionAt(offset int) Completion {
	offset = min(max(offset, 0), len(m.text))
	start := offset
	for start > 0 && m.isIdentifierByte(m.text[start-1]) {
		start--
	}
	isKeyword := start > 0 && m.text[start-1] == '@'
	if isKeyword {
		start--
	}

	// The first context is at offset zero, so there is always one at or before the start.
	idx, found := slices.BinarySearchFunc(m.completions, start, CompletionContext.compareOffset)
	if !found {
		idx--
	}
	kind := m.completions[idx].Kind
	return Completion{
		Offset:   start,
		Keywords: kind.Keywords(),
		Symbols:  kind.Symbols() && !isKeyword,
	}
}

// isIdentifierByte reports whether the byte can be part of an IDENTIFIER of golr.golr.
func (m *Model) isIdentifierByte(value byte) bool {
	return value == '_' || ('a' <= value && value <= 'z') || ('A' <= value && value <= 'Z') ||
		('0' <= value && value <= '9')
}
