package convert

import (
	"fmt"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// BisonGrammar2GoLR modifies grammar elements to make GNU Bison elements valid for GoLR. It returns an error for two
// terminals with the same alias, which Bison only warns about and GoLR does not allow.
func BisonGrammar2GoLR(grammar frontend.Grammar) (frontend.Grammar, error) {
	// Terminals and nonterminals share the names of a GoLR grammar file, so a renamed symbol has to avoid both.
	takenNames := make(map[string]bool, len(grammar.Terminals)+len(grammar.Nonterminals))
	for _, symbol := range slices.Concat(grammar.Terminals, grammar.Nonterminals) {
		takenNames[symbol.Name] = true
	}

	terminalNameByAlias := make(map[string]string)
	for i, terminal := range grammar.Terminals {
		if terminal.Name == frontend.SymbolError.Name {
			// The error symbol is no identifier on purpose, the GoLR writer knows it by its name.
			continue
		}

		// Bison uses _("...") as translatable display names for terminals (e.g. _("identifier")).
		// These are not valid GoLR scanner patterns — clear them so the output falls back to @empty.
		if isBisonTranslatableAlias(terminal.Alias) {
			grammar.Terminals[i].Alias = ""
		} else if isBisonCharLiteral(terminal.Alias) {
			grammar.Terminals[i].Alias = bisonCharLiteralToGoLRString(terminal.Alias)
		}

		if isBisonStringLiteral(terminal.Name) && terminal.Alias == "" {
			// A string literal which no %token declares is a token named by the string. GoLR names it by an
			// identifier and keeps the string as its alias, which is what the productions refer to it by.
			grammar.Terminals[i].Alias = terminal.Name
			grammar.Terminals[i].Name = sanitizeGoLRName("STRING_"+terminal.Name[1:len(terminal.Name)-1], takenNames)
		} else {
			grammar.Terminals[i].Name = sanitizeGoLRName(terminal.Name, takenNames)
		}

		// GoLR has no token for the empty string, so an empty alias falls back to @empty like a translatable one.
		if grammar.Terminals[i].Alias == `""` {
			grammar.Terminals[i].Alias = ""
		}

		// The aliases are compared after the conversion, which makes a character literal and a string the same.
		if alias := grammar.Terminals[i].Alias; alias != "" {
			if other, exists := terminalNameByAlias[alias]; exists {
				return frontend.Grammar{}, fmt.Errorf("terminals %s and %s share the alias %s", other, terminal.Name, alias)
			}
			terminalNameByAlias[alias] = terminal.Name
		}
	}

	// Bison allows dots in nonterminal names (e.g. "string.opt", "token_decl.1").
	// GoLR identifiers are restricted to [A-Za-z0-9_] — replace all other runes with '_'.
	for i, nonterminal := range grammar.Nonterminals {
		grammar.Nonterminals[i].Name = sanitizeGoLRName(nonterminal.Name, takenNames)
	}

	return grammar, nil
}

// isBisonStringLiteral reports whether name is a Bison string literal like "==".
func isBisonStringLiteral(name string) bool {
	return len(name) >= 2 && strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`)
}

// isBisonTranslatableAlias reports whether alias is a Bison translatable string alias like _("identifier").
func isBisonTranslatableAlias(alias string) bool {
	return strings.HasPrefix(alias, `_(`) && strings.HasSuffix(alias, `)`)
}

// isBisonCharLiteral reports whether alias is a Bison character literal like '=' or '\n'.
func isBisonCharLiteral(alias string) bool {
	return len(alias) >= 3 && strings.HasPrefix(alias, "'") && strings.HasSuffix(alias, "'")
}

// bisonCharLiteralToGoLRString converts a Bison character literal like '=' to a GoLR string
// pattern like "=".
func bisonCharLiteralToGoLRString(alias string) string {
	inner := alias[1 : len(alias)-1]
	inner = strings.ReplaceAll(inner, `\'`, `'`) // \' is not needed inside "..."
	inner = strings.ReplaceAll(inner, `"`, `\"`) // " must be escaped inside "..."
	return `"` + inner + `"`
}

// sanitizeGoLRName replaces characters that are not valid in GoLR identifiers with '_'. A name which changes gets "_1"
// appended until it is none of the taken names, and becomes a taken name itself.
func sanitizeGoLRName(name string, takenNames map[string]bool) string {
	candidate := strings.Map(func(r rune) rune {
		if ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)
	if candidate == name {
		// No special characters, we can return early.
		return candidate
	}

	for takenNames[candidate] {
		candidate += "_1"
	}
	takenNames[candidate] = true
	return candidate
}
