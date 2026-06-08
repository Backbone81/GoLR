package bison

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime/trace"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
	bisonparser "github.com/backbone81/golr/internal/parsergen/frontend/bison/parser"
)

// ToGrammar reads the context free grammar as GNU Bison grammar document from the given reader. Returns an error if the
// grammar document can not be parsed successfully.
func ToGrammar(reader io.Reader, filePath string) (frontend.Grammar, error) {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: Bison: ToGrammar").End()

	data, err := io.ReadAll(reader)
	if err != nil {
		return frontend.Grammar{}, err
	}

	scanner := bisonparser.TokenTransformer{
		Scanner: &bisonparser.WhitespaceSkipper{
			Scanner: &bisonparser.ContextScanner{
				Scanner: bisonparser.NewScanner(data, filePath),
			},
		},
	}

	parser := bisonparser.NewParser()
	rootNode, err := parser.Parse(&scanner)
	if err != nil {
		return frontend.Grammar{}, err
	}

	walker := NewASTWalker()
	return walker.BuildGrammar(rootNode)
}

// FromGrammar writes the context free grammar as GNU Bison grammar document to the given writer. Returns an error if
// the grammar document can not be encoded successfully.
func FromGrammar(writer io.Writer, grammar frontend.Grammar) error {
	defer trace.StartRegion(context.TODO(), "GoLR: Parsergen: Frontends: Bison: FromGrammar").End()

	if err := grammarToBisonGrammarFile(writer, grammar); err != nil {
		return fmt.Errorf("encoding grammar to GNU Bison: %w", err)
	}
	return nil
}

type PrecedenceGroup struct {
	Associativity frontend.Associativity
	TerminalIdxs  []int
}

