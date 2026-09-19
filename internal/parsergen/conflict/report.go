package conflict

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ReportConfig controls what WriteConflictReport writes.
type ReportConfig struct {
	// Verbose lists every conflict the policy resolved on its own in full, instead of only summarizing them.
	Verbose bool
}

// WriteConflictReport writes a report of the given conflicts to w. A caller which writes a serialized parser to stdout
// should give it os.Stderr, because mixing the report into that output would corrupt it.
//
// Conflicts the policy could not decide are always reported in full: the grammar author has to act on them, and the
// parser cannot be generated while they stand. Conflicts the policy resolved on its own, by shift over reduce or by the
// earliest production, can run into the hundreds for a large grammar, so they are only summarized by default and listed
// in full after the summary when verbose is set. Conflicts decided by precedence declarations are not reported at all.
func WriteConflictReport(w io.Writer, grammar frontend.Grammar, conflicts []Conflict, config ReportConfig) error {
	var builder strings.Builder

	// The summary always comes first, so that it sits at the same place in every report.
	writeResolvedConflictSummary(&builder, conflicts)

	// The unresolved conflicts are always reported in full.
	unresolved := slices.DeleteFunc(slices.Clone(conflicts), func(c Conflict) bool {
		return c.Decision.Kind != DecisionUnresolved
	})
	reports := buildConflictReports(grammar, unresolved)

	// The resolved conflicts are listed in full only when asked for.
	if config.Verbose {
		resolved := slices.DeleteFunc(slices.Clone(conflicts), func(c Conflict) bool {
			return c.Decision.Kind == DecisionUnresolved
		})
		reports = append(reports, buildConflictReports(grammar, resolved)...)
	}

	for _, report := range reports {
		if err := report.Write(&builder, config); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, builder.String())
	return err
}

// ConflictReport is the report of the conflicts of a single state, rendered with the names of the grammar symbols so it
// can be written without the grammar at hand.
type ConflictReport struct {
	// StateIdx is the state index of the conflicted state.
	StateIdx int

	// Entries are the conflicted terminals of the state.
	Entries []ConflictReportEntry
}

// ConflictReportEntry is the report of a single conflicted terminal of a state.
type ConflictReportEntry struct {
	// Terminal is the name of the conflicted terminal.
	Terminal string

	// Kind is the kind of the conflict, see conflictKind.
	Kind string

	// Contributions are the actions the declarations of the grammar left competing for the terminal.
	Contributions []string

	// Decision is what the policy decided about the conflict.
	Decision string
}

// Write writes the report of the state, one block per conflicted terminal.
func (r ConflictReport) Write(w io.Writer, _ ReportConfig) error {
	var builder strings.Builder
	for _, entry := range r.Entries {
		fmt.Fprintf(&builder, "%s in state %d on terminal %s\n", entry.Kind, r.StateIdx, entry.Terminal)
		for _, contribution := range entry.Contributions {
			fmt.Fprintf(&builder, "    %s\n", contribution)
		}
		fmt.Fprintf(&builder, "  %s\n\n", entry.Decision)
	}
	_, err := io.WriteString(w, builder.String())
	return err
}

// buildConflictReports builds one report per state from the conflicts. The conflicts of a state are adjacent, because
// Resolve returns them in state order.
func buildConflictReports(grammar frontend.Grammar, conflicts []Conflict) []ConflictReport {
	var reports []ConflictReport
	for _, c := range conflicts {
		if len(reports) == 0 || reports[len(reports)-1].StateIdx != c.StateIdx {
			reports = append(reports, ConflictReport{StateIdx: c.StateIdx})
		}
		last := &reports[len(reports)-1]
		last.Entries = append(last.Entries, buildConflictReportEntry(grammar, c))
	}
	return reports
}

// buildConflictReportEntry renders a single conflict: the terminal it occurred on, the actions which competed for that
// terminal once precedence and associativity had decided what they could, and what the policy decided about them.
func buildConflictReportEntry(grammar frontend.Grammar, c Conflict) ConflictReportEntry {
	entry := ConflictReportEntry{
		Terminal: grammar.Terminals[c.TerminalIdx].String(),
		Kind:     conflictKind(c),
		Decision: formatDecision(grammar, c.Decision),
	}
	for _, contribution := range c.Undeclared.All() {
		entry.Contributions = append(entry.Contributions, formatContribution(grammar, contribution))
	}
	return entry
}

