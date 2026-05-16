package nfa_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	thompsonsnfa "github.com/backbone81/golr/internal/scannergen/core/subset/nfa"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("ThompsonsConstruction", func() {
	It("should panic when given an invalid regex node", func() {
		Expect(func() {
			thompsonsnfa.NewThompsonsConstruction().Build(dsl.CharClass(
				dsl.CharRange('z', 'a'),
			), 0)
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
