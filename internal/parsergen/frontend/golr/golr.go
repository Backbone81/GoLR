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

	scanner := golrparser.NewTokenSkipper(
		golrparser.NewScanner(data, filePath),
	)

	parser := golrparser.NewParser()
	rootNode, err := parser.Parse(scanner)
	if err != nil {
		return nil, frontend.Grammar{}, err
	}

	walker := NewTreeWalker(scanner)
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
		if terminal.Name == frontend.SymbolError.Name {
			continue
		}
		maxNameLen = max(maxNameLen, len(terminal.Name))
	}

	for _, terminal := range grammar.Terminals {
		if terminal.Name == frontend.SymbolError.Name {
			// The error symbol is not a token a scanner section can declare: its name is reserved and no input
			// produces it. Reading the document back seeds it again.
			continue
		}
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
// A terminal with an alias is written with its alias, which is the string pattern it was read from, because writing
// any other pattern would drop the alias. A terminal without one is written with its scanner rule, and falls back to
// @empty without a rule.
func golrTerminalPattern(terminal frontend.Symbol, ruleByName map[string]frontend2.Rule) (string, bool) {
	rule, hasRule := ruleByName[terminal.Name]
	if terminal.Alias != "" {
		// Alias is stored with surrounding quotes already, e.g. `"+"`.
		return terminal.Alias, hasRule && rule.Skip
	}
	if hasRule {
		return golrNodeToPattern(&rule.Regex), rule.Skip
	}
	return "@empty", false
}

// golrNodeToPattern converts a regex Node to the appropriate GoLR scanner pattern string.
// An empty char class becomes @empty, everything else is wrapped in /.../ using Node.String() which produces valid
// regex syntax. A plain literal is no string pattern either, because a string pattern declares an alias.
func golrNodeToPattern(node *frontend2.Node) string {
	if node.Kind == frontend2.KindCharClass && !node.CharClass.Negate && len(node.CharClass.Ranges) == 0 {
		return "@empty"
	}
	return "/" + node.String() + "/"
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

// writeGoLRRules writes the productions in their order. Only productions which follow each other share a rule, a
// nonterminal whose productions are interleaved with those of others gets a rule for each run of them. The order of
// the productions decides reduce/reduce conflicts, see conflict.EarliestProductionPolicy, so it has to survive.
func writeGoLRRules(writer io.Writer, grammar frontend.Grammar) error {
	var nonterminalGroups []NonterminalGroup

	for _, production := range grammar.Productions {
		if last := len(nonterminalGroups) - 1; last >= 0 &&
			nonterminalGroups[last].NonterminalIdx == production.NonterminalIdx {
			nonterminalGroups[last].Productions = append(nonterminalGroups[last].Productions, production)
		} else {
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
			if production.Name != nil {
				rhs += " @name(" + *production.Name + ")"
			}
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
			if grammar.Terminals[ref.Idx()].Name == frontend.SymbolError.Name {
				parts[i] = "@error"
				continue
			}
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
