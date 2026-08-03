package table_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/backend/table"
)

var _ = Describe("Action", func() {
	It("returns the state a shift action carries", func() {
		action := table.NewShiftAction(4711)

		Expect(action.Kind()).To(Equal(table.ActionKindShift))
		Expect(action.StateIdx()).To(Equal(4711))
	})

	It("returns the production a reduce action carries", func() {
		action := table.NewReduceAction(4711)

		Expect(action.Kind()).To(Equal(table.ActionKindReduce))
		Expect(action.ProductionIdx()).To(Equal(4711))
	})

	It("encodes the reduce by the accept production as the accept", func() {
		// Reducing by `$accept -> Start $end` finishes the parse instead of pushing a nonterminal, which every caller
		// gets right for free because the encoding happens in one place.
		Expect(table.NewReduceAction(0)).To(Equal(table.NewAcceptAction()))
		Expect(table.NewAcceptAction().Kind()).To(Equal(table.ActionKindAccept))
	})

	It("keeps every kind apart", func() {
		kinds := []table.ActionKind{
			table.NewShiftAction(0).Kind(),
			table.NewReduceAction(1).Kind(),
			table.NewAcceptAction().Kind(),
			table.NewErrorAction().Kind(),
		}

		Expect(kinds).To(Equal([]table.ActionKind{
			table.ActionKindShift,
			table.ActionKindReduce,
			table.ActionKindAccept,
			table.ActionKindError,
		}))
	})

	It("keeps the absent entry apart from every encoded action", func() {
		// The absent entry is what a lookup returns for a terminal a state has no action for, so it must not collide
		// with an action, not even with the shift into the state with index 0.
		Expect(table.NewShiftAction(0)).ToNot(Equal(table.NoAction))
		Expect(table.NewReduceAction(1)).ToNot(Equal(table.NoAction))
		Expect(table.NewAcceptAction()).ToNot(Equal(table.NoAction))
		Expect(table.NewErrorAction()).ToNot(Equal(table.NoAction))
	})

	It("stays within a 16 bit table entry for a grammar of a realistic size", func() {
		// The magnitude of an action decides the integer type a backend can emit the table with, which is what putting
		// the kind into the low bits is for. A grammar with a few thousand states and productions has to fit into a
		// 16 bit entry, or every table would have to be emitted twice as wide.
		Expect(int(table.NewShiftAction(16000))).To(BeNumerically("<=", 0xFFFF))
		Expect(int(table.NewReduceAction(16000))).To(BeNumerically("<=", 0xFFFF))
	})

	It("describes itself", func() {
		Expect(table.NewShiftAction(42).String()).To(Equal("(shift, state 42)"))
		Expect(table.NewReduceAction(42).String()).To(Equal("(reduce, production 42)"))
		Expect(table.NewAcceptAction().String()).To(Equal("(accept)"))
		Expect(table.NewErrorAction().String()).To(Equal("(error)"))
		Expect(table.NoAction.String()).To(Equal("(no action)"))
	})
})
