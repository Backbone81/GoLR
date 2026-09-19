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
// produce, once summarized and once in full, or the error it is expected to fail with.
const (
	reportRootPath            = "testdata/report"
	reportSpecFileName        = "spec.golr"
	reportFileName            = "report.txt"
	reportVerboseFileName     = "report-verbose.txt"
	reportErrorFileName       = "error.txt"
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
			// changes.
			parser, conflicts, err := ielr1golr.GrammarToParser(grammar, policyFactory)
			if err != nil {
				// An unresolved conflict makes the core fail, and the error is what reports the conflict then.
				expectGoldenFile(filepath.Join(casePath, reportErrorFileName), err.Error())
				return
			}

			expectGoldenReport(filepath.Join(casePath, reportFileName), parser.Grammar, conflicts, conflict.ReportConfig{})
			expectGoldenReport(filepath.Join(casePath, reportVerboseFileName), parser.Grammar, conflicts, conflict.ReportConfig{
				Verbose: true,
			})
		},
		Entry("with shift/reduce conflicts", "shift-reduce", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with reduce/reduce conflicts", "reduce-reduce", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with conflicts decided by precedence", "precedence", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with split states sharing their kernel items", "split-states", conflict.PolicyFactory(conflict.DefaultPolicy)),
		Entry("with unresolved conflicts", "unresolved", conflict.PolicyFactory(conflict.PrecedencePolicy)),
		Entry(
			"with unresolved conflicts in split states sharing their kernel items",
			"split-states-unresolved",
			conflict.PolicyFactory(conflict.PrecedencePolicy),
		),
	)
})

// expectGoldenReport compares the report against the committed one, see expectGoldenFile.
func expectGoldenReport(
	goldenPath string,
	grammar frontend.Grammar,
	conflicts []conflict.Conflict,
	config conflict.ReportConfig,
) {
	var builder strings.Builder
	Expect(conflict.WriteConflictReport(&builder, grammar, conflicts, config)).To(Succeed())
	expectGoldenFile(goldenPath, builder.String())
}

// expectGoldenFile compares the content against the committed file, or rewrites the committed file when the update
// environment variable is set.
func expectGoldenFile(goldenPath string, content string) {
	if updatingGoldenReports() {
		Expect(os.WriteFile(goldenPath, []byte(content), 0o644)).To(Succeed())
		return
	}

	expected, err := os.ReadFile(goldenPath)
	Expect(err).ToNot(HaveOccurred(), "run the suite with %s=1 to create the missing file", updateGoldenReportsEnvVar)
	Expect(content).To(Equal(string(expected)))
}

var _ = Describe("WriteConflictReport stability", func() {
	It("should not change with grammar edits which do not touch a conflict", func() {
		basePath := filepath.Join(reportRootPath, "shift-reduce", reportSpecFileName)
		extendedPath := filepath.Join("testdata", "report-stability", "unrelated-edits.golr")

		// The state numbers are expected to differ, otherwise the added production does not prove anything.
		withStateNumbers := conflict.ReportConfig{Verbose: true, WithStateNumbers: true}
		Expect(writeReport(extendedPath, withStateNumbers)).ToNot(Equal(writeReport(basePath, withStateNumbers)))

		verbose := conflict.ReportConfig{Verbose: true}
		Expect(writeReport(extendedPath, verbose)).To(Equal(writeReport(basePath, verbose)))
	})
})

// writeReport writes the report of the grammar in the spec file, built by the same core as the golden reports.
func writeReport(specPath string, config conflict.ReportConfig) string {
	_, grammar, err := golrfrontend.GrammarFromFile(specPath)
	Expect(err).ToNot(HaveOccurred())

	parser, conflicts, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy)
	Expect(err).ToNot(HaveOccurred())

	var builder strings.Builder
	Expect(conflict.WriteConflictReport(&builder, parser.Grammar, conflicts, config)).To(Succeed())
	return builder.String()
}
