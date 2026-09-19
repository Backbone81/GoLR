package conflict_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golr "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/internal/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

// The report cases are a directory per case under reportRootPath, holding the grammar and the report it is expected to
// produce, once summarized and once in full.
const (
	reportRootPath            = "testdata/report"
	reportSpecFileName        = "spec.golr"
	reportFileName            = "report.txt"
	reportVerboseFileName     = "report-verbose.txt"
	updateGoldenReportsEnvVar = "UPDATE_GOLDEN"
)

// updatingGoldenReports reports whether this run rewrites the committed reports instead of comparing against them.
func updatingGoldenReports() bool {
	return os.Getenv(updateGoldenReportsEnvVar) == "1"
}

// A run which rewrote the reports fails on purpose, so the rewrite cannot go unnoticed and the diff gets reviewed.
var _ = AfterSuite(func() {
	if !updatingGoldenReports() {
		return
	}

	Fail(fmt.Sprintf(
		"The committed conflict reports were rewritten because the environment variable %s is set to '1'."+
			" Review the diff they produced, then run the suite again without the variable to confirm they pass.",
		updateGoldenReportsEnvVar,
	))
})

var _ = Describe("WriteConflictReport", func() {
	DescribeTable("should reproduce the committed report",
		func(caseName string, policyFactory conflict.PolicyFactory) {
			casePath := filepath.Join(reportRootPath, caseName)

			_, grammar, err := golrfrontend.GrammarFromFile(filepath.Join(casePath, reportSpecFileName))
			Expect(err).ToNot(HaveOccurred())

			// The core is named explicitly, so the reports do not change underneath the cases when the default core
			// changes. An unresolved conflict makes the core fail, but it still returns every conflict it found.
			_, conflicts, _ := ielr1golr.GrammarToParser(grammar, policyFactory)

			// The conflicts refer to the augmented grammar. The core does not return it when a conflict is left unresolved,
			// so the grammar is augmented here the same way the core does it.
			grammar = frontend.AugmentGrammar(grammar)

			expectGoldenReport(filepath.Join(casePath, reportFileName), grammar, conflicts, conflict.ReportConfig{})
			expectGoldenReport(filepath.Join(casePath, reportVerboseFileName), grammar, conflicts, conflict.ReportConfig{
				Verbose: true,
			})
		},
		Entry("with shift/reduce conflicts", "shift-reduce", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with reduce/reduce conflicts", "reduce-reduce", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with conflicts decided by precedence", "precedence", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with unresolved conflicts", "unresolved", conflict.PolicyFactory(conflict.PrecedencePolicy)),
	)
})

// expectGoldenReport compares the report against the committed one, or rewrites the committed one when the update
// environment variable is set.
func expectGoldenReport(
	goldenPath string,
	grammar frontend.Grammar,
	conflicts []conflict.Conflict,
	config conflict.ReportConfig,
) {
	var builder strings.Builder
	Expect(conflict.WriteConflictReport(&builder, grammar, conflicts, config)).To(Succeed())

	if updatingGoldenReports() {
		Expect(os.WriteFile(goldenPath, []byte(builder.String()), 0o644)).To(Succeed())
		return
	}

	expected, err := os.ReadFile(goldenPath)
	Expect(err).ToNot(HaveOccurred(), "run the suite with %s=1 to create the missing report", updateGoldenReportsEnvVar)
	Expect(builder.String()).To(Equal(string(expected)))
}
