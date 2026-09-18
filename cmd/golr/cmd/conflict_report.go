package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
)

// printConflictReport writes a report of the given conflicts to w. It is meant to be given os.Stderr, because stdout can
// carry the serialized parser of a backend, and mixing the report into that output would corrupt it.
//
// Conflicts the policy could not decide are always reported in full: the grammar author has to act on them, and the
// parser cannot be generated while they stand. Conflicts the policy resolved on its own, by shift over reduce or by the
// earliest production, can run into the hundreds for a large grammar, so they are only summarized by default and listed
// in full after the summary when verbose is set. Conflicts decided by precedence declarations are not reported at all.
func printConflictReport(w io.Writer, grammar frontend.Grammar, conflicts []conflict.Conflict, verbose bool) {
	// The summary always comes first, so that it sits at the same place in every report.
	printResolvedConflictSummary(w, conflicts)

	// The unresolved conflicts are always reported in full.
	for _, c := range conflicts {
		if c.Decision.Kind == conflict.DecisionUnresolved {
			printConflictDetail(w, grammar, c)
		}
	}

	// The resolved conflicts are listed in full only when asked for.
	if verbose {
		for _, c := range conflicts {
			if c.Decision.Kind != conflict.DecisionUnresolved {
				printConflictDetail(w, grammar, c)
			}
		}
	}
}

// printResolvedConflictSummary writes the summary of the conflicts the policy resolved, one line per kind with
// shift/reduce before reduce/reduce, so that a diff of two reports shows which kind changed.
//
// The conflicts are counted the way GNU Bison counts them, so the numbers can be compared with each other: a terminal
// on which a shift competes with reductions counts as one shift/reduce conflict, and every reduction beyond the first
// on a terminal counts as one reduce/reduce conflict. A terminal can count as both.
func printResolvedConflictSummary(w io.Writer, conflicts []conflict.Conflict) {
	var shiftReduce, reduceReduce int
	for _, c := range conflicts {
		if c.Decision.Kind == conflict.DecisionUnresolved {
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
	printConflictCount(w, shiftReduce, "shift/reduce")
	printConflictCount(w, reduceReduce, "reduce/reduce")
	fmt.Fprintln(w)
}

// printConflictCount writes the summary line for the conflicts of one kind, or nothing when there are none.
func printConflictCount(w io.Writer, count int, kind string) {
	if count == 0 {
		return
	}
	noun := "conflicts"
	if count == 1 {
		noun = "conflict"
	}
	fmt.Fprintf(w, "%d %s %s resolved\n", count, kind, noun)
}

// printConflictDetail writes the full report of a single conflict: the state and the terminal it occurred on, the
// actions which competed for that terminal once precedence and associativity had decided what they could, and what
// the policy decided about them.
func printConflictDetail(w io.Writer, grammar frontend.Grammar, c conflict.Conflict) {
	fmt.Fprintf(w, "%s in state %d on terminal %s\n",
		conflictKind(c),
		c.StateIdx,
		grammar.Terminals[c.TerminalIdx],
	)
	for _, contribution := range c.Undeclared.All() {
		fmt.Fprintf(w, "    %s\n", formatContribution(grammar, contribution))
	}
	fmt.Fprintf(w, "  %s\n\n", formatDecision(grammar, c.Decision))
}

// conflictKind classifies the conflict as a shift/reduce or a reduce/reduce conflict, which is the wording a grammar
// author expects from a parser generator. The classification is by the actions precedence and associativity left
// competing, because a shift which precedence removed is no part of the conflict the author has to deal with. A
// conflict which mixes a shift with more than one reduction is reported as a shift/reduce conflict, because the
// competing shift is the part the author usually reasons about first.
func conflictKind(c conflict.Conflict) string {
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
func countContributions(contributions conflict.ContributionSet) (bool, int) {
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
func formatDecision(grammar frontend.Grammar, decision conflict.Decision) string {
	switch decision.Kind {
	case conflict.DecisionDominant:
		return "resolved in favor of " + formatContribution(grammar, decision.Dominant)
	case conflict.DecisionError:
		return "resolved by rejecting the terminal, so the parser reports a syntax error on it"
	case conflict.DecisionUnresolved:
		var parts []string
		for _, contribution := range decision.Unresolved.All() {
			parts = append(parts, formatContribution(grammar, contribution))
		}
		return "unresolved, still undecided between " + strings.Join(parts, ", ")
	case conflict.DecisionUndefined:
		return "no action to decide about"
	}
	return "unknown decision"
}

// formatContribution renders a single competing action of a conflict. A shift is just a shift, and a reduction is
// spelled out with the production it reduces so the author does not have to look the production index up.
func formatContribution(grammar frontend.Grammar, contribution conflict.Contribution) string {
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
