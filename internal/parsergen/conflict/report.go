package conflict

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/frontend"
)

// ReportConfig controls what WriteConflictReport writes.
type ReportConfig struct {
	// Verbose lists every conflict the policy resolved on its own in full, instead of only summarizing them.
	Verbose bool

	// WithStateNumbers adds the state index to every state. It changes with unrelated grammar edits, so it is only a
	// handle to look the state up in a dump of the parser tables from the same run.
	WithStateNumbers bool
}

// WriteConflictReport writes a report of the given conflicts to w. A caller which writes a serialized parser to stdout
// should give it os.Stderr, because mixing the report into that output would corrupt it.
//
// Only the conflicts the policy resolved on its own, by shift over reduce or by the earliest production, are reported.
// They can run into the hundreds for a large grammar, so they are only summarized by default and listed in full after
// the summary when verbose is set. Conflicts decided by precedence declarations are not reported at all, and conflicts
// the policy could not decide are reported by the UnresolvedConflictError of each.
func WriteConflictReport(w io.Writer, grammar frontend.Grammar, conflicts []Conflict, config ReportConfig) error {
	var builder strings.Builder

	resolved := slices.DeleteFunc(slices.Clone(conflicts), func(c Conflict) bool {
		return c.Decision.Kind == DecisionUnresolved
	})

	// The summary always comes first, so that it sits at the same place in every report.
	writeResolvedConflictSummary(&builder, resolved)

	// The resolved conflicts are listed in full only when asked for.
	if !config.Verbose {
		_, err := io.WriteString(w, builder.String())
		return err
	}

	reports := buildConflictReports(grammar, resolved)
	keepDistinguishingLookaheads(reports)
	// The reports are sorted by their content instead of by the state index, so the report does not change when an
	// unrelated grammar edit renumbers the states.
	slices.SortFunc(reports, compareConflictReports)

	for _, report := range reports {
		// The summary and the reports of the states are separated by an empty line.
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
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

	// KernelItems are the kernel items of the conflicted state with a dot at their position, which name the state
	// independently of its index.
	KernelItems []string

	// Reductions tell the state apart from the other states of the same report which have the same kernel items, as
	// IELR(1) splits states by their lookaheads. It is empty unless there are such states, see
	// keepDistinguishingLookaheads.
	Reductions []ConflictReportReduction

	// Entries are the conflicted terminals of the state.
	Entries []ConflictReportEntry
}

// ConflictReportReduction is a reduction of a conflicted state together with the lookaheads which tell the state apart
// from the other states with the same kernel items.
type ConflictReportReduction struct {
	// Item is the completed item of the reduction, with the dot at its end. It is a kernel item, or an item of an empty
	// production which the closure of the state added.
	Item string

	// Lookaheads are the names of the terminals the reduction has in this state, before any conflict was resolved, but
	// only those which not every state with the same kernel items reduces on. It is empty when this state reduces on no
	// terminal the others do not reduce on as well.
	Lookaheads []string
}

// ConflictReportEntry is the report of a single conflicted terminal of a state.
type ConflictReportEntry struct {
	// Terminal is the name of the conflicted terminal.
	Terminal string

	// Kind is the kind of the conflict, see conflictKind.
	Kind string

	// Contributions are the actions the declarations of the grammar left competing for the terminal.
	Contributions []string

	// Chosen is the action which won the conflict, marked among the contributions. It is empty when no single action
	// won.
	Chosen string

	// Decision is what the policy decided about a conflict no single action won. It is empty when an action was chosen.
	Decision string

	// DecisionContributions are the actions the decision is about, which are those an unresolved conflict was left with.
	// It is empty for a decision which is not about a particular action.
	DecisionContributions []string
}

// Write writes the report of the state: the kernel items which name the state, followed by one indented block per
// conflicted terminal, all separated by an empty line. The action which won a conflict is marked in place, so the
// decision does not repeat it. The report ends with a single newline, so a caller which writes
// several reports separates them by an empty line of its own.
func (r ConflictReport) Write(w io.Writer, config ReportConfig) error {
	var builder strings.Builder
	if config.WithStateNumbers {
		fmt.Fprintf(&builder, "state %d:\n", r.StateIdx)
	} else {
		builder.WriteString("state:\n")
	}
	for _, kernelItem := range r.KernelItems {
		fmt.Fprintf(&builder, "  %s\n", kernelItem)
	}
	if len(r.Reductions) > 0 {
		builder.WriteString("  distinguished by lookaheads:\n")
		for _, reduction := range r.Reductions {
			fmt.Fprintf(&builder, "    %s  {%s}\n", reduction.Item, strings.Join(reduction.Lookaheads, ", "))
		}
	}
	for _, entry := range r.Entries {
		fmt.Fprintf(&builder, "\n  %s on terminal %s:\n", entry.Kind, entry.Terminal)
		for _, contribution := range entry.Contributions {
			if contribution == entry.Chosen {
				fmt.Fprintf(&builder, "    %s%s\n", contribution, chosenMarker)
				continue
			}
			fmt.Fprintf(&builder, "    %s\n", contribution)
		}
		if entry.Decision == "" {
			continue
		}
		if slices.Equal(entry.DecisionContributions, entry.Contributions) {
			// A decision which leaves every competing action standing is an unresolved conflict nothing was narrowed
			// down in, which the actions above already say.
			continue
		}
		fmt.Fprintf(&builder, "    %s\n", entry.Decision)
		for _, contribution := range entry.DecisionContributions {
			fmt.Fprintf(&builder, "      %s\n", contribution)
		}
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
			reports = append(reports, buildConflictReport(grammar, c))
		}
		last := &reports[len(reports)-1]
		last.Entries = append(last.Entries, buildConflictReportEntry(grammar, c))
	}
	for _, report := range reports {
		slices.SortFunc(report.Entries, func(a ConflictReportEntry, b ConflictReportEntry) int {
			return strings.Compare(a.Terminal, b.Terminal)
		})
	}
	return reports
}

