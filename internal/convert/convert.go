package convert

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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
		} else if isBisonCharLiteral(terminal.Alias) || isBisonStringLiteral(terminal.Alias) {
			grammar.Terminals[i].Alias = bisonLiteralToGoLRString(terminal.Alias)
		}

		if isBisonStringLiteral(terminal.Name) && terminal.Alias == "" {
			// A string literal which no %token declares is a token named by the string. GoLR names it by an
			// identifier and keeps the string as its alias, which is what the productions refer to it by.
			grammar.Terminals[i].Alias = bisonLiteralToGoLRString(terminal.Name)
			// The name is a new one even if the string has no special characters, so it always has to be unique.
			grammar.Terminals[i].Name = uniqueGoLRName(
				strings.Map(goLRIdentifierRune, "STRING_"+terminal.Name[1:len(terminal.Name)-1]),
				takenNames,
			)
		} else {
			grammar.Terminals[i].Name = sanitizeGoLRName(terminal.Name, takenNames)
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

// bisonLiteralToGoLRString converts a Bison character literal like '=' or string literal like "==" to a GoLR string
// literal like "=". GoLR has no token for the empty string and no string literal with a line feed, so it returns "" for
// those, which falls back to @empty like a translatable alias. It does the same for a literal with an invalid escape
// sequence.
func bisonLiteralToGoLRString(literal string) string {
	value, ok := unescapeBisonLiteral(literal[1 : len(literal)-1])
	if !ok || value == "" {
		return ""
	}
	goLRString, err := frontend.QuoteGoLRString(value)
	if err != nil {
		return ""
	}
	return goLRString
}

// unescapeBisonLiteral returns the value of the text between the quotes of a Bison character or string literal, which
// has the escape sequences of C. It reports false for an invalid escape sequence.
//
//nolint:cyclop,funlen // A switch over the escape sequences of C.
func unescapeBisonLiteral(text string) (string, bool) {
	var builder strings.Builder
	for idx := 0; idx < len(text); idx++ {
		if text[idx] != '\\' {
			builder.WriteByte(text[idx])
			continue
		}
		idx++
		if idx == len(text) {
			return "", false
		}
		switch text[idx] {
		case 'a':
			builder.WriteByte('\a')
		case 'b':
			builder.WriteByte('\b')
		case 'f':
			builder.WriteByte('\f')
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case 'v':
			builder.WriteByte('\v')
		case '\\', '"', '\'', '?':
			builder.WriteByte(text[idx])
		case '0', '1', '2', '3', '4', '5', '6', '7':
			// An octal escape is a byte, which is a character only up to 0x7F, and Bison refuses the null byte.
			length := bisonEscapeDigits(text[idx:], 8, 3)
			code, _ := strconv.ParseUint(text[idx:idx+length], 8, 32)
			if code == 0 || code > unicode.MaxASCII {
				return "", false
			}
			builder.WriteByte(byte(code))
			idx += length - 1
		case 'x':
			// A hexadecimal escape is a byte as well, written with as many hexadecimal digits as follow.
			digits := text[idx+1:]
			length := bisonEscapeDigits(digits, 16, len(digits))
			code, err := strconv.ParseUint(digits[:length], 16, 32)
			if err != nil || code == 0 || code > unicode.MaxASCII {
				return "", false
			}
			builder.WriteByte(byte(code))
			idx += length
		case 'u', 'U':
			length := 4
			if text[idx] == 'U' {
				length = 8
			}
			if bisonEscapeDigits(text[idx+1:], 16, length) != length {
				return "", false
			}
			code, _ := strconv.ParseUint(text[idx+1:idx+1+length], 16, 32)
			if code == 0 || !utf8.ValidRune(rune(code)) { //nolint:gosec
				return "", false
			}
			builder.WriteRune(rune(code)) //nolint:gosec
			idx += length
		default:
			return "", false
		}
	}
	return builder.String(), true
}

// bisonEscapeDigits returns how many of the leading bytes of the text, but at most maxLength, are digits of the base.
func bisonEscapeDigits(text string, base int, maxLength int) int {
	length := 0
	for length < len(text) && length < maxLength {
		digit := strings.IndexByte("0123456789abcdef", byte(unicode.ToLower(rune(text[length])))) //nolint:gosec
		if digit < 0 || digit >= base {
			break
		}
		length++
	}
	return length
}

// sanitizeGoLRName replaces characters that are not valid in GoLR identifiers with '_'. A name which changes is made
// unique by uniqueGoLRName.
func sanitizeGoLRName(name string, takenNames map[string]bool) string {
	candidate := strings.Map(goLRIdentifierRune, name)
	if candidate == name {
		// No special characters, we can return early.
		return candidate
	}
	return uniqueGoLRName(candidate, takenNames)
}

// uniqueGoLRName appends "_1" to the name until it is none of the taken names, and makes it a taken name itself.
func uniqueGoLRName(name string, takenNames map[string]bool) string {
	for takenNames[name] {
		name += "_1"
	}
	takenNames[name] = true
	return name
}

// goLRIdentifierRune returns the rune if it is valid in a GoLR identifier and '_' otherwise.
func goLRIdentifierRune(r rune) rune {
	if ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == '_' {
		return r
	}
	return '_'
}
