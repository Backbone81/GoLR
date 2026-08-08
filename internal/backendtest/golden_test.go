package backendtest_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/backendtest/interpreter"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	ielr1golrcore "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	subsetcore "github.com/backbone81/golr/internal/scannergen/core/subset"
	golrfrontend "github.com/backbone81/golr/internal/scannergen/frontend/golr"
)

// The corpus is a directory per case, and every file in it has a fixed name. A case is therefore fully described by the
// three or four files sitting next to each other, which is what makes reviewing the committed trace a single glance
// instead of a hunt through parallel directory trees.
const (
	// goldenRootPath is the directory holding one directory per case.
	goldenRootPath = "golden"

	// specFileName is the GoLR specification of the case, carrying the scanner rules and, for a parser case, the
	// grammar as well.
	specFileName = "spec.golr"

	// inputFileName is the input to scan and parse. It is read as raw bytes and is deliberately not required to be
	// valid UTF-8, because what a scanner does with a byte no rule can match is part of what the corpus pins down.
	inputFileName = "input.txt"

	// scannerTraceFileName is the trace every backend has to reproduce for the scanner.
	scannerTraceFileName = "scanner.trace"

	// parserTraceFileName is the trace every backend has to reproduce for the parser.
	parserTraceFileName = "parser.trace"
)

// updateGoldenEnvVar makes a run rewrite the committed traces instead of comparing against them. Reviewing the diff it
// produces is the point of committing the traces in the first place, so it is a deliberate step and never the default.
const updateGoldenEnvVar = "UPDATE_GOLDEN"

// updatingGolden reports whether this run rewrites the committed traces instead of comparing against them. The rewrite
// and the failure which reports it both read the environment through here, so the two can never disagree about what
// the run is doing.
func updatingGolden() bool {
	return os.Getenv(updateGoldenEnvVar) == "1"
}

// A run which rewrote the traces fails on purpose, once the whole corpus has been written. The traces are the review
// artifact of this corpus, so rewriting them has to be impossible to overlook, and a passing package prints nothing at
// all unless `go test` is run with -v. A failure is therefore the only signal which reaches a plain `make test`, and it
// doubles as the guard against the variable being left set where nobody looks at the traces afterwards. Reporting it
// here rather than at each rewrite is what keeps the corpus complete: a failure raised inside a spec would abort that
// spec, and the case would keep the trace which the aborted half never got around to writing.
var _ = AfterSuite(func() {
	if !updatingGolden() {
		return
	}

	Fail(fmt.Sprintf(
		"The committed traces of the golden corpus were rewritten because the environment variable %s is set to '1'."+
			" Review the diff they produced, then run the suite again without the variable to confirm the corpus passes.",
		updateGoldenEnvVar,
	))
})

// The golden corpus is the set of cases every language backend is held to.
var _ = Describe("Golden corpus", func() {
	for _, caseName := range goldenCaseNames() {
		It("reproduces the committed traces of "+caseName, func() {
			casePath := filepath.Join(goldenRootPath, caseName)

			// One read yields both the scanner rules and the grammar, because the GoLR format carries them in the same
			// document. That is what keeps a case to a single specification file.
			rules, grammar, err := golrfrontend.RulesFromFile(filepath.Join(casePath, specFileName))
			Expect(err).ToNot(HaveOccurred())
			Expect(rules).ToNot(BeEmpty())

			source, err := os.ReadFile(filepath.Join(casePath, inputFileName))
			Expect(err).ToNot(HaveOccurred())

			// Every case commits both traces, whichever of the two it was written for. A case aimed at the scanner
			// carries a grammar of a single empty production and its parser trace is correspondingly short; the point
			// of having it anyway is that a backend which fails both traces is broken somewhere else than one which
			// only fails the parser trace.
			dfa := subsetcore.RulesToDFA(rules)
			expectGolden(
				filepath.Join(casePath, scannerTraceFileName),
				interpreter.ScanTrace(dfa, source).String(),
			)

			// The core is named explicitly rather than left to the default, for the same reason scripts/generate.sh
			// pins one: the committed traces must not change underneath the corpus when the default core changes.
			parser, _, err := ielr1golrcore.GrammarToParser(grammar, conflict.DefaultPolicy)
			Expect(err).ToNot(HaveOccurred())
			expectGolden(
				filepath.Join(casePath, parserTraceFileName),
				interpreter.ParseTrace(parser, dfa, source).String(),
			)
		})
	}
})

// goldenCaseNames returns the name of every case in the corpus, which is the name of its directory. It runs while the
// spec tree is built rather than inside a spec, so every case is a spec of its own and a failure names the case which
// produced it. A corpus which cannot be read is a broken repository and not a failing case, so it stops the suite
// outright instead of being reported against one arbitrary case.
func goldenCaseNames() []string {
	entries, err := os.ReadDir(goldenRootPath)
	if err != nil {
		panic("reading the golden corpus: " + err.Error())
	}

	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name())
		}
	}
	return result
}

// expectGolden compares the produced trace against the committed one, or rewrites the committed one when the update
// environment variable is set. A missing trace file is reported with the way to create it, because that is the one
// failure a new case runs into first.
func expectGolden(goldenPath string, actual string) {
	if updatingGolden() {
		Expect(os.WriteFile(goldenPath, []byte(actual), 0o644)).To(Succeed())
		return
	}

	expected, err := os.ReadFile(goldenPath)
	Expect(err).ToNot(HaveOccurred(), "run the suite with %s=1 to create the missing trace", updateGoldenEnvVar)
	Expect(actual).To(Equal(string(expected)))
}
