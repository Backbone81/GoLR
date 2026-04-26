package frontend_test

import (
	"golr/internal/scannergen/frontend/dsl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Any", func() {
	It("should convert to string", func() {
		expression := dsl.Any()
		Expect(expression.String()).To(Equal("."))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := dsl.Any()
		Expect(expression.IsSingleNode()).To(BeTrue())
	})

	It("should correctly validate", func() {
		expression := dsl.Any()
		Expect(expression.Validate()).To(BeNil())
	})
})