// compareConflictReports orders the reports lexically by their kernel items, which puts them in the order of the first
// line of every state, then lexically by the reductions, then by the terminals of the entries. The state index is the
// final tie-breaker, so the order is total even when the state numbers are written.
func compareConflictReports(a ConflictReport, b ConflictReport) int {
	if result := slices.Compare(a.KernelItems, b.KernelItems); result != 0 {
		return result
	}
	if result := slices.CompareFunc(a.Reductions, b.Reductions, compareReductions); result != 0 {
		return result
	}
	result := slices.CompareFunc(a.Entries, b.Entries, func(a ConflictReportEntry, b ConflictReportEntry) int {
		return strings.Compare(a.Terminal, b.Terminal)
	})
	if result != 0 {
		return result
	}
	return cmp.Compare(a.StateIdx, b.StateIdx)
}

// compareReductions orders two reductions lexically by the item, then by the lookaheads.
func compareReductions(a ConflictReportReduction, b ConflictReportReduction) int {
	if result := strings.Compare(a.Item, b.Item); result != 0 {
		return result
	}
	return slices.Compare(a.Lookaheads, b.Lookaheads)
}

// keepDistinguishingLookaheads reduces the reductions of every report to what tells its state apart from the other
// states of the reports with the same kernel items. Lookahead sets on real grammars run to dozens of terminals and
// change with unrelated grammar edits, so only the difference is written:
//
//   - A state whose kernel items no other state shares is already told apart by them and keeps no reductions.
//   - A reduction every state of the group has loses the lookaheads all of them share. It is dropped when that leaves
//     nothing in any of the states, and otherwise kept in every state, even with an empty set, so that a missing line
//     never reads as a missing reduction.
//   - A reduction only some states of the group have keeps its full set, because having it at all tells them apart.
//
// Several reports can belong to the same state, which does not make their kernel items shared.
func keepDistinguishingLookaheads(reports []ConflictReport) {
	reductionsByStateIdxByKernel := make(map[string]map[int][]ConflictReportReduction)
	for _, report := range reports {
		kernel := strings.Join(report.KernelItems, "\n")
		if reductionsByStateIdxByKernel[kernel] == nil {
			reductionsByStateIdxByKernel[kernel] = make(map[int][]ConflictReportReduction)
		}
		reductionsByStateIdxByKernel[kernel][report.StateIdx] = report.Reductions
	}

	distinguishingByStateIdx := make(map[int][]ConflictReportReduction)
	for _, reductionsByStateIdx := range reductionsByStateIdxByKernel {
		if len(reductionsByStateIdx) < 2 {
			continue
		}
		maps.Copy(distinguishingByStateIdx, distinguishingReductions(reductionsByStateIdx))
	}

	for i := range reports {
		reports[i].Reductions = distinguishingByStateIdx[reports[i].StateIdx]
	}
}

