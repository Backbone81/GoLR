package golr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"sort"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrparser "github.com/backbone81/golr/internal/parsergen/frontend/golr/parser"
	frontend2 "github.com/backbone81/golr/internal/scannergen/frontend"
)

// ToGrammar reads the context free grammar as GoLR grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
func ToGrammar(reader io.Reader, filePath string) ([]frontend2.Rule, frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: GoLR: ToGrammar").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, frontend.Grammar{}, err
	}

	scanner := golrparser.WhitespaceSkipper{
		Scanner: golrparser.NewScanner(data, filePath),
	}

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return nil, frontend.Grammar{}, err
	}

	walker := NewASTWalker()
	rules, grammar, err := walker.BuildGrammar(rootNode)
	if err != nil {
		return nil, frontend.Grammar{}, err
	}
	return rules, grammar, nil
}

// FromGrammar writes the context free grammar as GoLR grammar document to the given writer. Returns an error if
// the grammar document can not be encoded successfully.
func FromGrammar(writer io.Writer, rules []frontend2.Rule, grammar frontend.Grammar) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: GoLR: FromGrammar").End()

	if err := grammarToGoLRGrammarFile(writer, rules, grammar); err != nil {
		return fmt.Errorf("encoding grammar to GoLR: %w", err)
	}
	return nil
}

func grammarToGoLRGrammarFile(writer io.Writer, rules []frontend2.Rule, grammar frontend.Grammar) error {
	if _, err := fmt.Fprintln(writer, "@scanner {"); err != nil {
		return err
	}
	if err := writeGoLRScannerSection(writer, rules, grammar); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "}"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "@parser {"); err != nil {
		return err
	}
	if err := writeGoLRParserSection(writer, grammar); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "}"); err != nil {
		return err
	}
	return nil
}

func writeGoLRScannerSection(writer io.Writer, rules []frontend2.Rule, grammar frontend.Grammar) error {
	ruleByName := make(map[string]frontend2.Rule, len(rules))
	for _, rule := range rules {
		ruleByName[rule.Name] = rule
	}

	var maxNameLen int
	for _, terminal := range grammar.Terminals {
		maxNameLen = max(maxNameLen, len(terminal.Name))
	}

	for _, terminal := range grammar.Terminals {
		if err := writeGoLRTerminalDecl(writer, terminal, ruleByName, maxNameLen); err != nil {
			return err
		}
	}
	return nil
}

