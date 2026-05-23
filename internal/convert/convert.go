package convert

import (
	"io"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonfrontend "github.com/backbone81/golr/internal/parsergen/frontend/bison"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// BisonToGoLR reads a GNU Bison grammar from reader and writes a GoLR grammar to writer.
func BisonToGoLR(reader io.Reader, writer io.Writer, filePath string) error {
	grammar, err := bisonfrontend.ToGrammar(reader, filePath)
	if err != nil {
		return err
	}
	return golrfrontend.FromGrammar(writer, nil, sanitizeBisonGrammar(grammar))
}

// sanitizeBisonGrammar fixes grammar elements that are valid in GNU Bison but not in GoLR.
func sanitizeBisonGrammar(grammar frontend.Grammar) frontend.Grammar {
	// Bison uses _("...") as translatable display names for terminals (e.g. _("identifier")).
	// These are not valid GoLR scanner patterns — clear them so the output falls back to @empty.
	for i, terminal := range grammar.Terminals {
		if isBisonTranslatableAlias(terminal.Alias) {
			grammar.Terminals[i].Alias = ""
		} else if isBisonCharLiteral(terminal.Alias) {
			grammar.Terminals[i].Alias = bisonCharLiteralToGoLRString(terminal.Alias)
		}
	}

	// Bison allows dots in nonterminal names (e.g. "string.opt", "token_decl.1").
	// GoLR identifiers are restricted to [A-Za-z0-9_] — replace all other runes with '_'.
	for i, nonterminal := range grammar.Nonterminals {
		grammar.Nonterminals[i].Name = sanitizeGoLRName(nonterminal.Name, grammar.Nonterminals)
	}

	return grammar
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

// sanitizeGoLRName replaces characters that are not valid in GoLR identifiers with '_'.
func sanitizeGoLRName(name string, nonterminals []frontend.Symbol) string {
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

	// We need to make sure that our sanitized candidate does not collide with some other nonterminal. We append "_1"
	// as long as we find collisions.
	for {
		if slices.IndexFunc(nonterminals, func(symbol frontend.Symbol) bool {
			return symbol.Name == candidate
		}) == -1 {
			return candidate
		}
		candidate += "_1"
	}
}

// BisonToGoLRFile reads a GNU Bison grammar from inputFilePath and writes a GoLR grammar to outputFilePath.
func BisonToGoLRFile(inputFilePath string, outputFilePath string) error {
	grammar, err := bisonfrontend.GrammarFromFile(inputFilePath)
	if err != nil {
		return err
	}
	return golrfrontend.GrammarToFile(outputFilePath, nil, sanitizeBisonGrammar(grammar))
}

// BisonToGoLRString reads a GNU Bison grammar from bisonGrammar and returns a GoLR grammar as return value.
func BisonToGoLRString(bisonGrammar string) (string, error) {
	grammar, err := bisonfrontend.GrammarFromString(bisonGrammar)
	if err != nil {
		return "", err
	}
	return golrfrontend.GrammarToString(nil, sanitizeBisonGrammar(grammar))
}