// writeResolvedConflictSummary writes the summary of the conflicts the policy resolved, one line per kind with
// shift/reduce before reduce/reduce, so that a diff of two reports shows which kind changed.
//
// The conflicts are counted the way GNU Bison counts them, so the numbers can be compared with each other: a terminal
// on which a shift competes with reductions counts as one shift/reduce conflict, and every reduction beyond the first
// on a terminal counts as one reduce/reduce conflict. A terminal can count as both.
func writeResolvedConflictSummary(builder *strings.Builder, conflicts []Conflict) {
	var shiftReduce, reduceReduce int
	for _, c := range conflicts {
		if c.Decision.Kind == DecisionUnresolved {
			continue
		}
		shift, reduces := countContributions(c.Undeclared)
		if shift && reduces > 0 {
			shiftReduce++
		}
		if reduces > 1 {
			reduceReduce += reduces - 1
		}
	}

	if shiftReduce == 0 && reduceReduce == 0 {
		return
	}
	writeConflictCount(builder, shiftReduce, "shift/reduce")
	writeConflictCount(builder, reduceReduce, "reduce/reduce")
	builder.WriteString("\n")
}

// writeConflictCount writes the summary line for the conflicts of one kind, or nothing when there are none.
func writeConflictCount(builder *strings.Builder, count int, kind string) {
	if count == 0 {
		return
	}
	noun := "conflicts"
	if count == 1 {
		noun = "conflict"
	}
	fmt.Fprintf(builder, "%d %s %s resolved\n", count, kind, noun)
}

// conflictKind classifies the conflict as a shift/reduce or a reduce/reduce conflict, which is the wording a grammar
// author expects from a parser generator. The classification is by the actions precedence and associativity left
// competing, because a shift which precedence removed is no part of the conflict the author has to deal with. A
// conflict which mixes a shift with more than one reduction is reported as a shift/reduce conflict, because the
// competing shift is the part the author usually reasons about first.
func conflictKind(c Conflict) string {
	shift, reduces := countContributions(c.Undeclared)
	switch {
	case shift && reduces > 0:
		return "shift/reduce conflict"
	case reduces > 1:
		return "reduce/reduce conflict"
	default:
		// A single contribution is not a conflict, so this only guards against unexpected input.
		return "conflict"
	}
}

// countContributions reports whether the contributions hold a shift, and how many reductions they hold.
func countContributions(contributions ContributionSet) (bool, int) {
	var shift bool
	var reduces int
	for _, contribution := range contributions.All() {
		if contribution.IsShiftAction() {
			shift = true
		} else {
			reduces++
		}
	}
	return shift, reduces
}

// formatDecision renders what the policy decided about a conflict in a way a grammar author can read without knowing
// the internals of the resolution.
func formatDecision(grammar frontend.Grammar, decision Decision) string {
	switch decision.Kind {
	case DecisionDominant:
		return "resolved in favor of " + formatContribution(grammar, decision.Dominant)
	case DecisionError:
		return "resolved by rejecting the terminal, so the parser reports a syntax error on it"
	case DecisionUnresolved:
		var parts []string
		for _, contribution := range decision.Unresolved.All() {
			parts = append(parts, formatContribution(grammar, contribution))
		}
		return "unresolved, still undecided between " + strings.Join(parts, ", ")
	case DecisionUndefined:
		return "no action to decide about"
	}
	return "unknown decision"
}

// formatContribution renders a single competing action of a conflict. A shift is just a shift, and a reduction is
// spelled out with the production it reduces so the author does not have to look the production index up.
func formatContribution(grammar frontend.Grammar, contribution Contribution) string {
	if contribution.IsShiftAction() {
		return "shift"
	}
	return fmt.Sprintf(
		"reduce production %d  (%s)",
		contribution.ProductionIdx(),
		formatProduction(grammar, contribution.ProductionIdx()),
	)
}

// formatProduction renders a production with the names of its symbols instead of their indexes, which is what makes the
// report readable next to the grammar file the author wrote.
func formatProduction(grammar frontend.Grammar, productionIdx int) string {
	production := grammar.Productions[productionIdx]

	var builder strings.Builder
	builder.WriteString(grammar.Nonterminals[production.NonterminalIdx].String())
	builder.WriteString(" ->")
	if len(production.SymbolRefs) == 0 {
		// An empty right hand side reduces on the empty string, which is easy to miss without a hint.
		builder.WriteString(" (empty)")
		return builder.String()
	}
	for _, symbolRef := range production.SymbolRefs {
		builder.WriteString(" ")
		if symbolRef.IsTerminal() {
			builder.WriteString(grammar.Terminals[symbolRef.Idx()].String())
		} else {
			builder.WriteString(grammar.Nonterminals[symbolRef.Idx()].String())
		}
	}
	return builder.String()
}
