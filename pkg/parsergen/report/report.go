// Package report provides the public API re-exports for the report of the conflicts a parser core returns.
package report

import (
	intreport "github.com/backbone81/golr/internal/parsergen/report"
)

// Config controls what WriteConflictReport and WriteUnresolvedConflictReport write.
type Config = intreport.Config

// WriteConflictReport writes a report of the given conflicts to w. Conflicts the policy resolved on its own are
// summarized and listed in full only with Config.Verbose. Conflicts decided by precedence declarations are not
// reported, and conflicts the policy could not decide are reported by WriteUnresolvedConflictReport.
var WriteConflictReport = intreport.WriteConflictReport

// WriteUnresolvedConflictReport writes the report of a core which failed on unresolved conflicts: the summary of the
// conflicts it returned, followed by the report of every unresolved one. The grammar is the one of the parser tables
// the core returned with the error.
var WriteUnresolvedConflictReport = intreport.WriteUnresolvedConflictReport