func grammarToBisonGrammarFile(writer io.Writer, grammar frontend.Grammar) error {
	if err := writeBisonGrammarTokens(writer, grammar); err != nil {
		return err
	}
	if err := writeBisonAssociativityAndPrecedence(writer, grammar); err != nil {
		return err
	}
	if err := writeBisonStartSymbol(writer, grammar); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer, "%%"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}

	if err := writeBisonGrammarProductions(writer, grammar); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "  ;"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func writeBisonStartSymbol(writer io.Writer, grammar frontend.Grammar) error {
	if _, err := fmt.Fprintf(
		writer,
		"%%start %s\n",
		grammar.Nonterminals[grammar.StartNonterminalIdx].Name,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func writeBisonGrammarTokens(writer io.Writer, grammar frontend.Grammar) error {
	if _, err := fmt.Fprintf(writer, "%%token\n"); err != nil {
		return err
	}
	for _, symbol := range grammar.Terminals {
		if _, err := fmt.Fprintf(writer, "  %s\n", symbol.Name); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

//nolint:cyclop // Moving code out would make it more difficult to understand.
func writeBisonAssociativityAndPrecedence(writer io.Writer, grammar frontend.Grammar) error {
	precedenceGroups := buildPrecedenceGroups(grammar)
	if len(precedenceGroups) == 0 {
		return nil
	}
	for _, group := range precedenceGroups {
		switch group.Associativity {
		case frontend.AssociativityLeft:
			if _, err := fmt.Fprintf(writer, "%%left"); err != nil {
				return err
			}
		case frontend.AssociativityRight:
			if _, err := fmt.Fprintf(writer, "%%right"); err != nil {
				return err
			}
		case frontend.AssociativityNone:
			if _, err := fmt.Fprintf(writer, "%%nonassoc"); err != nil {
				return err
			}
		case frontend.AssociativityUndeclared:
			// Nothing to output for undeclared associativity as that is the default situation.
		default:
			if _, err := fmt.Fprintf(writer, "%%precedence"); err != nil {
				return err
			}
		}

		for _, idx := range group.TerminalIdxs {
			if _, err := fmt.Fprintf(writer, " %s", grammar.Terminals[idx].Name); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
}

func buildPrecedenceGroups(grammar frontend.Grammar) []*PrecedenceGroup {
	precedenceGroupByLevel := make(map[int]*PrecedenceGroup, 10)
	for idx, terminal := range grammar.Terminals {
		if terminal.Precedence == 0 {
			// Precedence of 0 means no precedence was set.
			continue
		}

		group, ok := precedenceGroupByLevel[terminal.Precedence]
		if !ok {
			group = &PrecedenceGroup{
				Associativity: terminal.Associativity,
			}
			precedenceGroupByLevel[terminal.Precedence] = group
		}
		group.TerminalIdxs = append(group.TerminalIdxs, idx)
	}

	levels := slices.Collect(maps.Keys(precedenceGroupByLevel))
	slices.Sort(levels)

	result := make([]*PrecedenceGroup, 0, len(levels))
	for _, level := range levels {
		result = append(result, precedenceGroupByLevel[level])
	}
	return result
}

func writeBisonGrammarProductions(writer io.Writer, grammar frontend.Grammar) error {
	currNonterminalIdx := -1
	for _, production := range grammar.Productions {
		if currNonterminalIdx != production.NonterminalIdx {
			if err := writeBisonGrammarProductionEnd(writer, currNonterminalIdx); err != nil {
				return err
			}
			if err := writeBisonGrammarProductionStart(writer, grammar, production); err != nil {
				return err
			}
			currNonterminalIdx = production.NonterminalIdx
		} else {
			if err := writeBisonGrammarProductionAlternative(writer); err != nil {
				return err
			}
		}

		if err := writeBisonGrammarProductionRhs(writer, grammar, production); err != nil {
			return err
		}
		if err := writeBisonGrammarProductionPrecedence(writer, grammar, production); err != nil {
			return err
		}
	}
	return nil
}

func writeBisonGrammarProductionStart(
	writer io.Writer,
	grammar frontend.Grammar,
	production frontend.Production,
) error {
	if _, err := fmt.Fprintf(writer, "%s\n", grammar.Nonterminals[production.NonterminalIdx].Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "  :"); err != nil {
		return err
	}
	return nil
}

func writeBisonGrammarProductionAlternative(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "  |"); err != nil {
		return err
	}
	return nil
}

func writeBisonGrammarProductionEnd(writer io.Writer, currNonterminalIdx int) error {
	if currNonterminalIdx != -1 {
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer, "  ;"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}
	return nil
}

func writeBisonGrammarProductionPrecedence(
	writer io.Writer,
	grammar frontend.Grammar,
	production frontend.Production,
) error {
	if production.PrecedenceTerminalIdx != nil {
		precedenceTerminal := grammar.Terminals[*production.PrecedenceTerminalIdx]
		if _, err := fmt.Fprintf(writer, " %%prec %s", precedenceTerminal.Name); err != nil {
			return err
		}
	}
	return nil
}

func writeBisonGrammarProductionRhs(writer io.Writer, grammar frontend.Grammar, production frontend.Production) error {
	if len(production.SymbolRefs) == 0 {
		if _, err := fmt.Fprintf(writer, " %%empty"); err != nil {
			return err
		}
		return nil
	}

	for _, symbolRef := range production.SymbolRefs {
		if symbolRef.IsTerminal() {
			if _, err := fmt.Fprintf(writer, " %s", grammar.Terminals[symbolRef.Idx()].Name); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(writer, " %s", grammar.Nonterminals[symbolRef.Idx()].Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// GrammarFromFile reads the context free grammar as GNU Bison grammar document from the given file path. Returns an
// error if the file can not be read or the grammar document can not be parsed successfully.
//
//nolint:nonamedreturns // Required for defer
func GrammarFromFile(filePath string) (grammar frontend.Grammar, err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Open(filePath)
	if err != nil {
		return frontend.Grammar{}, fmt.Errorf("opening the Bison file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return ToGrammar(file, filePath)
}

// GrammarToFile writes the context free grammar as GNU Bison grammar document to the given file path. Returns an error
// if the file can not be written or the GNU Bison document can not be encoded successfully.
func GrammarToFile(filePath string, grammar frontend.Grammar) (err error) {
	//nolint:gosec // It is the responsibility of the caller to make sure that the path is safe.
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating the GNU Bison file %q: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing file: %w", closeErr))
		}
	}()

	return FromGrammar(file, grammar)
}

// GrammarFromString reads the context free grammar as GNU Bison grammar document from the given string. Returns an
// error if the grammar document can not be parsed successfully.
func GrammarFromString(bisonGrammar string) (frontend.Grammar, error) {
	return ToGrammar(strings.NewReader(bisonGrammar), "in-memory")
}

// GrammarToString returns the context free grammar as GNU Bison grammar document. Returns an error if the GNU Bison
// document can not be encoded successfully.
func GrammarToString(grammar frontend.Grammar) (string, error) {
	var builder strings.Builder
	if err := FromGrammar(&builder, grammar); err != nil {
		return "", err
	}
	return builder.String(), nil
}
