package core_test

import (
	"maps"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	intfrontend "github.com/backbone81/golr/internal/parsergen/frontend"
	"github.com/backbone81/golr/pkg/parsergen/conflict"
	"github.com/backbone81/golr/pkg/parsergen/core"
	ielr1bison "github.com/backbone81/golr/pkg/parsergen/core/ielr1/bison"
	ielr1golr "github.com/backbone81/golr/pkg/parsergen/core/ielr1/golr"
	"github.com/backbone81/golr/pkg/parsergen/frontend"
	golrfrontend "github.com/backbone81/golr/pkg/parsergen/frontend/golr"
	"github.com/backbone81/golr/pkg/utils"
)

const (
	unreachableWarning = `nonterminal "orphan" is unreachable from the start nonterminal "s"`
	neverReducedY      = `production y -> "a" is never reduced after conflict resolution`
	neverReducedX      = `production x -> "a" is never reduced after conflict resolution`
)

var _ = Describe("Warnings", func() {
	It("should make every core fail on an unproductive start nonterminal", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			}

			@parser {
			    s
			        : s "a"
			        ;
			}
		`)
		for coreName, grammarToParser := range allCores() {
			parser, _, warnings, err := grammarToParser(grammar)
			Expect(err).To(MatchError(`start nonterminal "s" does not derive any finite string`), "core %s", coreName)
			Expect(parser).To(BeZero(), "core %s", coreName)
			Expect(warnings).To(BeEmpty(), "core %s", coreName)
		}
	})

	It("should make every core warn about an unreachable nonterminal", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    s
			        : "a"
			        ;

			    orphan
			        : "b"
			        ;
			}
		`)
		for coreName, grammarToParser := range allCores() {
			_, _, warnings, err := grammarToParser(grammar)
			Expect(err).ToNot(HaveOccurred(), "core %s", coreName)
			Expect(warningMessages(warnings)).To(Equal([]string{unreachableWarning}), "core %s", coreName)
		}
	})

	DescribeTable("should make the IELR(1) cores agree on the productions which are never reduced",
		func(spec string, wantWarnings []string) {
			grammar := grammarFromSpec(spec)
			_, _, golrWarnings, err := ielr1golr.GrammarToParser(grammar)
			Expect(err).ToNot(HaveOccurred())
			_, _, bisonWarnings, err := ielr1bison.GrammarToParser(grammar)
			Expect(err).ToNot(HaveOccurred())

			Expect(warningMessages(golrWarnings)).To(Equal(wantWarnings))
			Expect(warningMessages(bisonWarnings)).To(Equal(wantWarnings))
		},
		Entry("a reduce/reduce conflict",
			`
			@scanner {
			    A: "a";
			}

			@parser {
			    s
			        : x
			        | y
			        ;

			    x
			        : "a"
			        ;

			    y
			        : "a"
			        ;
			}
			`,
			[]string{neverReducedY},
		),
		Entry("a shift/reduce conflict decided by precedence",
			`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    @precedence {
			        @left: "b";
			        @left: "a";
			    }

			    s
			        : "a" "b"
			        | x "b"
			        ;

			    x
			        : "a"
			        ;
			}
			`,
			[]string{neverReducedX},
		),
	)

	It("should keep the grammar as written and still report the conflicts next to the warnings", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    s
			        : x
			        | y
			        ;

			    x
			        : "a"
			        ;

			    y
			        : "a"
			        ;

			    orphan
			        : "b"
			        ;
			}
		`)
		for coreName, grammarToParser := range golrCores {
			parser, conflicts, warnings, err := grammarToParser(grammar)
			Expect(err).ToNot(HaveOccurred(), "core %s", coreName)
			Expect(parser.Grammar).To(Equal(intfrontend.AugmentGrammar(grammar)), "core %s", coreName)
			Expect(conflicts).To(HaveLen(1), "core %s", coreName)
			Expect(warningMessages(warnings)).To(
				Equal([]string{unreachableWarning, neverReducedY}),
				"core %s", coreName,
			)
		}
	})

	It("should make every core fail on warnings when asked to", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    s
			        : "a"
			        ;

			    orphan
			        : "b"
			        ;
			}
		`)
		for coreName, grammarToParser := range allCores() {
			parser, _, warnings, err := grammarToParser(grammar, core.FailOnWarnings())
			Expect(err).To(MatchError(unreachableWarning), "core %s", coreName)
			Expect(parser).To(BeZero(), "core %s", coreName)
			Expect(warnings).To(BeEmpty(), "core %s", coreName)
		}
	})

	It("should make every core join the warnings into the error of an unproductive nonterminal", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    s
			        : "a"
			        ;

			    orphan
			        : orphan "b"
			        ;
			}
		`)
		for coreName, grammarToParser := range allCores() {
			_, _, warnings, err := grammarToParser(grammar, core.FailOnWarnings())
			Expect(err).To(MatchError(ContainSubstring(`nonterminal "orphan" does not derive any finite string`)),
				"core %s", coreName)
			Expect(err).To(MatchError(ContainSubstring(unreachableWarning)), "core %s", coreName)
			Expect(warnings).To(BeEmpty(), "core %s", coreName)
		}
	})

	It("should leave the result of every core unchanged when failing on warnings without any", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			}

			@parser {
			    s
			        : x x
			        ;

			    x
			        : "a"
			        | @empty
			        ;
			}
		`)
		for coreName, grammarToParser := range allCores() {
			wantParser, wantConflicts, wantWarnings, err := grammarToParser(grammar)
			Expect(err).ToNot(HaveOccurred(), "core %s", coreName)
			Expect(wantWarnings).To(BeEmpty(), "core %s", coreName)

			parser, conflicts, warnings, err := grammarToParser(grammar, core.FailOnWarnings())
			Expect(err).ToNot(HaveOccurred(), "core %s", coreName)
			Expect(parser).To(Equal(wantParser), "core %s", coreName)
			Expect(conflicts).To(Equal(wantConflicts), "core %s", coreName)
			Expect(warnings).To(BeEmpty(), "core %s", coreName)
		}
	})

	It("should make the GoLR cores join the warnings into the error of unresolved conflicts", func() {
		grammar := grammarFromSpec(`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    s
			        : x
			        | y
			        ;

			    x
			        : "a"
			        ;

			    y
			        : "a"
			        ;

			    orphan
			        : "b"
			        ;
			}
		`)
		for coreName, grammarToParser := range golrCores {
			parser, conflicts, warnings, err := grammarToParser(grammar, core.FailOnConflicts(), core.FailOnWarnings())
			Expect(err).To(MatchError(ContainSubstring(unreachableWarning)), "core %s", coreName)
			Expect(conflict.UnresolvedConflictErrors(err)).To(HaveLen(1), "core %s", coreName)
			Expect(parser).To(BeZero(), "core %s", coreName)
			Expect(conflicts).To(HaveLen(1), "core %s", coreName)
			Expect(warnings).To(BeEmpty(), "core %s", coreName)
		}
	})
})

// allCores returns the GoLR and the Bison cores together.
func allCores() map[string]grammarToParser {
	result := maps.Clone(golrCores)
	maps.Copy(result, bisonCores)
	return result
}

// grammarFromSpec reads the grammar of a GoLR specification.
func grammarFromSpec(spec string) frontend.Grammar {
	_, grammar, err := golrfrontend.GrammarFromString(spec)
	Expect(err).ToNot(HaveOccurred())
	return grammar
}

// warningMessages returns the messages of the warnings, which is what the expectations compare.
func warningMessages(warnings []utils.Warning) []string {
	var result []string
	for _, warning := range warnings {
		result = append(result, warning.Error())
	}
	return result
}
