package model

import (
	"slices"
	"strings"
)

// Model is the symbol model of a text. It is immutable, and the slices it returns must not be modified.
type Model struct {
	definitions []Definition
	aliases     []Alias
	references  []Reference

	// occurrences holds all definitions, aliases and references sorted by offset. Their spans do not overlap.
	occurrences []Occurrence

	definitionsByName map[string][]Definition
	aliasesByName     map[string][]Alias
	referencesByName  map[string][]Reference

	// aliasNamesByTerminal holds the names of the aliases of a terminal without duplicates, as a terminal declared
	// twice with the same string has the alias twice.
	aliasNamesByTerminal map[string][]string
}

// index builds the lookups by name from the lists.
func (m *Model) index() {
	m.definitionsByName = make(map[string][]Definition, len(m.definitions))
	for _, definition := range m.definitions {
		m.definitionsByName[definition.Name] = append(m.definitionsByName[definition.Name], definition)
	}

	m.aliasesByName = make(map[string][]Alias, len(m.aliases))
	m.aliasNamesByTerminal = make(map[string][]string, len(m.aliases))
	for _, alias := range m.aliases {
		m.aliasesByName[alias.Name] = append(m.aliasesByName[alias.Name], alias)
		if !slices.Contains(m.aliasNamesByTerminal[alias.Terminal], alias.Name) {
			m.aliasNamesByTerminal[alias.Terminal] = append(m.aliasNamesByTerminal[alias.Terminal], alias.Name)
		}
	}

	m.referencesByName = map[string][]Reference{}
	for _, reference := range m.references {
		m.referencesByName[reference.Name] = append(m.referencesByName[reference.Name], reference)
	}
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
	return m.definitionsByName[name]
}

// AliasesNamed returns the aliases with the name.
func (m *Model) AliasesNamed(name string) []Alias {
	return m.aliasesByName[name]
}

// DeclarationsNamed returns what a reference with the name resolves to: the aliases for a string, the definitions
// otherwise. A name declared twice resolves to both declarations.
func (m *Model) DeclarationsNamed(name string) []Occurrence {
	if m.IsAlias(name) {
		aliases := m.aliasesByName[name]
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

	definitions := m.definitionsByName[name]
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
	return m.referencesByName[name]
}

// UsagesNamed returns the references of the symbol with the name in the order of the text: those by the name and,
// for a terminal, those by its aliases. An alias has only the references by its string.
func (m *Model) UsagesNamed(name string) []Reference {
	aliasNames := m.aliasNamesByTerminal[name]
	if len(aliasNames) == 0 {
		return m.referencesByName[name]
	}

	usages := slices.Clone(m.referencesByName[name])
	for _, aliasName := range aliasNames {
		usages = append(usages, m.referencesByName[aliasName]...)
	}
	slices.SortFunc(usages, Reference.compareOffset)
	return usages
}
