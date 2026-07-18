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
// parser cannot be generated while they stand. Conflicts the policy resolved on its own are expected for a grammar which
// leans on precedence declarations, and a large grammar can have hundreds of them, so they are only summarized by
// default and listed in full when verbose is set.
func printConflictReport(w io.Writer, grammar frontend.Grammar, conflicts []conflict.Conflict, verbose bool) {
	// The unresolved conflicts are always reported in full.
	for _, c := range conflicts {
		if c.Decision.Kind == conflict.DecisionUnresolved {
			printConflictDetail(w, grammar, c)
		}
	}

	// The resolved conflicts are listed in full only when asked for, and summarized otherwise.
	if verbose {
		for _, c := range conflicts {
			if c.Decision.Kind != conflict.DecisionUnresolved {
				printConflictDetail(w, grammar, c)
			}
		}
		return
	}
	printResolvedConflictSummary(w, conflicts)
}

// printResolvedConflictSummary writes the one-line summary of the conflicts the policy resolved, broken down by kind, so
// that the grammar author sees how many there were without the full listing drowning out anything else. Nothing is
// written when no conflict was resolved.
func printResolvedConflictSummary(w io.Writer, conflicts []conflict.Conflict) {
	var shiftReduce, reduceReduce, other int
	for _, c := range conflicts {
		if c.Decision.Kind == conflict.DecisionUnresolved {
			continue
		}
		switch conflictKind(c) {
		case "shift/reduce conflict":
			shiftReduce++
		case "reduce/reduce conflict":
			reduceReduce++
		default:
			other++
		}
	}

	total := shiftReduce + reduceReduce + other
	if total == 0 {
		return
	}

	var parts []string
	if shiftReduce > 0 {
		parts = append(parts, fmt.Sprintf("%d shift/reduce", shiftReduce))
	}
	if reduceReduce > 0 {
		parts = append(parts, fmt.Sprintf("%d reduce/reduce", reduceReduce))
	}
	if other > 0 {
		parts = append(parts, fmt.Sprintf("%d other", other))
	}

	noun := "conflicts"
	if total == 1 {
		noun = "conflict"
	}
	fmt.Fprintf(w, "%s %s resolved\n", strings.Join(parts, ", "), noun)
}

// printConflictDetail writes the full report of a single conflict: the state and the terminal it occurred on, the
// actions which competed for that terminal, and what the policy decided about them.
func printConflictDetail(w io.Writer, grammar frontend.Grammar, c conflict.Conflict) {
	fmt.Fprintf(w, "\n%s in state %d on terminal %s\n",
		conflictKind(c),
		c.StateIdx,
		grammar.Terminals[c.TerminalIdx],
	)
	for _, contribution := range c.Contributions.All() {
		fmt.Fprintf(w, "    %s\n", formatContribution(grammar, contribution))
	}
	fmt.Fprintf(w, "  %s\n", formatDecision(grammar, c.Decision))
}

// conflictKind classifies the conflict as a shift/reduce or a reduce/reduce conflict, which is the wording a grammar
// author expects from a parser generator. A conflict which mixes a shift with more than one reduction is reported as a
// shift/reduce conflict, because the competing shift is the part the author usually reasons about first.
func conflictKind(c conflict.Conflict) string {
	var shifts, reduces int
	for _, contribution := range c.Contributions.All() {
		if contribution.IsShiftAction() {
			shifts++
		} else {
			reduces++
		}
	}
	switch {
	case shifts > 0 && reduces > 0:
		return "shift/reduce conflict"
	case reduces > 1:
		return "reduce/reduce conflict"
	default:
		// A single contribution is not a conflict, so this only guards against unexpected input.
		return "conflict"
	}
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
