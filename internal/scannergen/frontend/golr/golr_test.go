package golr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/frontend"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
	"github.com/backbone81/golr/internal/scannergen/frontend/golr"
)

var _ = Describe("GoLR Grammar Files", func() {
	It("should return no rules for an empty scanner section", func() {
		source := `
            @scanner {
            }
            @parser {
                file: @empty;
            }
        `
		rules, err := golr.RulesFromString(source)
		Expect(err).ToNot(HaveOccurred())
		Expect(rules).To(BeEmpty())
	})

	Context("Tokens", func() {
		It("should build a literal node for a regex pattern", func() {
			source := `
                @scanner {
                    FOO: /foo/;
                }
                @parser {
                    file: @empty;
                }
            `
			rules, err := golr.RulesFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(rules).To(Equal([]frontend.Rule{
				dsl.Rule("FOO", dsl.Literal("foo")),
			}))
		})

		It("should build a literal node for a string pattern", func() {
			source := `
                @scanner {
                    FOO: "foo";
                }
                @parser {
                    file: @empty;
                }
            `
			rules, err := golr.RulesFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(rules).To(Equal([]frontend.Rule{
				dsl.Rule("FOO", dsl.Literal("foo")),
			}))
		})

		It("should build an empty char class node for an empty token", func() {
			source := `
                @scanner {
                    FOO: @empty;
                }
                @parser {
                    file: @empty;
                }
            `
			rules, err := golr.RulesFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(rules).To(Equal([]frontend.Rule{
				dsl.Rule("FOO", dsl.CharClass()),
			}))
		})

		It("should preserve declaration order for multiple tokens", func() {
			source := `
                @scanner {
                    FOO: /foo/;
                    BAR: "bar";
                    BAZ: @empty;
                }
                @parser {
                    file: @empty;
                }
            `
			rules, err := golr.RulesFromString(source)
			Expect(err).ToNot(HaveOccurred())
			Expect(rules).To(Equal([]frontend.Rule{
				dsl.Rule("FOO", dsl.Literal("foo")),
				dsl.Rule("BAR", dsl.Literal("bar")),
				dsl.Rule("BAZ", dsl.CharClass()),
			}))
		})

		It("should reject a token with an invalid regular expression", func() {
			source := `
                @scanner {
                    FOO: /[unclosed/;
                }
                @parser {
                    file: @empty;
                }
            `
			_, err := golr.RulesFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject duplicate token declarations", func() {
			source := `
                @scanner {
                    FOO: /foo/;
                    FOO: "bar";
                }
                @parser {
                    file: @empty;
                }
            `
			_, err := golr.RulesFromString(source)
			Expect(err).To(HaveOccurred())
		})

		It("should reject duplicate token aliases", func() {
			source := `
                @scanner {
                    FOO: "baz";
                    BAR: "baz";
                }
                @parser {
                    file: @empty;
                }
            `
			_, err := golr.RulesFromString(source)
			Expect(err).To(HaveOccurred())
		})
	})
})
