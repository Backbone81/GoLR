package frontend_test

import (
	"golr/internal/scannergen/frontend"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CharClass", func() {
	It("should convert to string with a single character", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'a',
				},
			},
		}
		Expect(expression.String()).To(Equal("[a]"))
	})

	It("should convert to string with a single character negated", func() {
		expression := frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'a',
				},
			},
		}
		Expect(expression.String()).To(Equal("[^a]"))
	})

	It("should convert to string with two characters", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'a',
				},
				{
					Low:  'b',
					High: 'b',
				},
			},
		}
		Expect(expression.String()).To(Equal("[ab]"))
	})

	It("should convert to string with two characters negated", func() {
		expression := frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'a',
				},
				{
					Low:  'b',
					High: 'b',
				},
			},
		}
		Expect(expression.String()).To(Equal("[^ab]"))
	})

	It("should convert to string with a single character range", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			},
		}
		Expect(expression.String()).To(Equal("[a-z]"))
	})

	It("should convert to string with a single character range negated", func() {
		expression := frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
			},
		}
		Expect(expression.String()).To(Equal("[^a-z]"))
	})

	It("should convert to string with two character ranges", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
				{
					Low:  '0',
					High: '9',
				},
			},
		}
		Expect(expression.String()).To(Equal("[a-z0-9]"))
	})

	It("should convert to string with two character ranges negated", func() {
		expression := frontend.CharClass{
			Negate: true,
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'z',
				},
				{
					Low:  '0',
					High: '9',
				},
			},
		}
		Expect(expression.String()).To(Equal("[^a-z0-9]"))
	})

	It("should provide the correct value for IsSingleNode", func() {
		expression := frontend.CharClass{}
		Expect(expression.IsSingleNode()).To(BeTrue())
	})

	It("should fail validation with zero value", func() {
		expression := frontend.CharClass{}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should fail validation with invalid character range", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  -1,
					High: 'a',
				},
			},
		}
		Expect(expression.Validate()).ToNot(Succeed())
	})

	It("should validate successfully", func() {
		expression := frontend.CharClass{
			Ranges: []frontend.CharRange{
				{
					Low:  'a',
					High: 'a',
				},
			},
		}
		Expect(expression.Validate()).To(Succeed())
	})
})
