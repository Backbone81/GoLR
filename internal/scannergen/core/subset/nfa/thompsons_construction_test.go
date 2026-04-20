package nfa_test

import (
	thompsonsnfa "golr/internal/scannergen/core/subset/nfa"
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThompsonsConstruction", func() {
	It("should panic when given an invalid regex node", func() {
		Expect(func() {
			thompsonsnfa.NewThompsonsConstruction().Build(frontend.NewNodeCharClass(frontend.CharClass{
				Ranges: []frontend.CharRange{
					{
						Low:  'z',
						High: 'a',
					},
				},
			}), 0)
		}).To(Panic())
	})

	It("should panic when constructing an invalid NFA", func() {
		Expect(func() {
			thompsonsnfa.NewThompsonsConstruction().MustBeValidNFA([]thompsonsnfa.State{})
		}).To(Panic())

		Expect(func() {
			thompsonsnfa.NewThompsonsConstruction().MustBeValidNFA([]thompsonsnfa.State{
				{
					Accept: true,
				},
			})
		}).To(Panic())

		Expect(func() {
			thompsonsnfa.NewThompsonsConstruction().MustBeValidNFA([]thompsonsnfa.State{
				{},
				{
					Accept: true,
				},
				{
					Accept: true,
				},
			})
		}).To(Panic())
	})
})
