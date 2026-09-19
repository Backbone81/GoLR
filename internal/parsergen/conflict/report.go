package conflict

import (
	"cmp"
	"fmt"
	"io"
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
	removeUnambiguousLookaheads(reports)
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

	// KernelItems are the kernel items of the conflicted state, which name the state independently of its index.
	KernelItems []ConflictReportKernelItem

	// Entries are the conflicted terminals of the state.
	Entries []ConflictReportEntry
}

// ConflictReportKernelItem is a single kernel item of a conflicted state.
type ConflictReportKernelItem struct {
	// Item is the kernel item with a dot at its position.
	Item string

	// Lookaheads are the names of the terminals the completed item reduces on, before any conflict was resolved. They are
	// removed unless another state of the same report has the same kernel items, because the lookaheads are what tells
	// those states apart, and they change far more often with unrelated grammar edits than the kernel items do.
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

	// Decision is what the policy decided about the conflict.
	Decision string

	// DecisionContributions are the actions the decision is about: the one which won, or those an unresolved conflict
	// was left with. It is empty for a decision which is not about a particular action.
	DecisionContributions []string
}

// Write writes the report of the state: the kernel items which name the state, followed by one indented block per
// conflicted terminal, all separated by an empty line. The report ends with a single newline, so a caller which writes
// several reports separates them by an empty line of its own.
func (r ConflictReport) Write(w io.Writer, config ReportConfig) error {
	var builder strings.Builder
	if config.WithStateNumbers {
		fmt.Fprintf(&builder, "state %d:\n", r.StateIdx)
	} else {
		builder.WriteString("state:\n")
	}
	for _, kernelItem := range r.KernelItems {
		if len(kernelItem.Lookaheads) == 0 {
			fmt.Fprintf(&builder, "  %s\n", kernelItem.Item)
			continue
		}
		fmt.Fprintf(&builder, "  %s  {%s}\n", kernelItem.Item, strings.Join(kernelItem.Lookaheads, ", "))
	}
	for _, entry := range r.Entries {
		fmt.Fprintf(&builder, "\n  %s on terminal %s:\n", entry.Kind, entry.Terminal)
		for _, contribution := range entry.Contributions {
			fmt.Fprintf(&builder, "    %s\n", contribution)
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

// compareConflictReports orders the reports by the number of kernel items, then by the kernel items, then by the
// terminals of the entries. The state index is the final tie-breaker, so the order is total even when the state numbers
// are written.
func compareConflictReports(a ConflictReport, b ConflictReport) int {
	if result := cmp.Compare(len(a.KernelItems), len(b.KernelItems)); result != 0 {
		return result
	}
	result := slices.CompareFunc(a.KernelItems, b.KernelItems, compareKernelItems)
	if result != 0 {
		return result
	}
	result = slices.CompareFunc(a.KernelItems, b.KernelItems, compareLookaheads)
	if result != 0 {
		return result
	}
	result = slices.CompareFunc(a.Entries, b.Entries, func(a ConflictReportEntry, b ConflictReportEntry) int {
		return strings.Compare(a.Terminal, b.Terminal)
	})
	if result != 0 {
		return result
	}
	return cmp.Compare(a.StateIdx, b.StateIdx)
}

// compareKernelItems orders two kernel items by the item alone, which is unique within a state.
func compareKernelItems(a ConflictReportKernelItem, b ConflictReportKernelItem) int {
	return strings.Compare(a.Item, b.Item)
}

// compareLookaheads orders two kernel items by the number of their lookaheads, then by the lookaheads.
func compareLookaheads(a ConflictReportKernelItem, b ConflictReportKernelItem) int {
	if result := cmp.Compare(len(a.Lookaheads), len(b.Lookaheads)); result != 0 {
		return result
	}
	return slices.Compare(a.Lookaheads, b.Lookaheads)
}

// removeUnambiguousLookaheads removes the lookaheads from every report whose kernel items no report of another state
// shares. A state is then already told apart by its kernel items, and leaving the lookaheads out keeps the report
// stable when an unrelated grammar edit changes them. Several reports can belong to the same state, which does not
// make their kernel items ambiguous.
func removeUnambiguousLookaheads(reports []ConflictReport) {
	stateIdxsByKernel := make(map[string]map[int]struct{})
	for _, report := range reports {
		kernel := report.kernel()
		if stateIdxsByKernel[kernel] == nil {
			stateIdxsByKernel[kernel] = make(map[int]struct{})
		}
		stateIdxsByKernel[kernel][report.StateIdx] = struct{}{}
	}

	for _, report := range reports {
		if len(stateIdxsByKernel[report.kernel()]) > 1 {
			continue
		}
		for i := range report.KernelItems {
			report.KernelItems[i].Lookaheads = nil
		}
	}
}

// kernel returns the kernel items of the report as a single string, which identifies the kernel in a map.
func (r ConflictReport) kernel() string {
	items := make([]string, 0, len(r.KernelItems))
	for _, kernelItem := range r.KernelItems {
		items = append(items, kernelItem.Item)
	}
	return strings.Join(items, "\n")
}

// buildConflictReport builds the report of the state of the conflict, without any entry yet.
func buildConflictReport(grammar frontend.Grammar, c Conflict) ConflictReport {
	report := ConflictReport{
		StateIdx: c.StateIdx,
	}
	for _, core := range c.KernelItems.All() {
		report.KernelItems = append(report.KernelItems, ConflictReportKernelItem{
			Item:       formatKernelItem(grammar, core),
			Lookaheads: formatKernelItemLookaheads(grammar, c.ReduceActions, core),
		})
	}
	slices.SortFunc(report.KernelItems, compareKernelItems)
	return report
}

// buildConflictReportEntry renders a single conflict: the terminal it occurred on, the actions which competed for that
// terminal once precedence and associativity had decided what they could, and what the policy decided about them.
func buildConflictReportEntry(grammar frontend.Grammar, c Conflict) ConflictReportEntry {
	entry := ConflictReportEntry{
		Terminal: grammar.Terminals[c.TerminalIdx].String(),
		Kind:     conflictKind(c),
	}
	entry.Decision, entry.DecisionContributions = formatDecision(grammar, c.Decision)
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
// the internals of the resolution. It returns the decision and the sorted actions the decision is about.
func formatDecision(grammar frontend.Grammar, decision Decision) (string, []string) {
	switch decision.Kind {
	case DecisionDominant:
		return "resolved in favor of:", []string{formatContribution(grammar, decision.Dominant)}
	case DecisionError:
		return "resolved by rejecting the terminal, so the parser reports a syntax error on it", nil
	case DecisionUnresolved:
		var contributions []string
		for _, contribution := range decision.Unresolved.All() {
			contributions = append(contributions, formatContribution(grammar, contribution))
		}
		slices.SortFunc(contributions, compareContributions)
		return "unresolved between:", contributions
	case DecisionUndefined:
		return "no action to decide about", nil
	}
	return "unknown decision", nil
}

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
	production := grammar.Productions[productionIdx]
	if len(production.SymbolRefs) == 0 {
		// An empty right hand side reduces on the empty string, which is easy to miss without a hint.
		return grammar.Nonterminals[production.NonterminalIdx].String() + " -> (empty)"
	}
	return formatSymbols(grammar, production, -1)
}

// formatKernelItem renders a kernel item as its production with a dot at the position of the item.
func formatKernelItem(grammar frontend.Grammar, core backend.Core) string {
	return formatSymbols(grammar, grammar.Productions[core.ProductionIdx()], core.Position())
}

// formatKernelItemLookaheads returns the sorted names of the terminals the kernel item reduces on, or nil when the item
// is not completed and so does not reduce.
func formatKernelItemLookaheads(
	grammar frontend.Grammar,
	reduceActions backend.ReduceActionSet,
	core backend.Core,
) []string {
	if core.Position() != len(grammar.Productions[core.ProductionIdx()].SymbolRefs) {
		return nil
	}

	var result []string
	for _, reduceAction := range reduceActions.All() {
		if reduceAction.ProductionIdx != core.ProductionIdx() {
			continue
		}
		for terminalIdx := range reduceAction.LookaheadSet.All() {
			result = append(result, grammar.Terminals[terminalIdx].String())
		}
	}
	slices.Sort(result)
	return result
}

// formatSymbols renders the production with the names of its symbols, and with a dot in front of the symbol at the
// given position. A position past the last symbol puts the dot at the end, a negative position omits it.
func formatSymbols(grammar frontend.Grammar, production frontend.Production, dotPosition int) string {
	var builder strings.Builder
	builder.WriteString(grammar.Nonterminals[production.NonterminalIdx].String())
	builder.WriteString(" ->")
	for position, symbolRef := range production.SymbolRefs {
		if position == dotPosition {
			builder.WriteString(" .")
		}
		builder.WriteString(" ")
		if symbolRef.IsTerminal() {
			builder.WriteString(grammar.Terminals[symbolRef.Idx()].String())
		} else {
			builder.WriteString(grammar.Nonterminals[symbolRef.Idx()].String())
		}
	}
	if dotPosition == len(production.SymbolRefs) {
		builder.WriteString(" .")
	}
	return builder.String()
}