func writeGoLRTerminalDecl(
	writer io.Writer,
	terminal frontend.Symbol,
	ruleByName map[string]frontend2.Rule,
	maxNameLen int,
) error {
	pattern, skip := golrTerminalPattern(terminal, ruleByName)
	padding := strings.Repeat(" ", maxNameLen-len(terminal.Name))
	if _, err := fmt.Fprintf(writer, "    %s: %s%s", terminal.Name, padding, pattern); err != nil {
		return err
	}
	if skip {
		if _, err := fmt.Fprintf(writer, " @skip"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(writer, ";\n"); err != nil {
		return err
	}
	return nil
}

// golrTerminalPattern returns the GoLR pattern string and skip flag for a terminal.
// If a scanner rule exists for the terminal it takes precedence. Otherwise the heuristic is applied:
// a string alias is emitted as a string literal, anything else falls back to @empty.
func golrTerminalPattern(terminal frontend.Symbol, ruleByName map[string]frontend2.Rule) (string, bool) {
	if rule, ok := ruleByName[terminal.Name]; ok {
		return golrNodeToPattern(&rule.Regex), rule.Skip
	}
	if terminal.Alias != "" {
		// Alias is stored with surrounding quotes already, e.g. `"+"`.
		return terminal.Alias, false
	}
	return "@empty", false
}

// golrNodeToPattern converts a regex Node to the appropriate GoLR scanner pattern string.
// An empty char class becomes @empty, a plain literal becomes a double-quoted string, everything
// else is wrapped in /.../ using Node.String() which produces valid regex syntax.
func golrNodeToPattern(node *frontend2.Node) string {
	if node.Kind == frontend2.KindCharClass && !node.CharClass.Negate && len(node.CharClass.Ranges) == 0 {
		return "@empty"
	}
	if node.Kind == frontend2.KindLiteral {
		return golrLiteralToStringPattern(node.Literal.Text)
	}
	return "/" + node.String() + "/"
}

// golrLiteralToStringPattern formats literal text as a double-quoted GoLR string pattern,
// escaping " and \ which are the only characters that need escaping in this context.
func golrLiteralToStringPattern(text string) string {
	var b strings.Builder
	b.WriteRune('"')
	for _, r := range text {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteRune('"')
	return b.String()
}

func writeGoLRParserSection(writer io.Writer, grammar frontend.Grammar) error {
	if err := writeGoLRStart(writer, grammar); err != nil {
		return err
	}
	if err := writeGoLRPrecedenceSection(writer, grammar); err != nil {
		return err
	}
	if err := writeGoLRRules(writer, grammar); err != nil {
		return err
	}
	return nil
}

func writeGoLRStart(writer io.Writer, grammar frontend.Grammar) error {
	if grammar.StartNonterminalIdx == grammar.Productions[0].NonterminalIdx {
		return nil
	}
	if _, err := fmt.Fprintf(
		writer,
		"    @start: %s;\n",
		grammar.Nonterminals[grammar.StartNonterminalIdx].Name,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

type precedenceGroup struct {
	associativity frontend.Associativity
	names         []string
}

func writeGoLRPrecedenceSection(writer io.Writer, grammar frontend.Grammar) error {
	groupByPrec := make(map[int]*precedenceGroup)
	var precValues []int

	for _, terminal := range grammar.Terminals {
		if terminal.Precedence == 0 && terminal.Associativity == frontend.AssociativityUndeclared {
			continue
		}
		if _, exists := groupByPrec[terminal.Precedence]; !exists {
			groupByPrec[terminal.Precedence] = &precedenceGroup{
				associativity: terminal.Associativity,
			}
			precValues = append(precValues, terminal.Precedence)
		}
		groupByPrec[terminal.Precedence].names = append(groupByPrec[terminal.Precedence].names, terminal.String())
	}

	if len(precValues) == 0 {
		return nil
	}

	// Sort descending: higher value = higher precedence = declared first
	sort.Slice(precValues, func(i, j int) bool {
		return precValues[i] > precValues[j]
	})

	if _, err := fmt.Fprintln(writer, "    @precedence {"); err != nil {
		return err
	}
	for _, prec := range precValues {
		group := groupByPrec[prec]
		if _, err := fmt.Fprintf(
			writer,
			"        %s: %s;\n",
			golrAssociativityKeyword(group.associativity),
			strings.Join(group.names, " "),
		); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "    }"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func golrAssociativityKeyword(a frontend.Associativity) string {
	//nolint:exhaustive // We do not have to be exhaustive here.
	switch a {
	case frontend.AssociativityLeft:
		return "@left"
	case frontend.AssociativityRight:
		return "@right"
	case frontend.AssociativityNone:
		return "@none"
	default: // AssociativityUndeclared
		return "@precedence"
	}
}

type NonterminalGroup struct {
	NonterminalIdx int
	Productions    []frontend.Production
}

func writeGoLRRules(writer io.Writer, grammar frontend.Grammar) error {
	var nonterminalGroups []NonterminalGroup
	groupIdxByNonterminalIdx := make(map[int]int)

	for _, production := range grammar.Productions {
		if nonterminalGroupIdx, ok := groupIdxByNonterminalIdx[production.NonterminalIdx]; ok {
			nonterminalGroups[nonterminalGroupIdx].Productions = append(
				nonterminalGroups[nonterminalGroupIdx].Productions,
				production,
			)
		} else {
			groupIdxByNonterminalIdx[production.NonterminalIdx] = len(nonterminalGroups)
			nonterminalGroups = append(nonterminalGroups, NonterminalGroup{
				NonterminalIdx: production.NonterminalIdx,
				Productions:    []frontend.Production{production},
			})
		}
	}

	for nonterminalGroupIdx, nonterminalGroup := range nonterminalGroups {
		if nonterminalGroupIdx > 0 {
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "    %s\n", grammar.Nonterminals[nonterminalGroup.NonterminalIdx].Name); err != nil {
			return err
		}
		for productionIdx, production := range nonterminalGroup.Productions {
			sep := ":"
			if productionIdx > 0 {
				sep = "|"
			}
			rhs := golrProductionRHS(production, grammar)
			if production.PrecedenceTerminalIdx != nil {
				rhs += " @precedence(" + grammar.Terminals[*production.PrecedenceTerminalIdx].String() + ")"
			}
			if _, err := fmt.Fprintf(writer, "        %s %s\n", sep, rhs); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, "        ;"); err != nil {
			return err
		}
	}
	return nil
}

func golrProductionRHS(prod frontend.Production, grammar frontend.Grammar) string {
	if len(prod.SymbolRefs) == 0 {
		return "@empty"
	}
	parts := make([]string, len(prod.SymbolRefs))
	for i, ref := range prod.SymbolRefs {
		if ref.IsTerminal() {
			parts[i] = grammar.Terminals[ref.Idx()].String()
		} else {
			parts[i] = grammar.Nonterminals[ref.Idx()].Name
		}
	}
	return strings.Join(parts, " ")
}

// GrammarFromFile reads the context free grammar as GoLR grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
//
//nolint:nonamedreturns // Required for defer
func GrammarFromFile(filePath string) (rules []frontend2.Rule, grammar frontend.Grammar, err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, frontend.Grammar{}, fmt.Errorf("opening the GoLR file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToGrammar(file, filePath)
}

// GrammarToFile writes the context free grammar as GoLR grammar document to the given file path. Returns an error
// if the file can not be written or the GoLR document can not be encoded successfully.
func GrammarToFile(filePath string, rules []frontend2.Rule, grammar frontend.Grammar) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the GoLR file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromGrammar(file, rules, grammar)
}

// GrammarFromString reads the context free grammar as GoLR grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
func GrammarFromString(golrGrammar string) ([]frontend2.Rule, frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(golrGrammar), "in-memory")
}

// GrammarToString returns the context free grammar as GoLR grammar document. Returns an error if the GoLR
// document can not be encoded successfully.
func GrammarToString(rules []frontend2.Rule, grammar frontend.Grammar) (string, error) {
	var builder strings.Builder
	if err := FromGrammar(&builder, rules, grammar); err != nil {
		return "", err
	}
	return builder.String(), nil
}