// distinguishingReductions returns the reductions of every state of a group with the same kernel items, reduced to the
// lookaheads which tell the states apart, see keepDistinguishingLookaheads.
func distinguishingReductions(
	reductionsByStateIdx map[int][]ConflictReportReduction,
) map[int][]ConflictReportReduction {
	// The lookaheads every state of the group shares for a reduction, present only for reductions every state has.
	shared := make(map[string][]string)
	stateCountByItem := make(map[string]int)
	for _, reductions := range reductionsByStateIdx {
		for _, reduction := range reductions {
			stateCountByItem[reduction.Item]++
		}
	}
	for _, reductions := range reductionsByStateIdx {
		for _, reduction := range reductions {
			if stateCountByItem[reduction.Item] != len(reductionsByStateIdx) {
				continue
			}
			if lookaheads, found := shared[reduction.Item]; found {
				shared[reduction.Item] = intersectSorted(lookaheads, reduction.Lookaheads)
			} else {
				shared[reduction.Item] = reduction.Lookaheads
			}
		}
	}

	// A reduction is dropped when it has the same lookaheads in every state, which means nothing is left of it anywhere.
	distinguishing := make(map[string]bool)
	for _, reductions := range reductionsByStateIdx {
		for _, reduction := range reductions {
			lookaheads, found := shared[reduction.Item]
			if !found || len(lookaheads) != len(reduction.Lookaheads) {
				distinguishing[reduction.Item] = true
			}
		}
	}

	result := make(map[int][]ConflictReportReduction, len(reductionsByStateIdx))
	for stateIdx, reductions := range reductionsByStateIdx {
		var kept []ConflictReportReduction
		for _, reduction := range reductions {
			if !distinguishing[reduction.Item] {
				continue
			}
			kept = append(kept, ConflictReportReduction{
				Item:       reduction.Item,
				Lookaheads: subtractSorted(reduction.Lookaheads, shared[reduction.Item]),
			})
		}
		result[stateIdx] = kept
	}
	return result
}

// intersectSorted returns the elements of the sorted slice a which the sorted slice b holds as well.
func intersectSorted(a []string, b []string) []string {
	var result []string
	for _, element := range a {
		if _, found := slices.BinarySearch(b, element); found {
			result = append(result, element)
		}
	}
	return result
}

// subtractSorted returns the elements of the sorted slice a which the sorted slice b does not hold.
func subtractSorted(a []string, b []string) []string {
	var result []string
	for _, element := range a {
		if _, found := slices.BinarySearch(b, element); !found {
			result = append(result, element)
		}
	}
	return result
}

// buildConflictReport builds the report of the state of the conflict, without any entry yet. It holds every reduction
// of the state with its complete lookaheads, which keepDistinguishingLookaheads reduces to what is needed.
func buildConflictReport(grammar frontend.Grammar, c Conflict) ConflictReport {
	report := ConflictReport{
		StateIdx: c.StateIdx,
	}
	for _, core := range c.KernelItems.All() {
		report.KernelItems = append(report.KernelItems, formatKernelItem(grammar, core))
	}
	slices.Sort(report.KernelItems)

	lookaheadsByItem := make(map[string][]string)
	for _, reduceAction := range c.ReduceActions.All() {
		production := grammar.Productions[reduceAction.ProductionIdx]
		item := formatSymbols(grammar, production, len(production.SymbolRefs))
		for terminalIdx := range reduceAction.LookaheadSet.All() {
			lookaheadsByItem[item] = append(lookaheadsByItem[item], grammar.Terminals[terminalIdx].String())
		}
	}
	for item, lookaheads := range lookaheadsByItem {
		slices.Sort(lookaheads)
		report.Reductions = append(report.Reductions, ConflictReportReduction{
			Item:       item,
			Lookaheads: slices.Compact(lookaheads),
		})
	}
	slices.SortFunc(report.Reductions, compareReductions)
	return report
}

