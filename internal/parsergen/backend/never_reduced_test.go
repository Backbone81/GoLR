package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend"
	"github.com/backbone81/golr/internal/parsergen/conflict"
	"github.com/backbone81/golr/internal/parsergen/core"
	ielr1golr "github.com/backbone81/golr/internal/parsergen/core/ielr1/golr"
	golrfrontend "github.com/backbone81/golr/internal/parsergen/frontend/golr"
)

var _ = Describe("NeverReducedWarnings", func() {
	DescribeTable("should warn about the productions without a reduction",
		func(spec string, options []core.Option, wantWarnings []string) {
			_, grammar, err := golrfrontend.GrammarFromString(spec)
			Expect(err).ToNot(HaveOccurred())
			parser, _, _, err := ielr1golr.GrammarToParser(grammar, conflict.DefaultPolicy, options...)
			Expect(err).ToNot(HaveOccurred())

			var messages []string
			for _, warning := range backend.NeverReducedWarnings(parser) {
				messages = append(messages, warning.Error())
			}
			Expect(messages).To(Equal(wantWarnings))
		},
		Entry("a grammar without conflicts",
			`
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
			`,
			nil,
			nil,
		),
		Entry("a grammar without conflicts and without default reductions",
			`
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
			`,
			[]core.Option{core.WithoutDefaultReductions()},
			nil,
		),
		Entry("a reduce/reduce conflict decided by the earliest production",
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
			nil,
			[]string{`production y -> "a" is never reduced after conflict resolution`},
		),
		Entry("a shift/reduce conflict decided by precedence, without the productions it cascades to",
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
			nil,
			[]string{`production x -> "a" is never reduced after conflict resolution`},
		),
		Entry("a shift/reduce conflict decided by a nonassociative precedence",
			`
			@scanner {
			    A: "a";
			    B: "b";
			}

			@parser {
			    @precedence {
			        @none: "a" "b";
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
			nil,
			[]string{
				`production s -> "a" "b" is never reduced after conflict resolution`,
				`production x -> "a" is never reduced after conflict resolution`,
			},
		),
		Entry("a production which loses in one state and is reduced in another",
			`
			@scanner {
			    A: "a";
			    B: "b";
			    C: "c";
			}

			@parser {
			    @precedence {
			        @left: "b";
			        @left: "a";
			    }

			    s
			        : "a" "b"
			        | x "b"
			        | "c" x
			        ;

			    x
			        : "a"
			        ;
			}
			`,
			nil,
			nil,
		),
		Entry("an unreachable nonterminal",
			`
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
			`,
			nil,
			nil,
		),
	)
})