// buildConflictReportEntry renders a single conflict: the terminal it occurred on, the actions which competed for that
// terminal once precedence and associativity had decided what they could, and what the policy decided about them.
func buildConflictReportEntry(grammar frontend.Grammar, c Conflict) ConflictReportEntry {
	entry := ConflictReportEntry{
		Terminal: grammar.Terminals[c.TerminalIdx].String(),
		Kind:     conflictKind(c),
	}
	entry.Chosen, entry.Decision, entry.DecisionContributions = formatDecision(grammar, c.Decision)
	for _, contribution := range c.Undeclared.All() {
		entry.Contributions = append(entry.Contributions, formatContribution(grammar, contribution))
	}
	slices.SortFunc(entry.Contributions, compareContributions)
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
// the internals of the resolution. It returns the action which won, or else the decision and the sorted actions the
// decision is about.
func formatDecision(grammar frontend.Grammar, decision Decision) (string, string, []string) {
	switch decision.Kind {
	case DecisionDominant:
		return formatContribution(grammar, decision.Dominant), "", nil
	case DecisionError:
		return "", "resolved by rejecting the terminal, so the parser reports a syntax error on it", nil
	case DecisionUnresolved:
		var contributions []string
		for _, contribution := range decision.Unresolved.All() {
			contributions = append(contributions, formatContribution(grammar, contribution))
		}
		slices.SortFunc(contributions, compareContributions)
		return "", "unresolved between:", contributions
	case DecisionUndefined:
		return "", "no action to decide about", nil
	}
	return "", "unknown decision", nil
}

// chosenMarker follows the action which won a conflict.
const chosenMarker = " (chosen)"

// itemDot marks the position of an item.
const itemDot = "•"

// shiftText is how a shift is rendered in the report.
const shiftText = "shift"

// compareContributions orders the rendered actions lexically, except that a shift comes first.
func compareContributions(a string, b string) int {
	if (a == shiftText) != (b == shiftText) {
		if a == shiftText {
			return -1
		}
		return 1
	}
	return strings.Compare(a, b)
}

// formatContribution renders a single competing action of a conflict. A shift is just a shift, and a reduction is
// spelled out with the production it reduces, because a production index changes with unrelated grammar edits.
func formatContribution(grammar frontend.Grammar, contribution Contribution) string {
	if contribution.IsShiftAction() {
		return shiftText
	}
	return "reduce: " + formatProduction(grammar, contribution.ProductionIdx())
}

// formatProduction renders a production with the names of its symbols instead of their indexes, which is what makes the
// report readable next to the grammar file the author wrote.
func formatProduction(grammar frontend.Grammar, productionIdx int) string {
	return formatSymbols(grammar, grammar.Productions[productionIdx], -1)
}

// formatKernelItem renders a kernel item as its production with a dot at the position of the item.
func formatKernelItem(grammar frontend.Grammar, core backend.Core) string {
	return formatSymbols(grammar, grammar.Productions[core.ProductionIdx()], core.Position())
}

// formatSymbols renders the production with the names of its symbols, and with a dot in front of the symbol at the
// given position. A position past the last symbol puts the dot at the end, a negative position omits it. The dot is a
// bullet, so it cannot be mistaken for a terminal like '.'. An empty right hand side is written as (empty), so that a
// production reads the same with and without a dot.
func formatSymbols(grammar frontend.Grammar, production frontend.Production, dotPosition int) string {
	var builder strings.Builder
	builder.WriteString(grammar.Nonterminals[production.NonterminalIdx].String())
	builder.WriteString(" ->")
	if len(production.SymbolRefs) == 0 {
		// An empty right hand side reduces on the empty string, which is easy to miss without a hint.
		builder.WriteString(" (empty)")
	}
	for position, symbolRef := range production.SymbolRefs {
		if position == dotPosition {
			builder.WriteString(" " + itemDot)
		}
		builder.WriteString(" ")
		if symbolRef.IsTerminal() {
			builder.WriteString(grammar.Terminals[symbolRef.Idx()].String())
		} else {
			builder.WriteString(grammar.Nonterminals[symbolRef.Idx()].String())
		}
	}
	if dotPosition == len(production.SymbolRefs) {
		builder.WriteString(" " + itemDot)
	}
	return builder.String()
}
