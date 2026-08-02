package table_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/scannergen/backend"
	"github.com/backbone81/golr/internal/scannergen/backend/table"
	"github.com/backbone81/golr/internal/scannergen/frontend/dsl"
)

var _ = Describe("ByteClasses", func() {
	It("puts all bytes into a single class when no state tells them apart", func() {
		// Without any state there is nothing which could distinguish two bytes, so the initial partition survives.
		byteClasses := table.NewByteClasses(backend.DFA{})

		Expect(byteClasses.Count()).To(Equal(1))
		for byteValue := range table.ByteValueCount {
			Expect(byteClasses.ClassByByte[byteValue]).To(Equal(0))
		}
	})

	It("separates the bytes a state transitions on from the ones it does not", func() {
		dfa := backend.DFA{
			States: []backend.State{
				{
					Transitions: []backend.Transition{
						{ByteRange: backend.ByteRange{Low: 'a', High: 'a'}, StateIdx: 1},
					},
				},
				{},
			},
		}

		byteClasses := table.NewByteClasses(dfa)

		// One class for 'a' and one for every other byte.
		Expect(byteClasses.Count()).To(Equal(2))
		Expect(byteClasses.ClassByByte['b']).To(Equal(byteClasses.ClassByByte['z']))
		Expect(byteClasses.ClassByByte['a']).ToNot(Equal(byteClasses.ClassByByte['b']))
	})

	It("keeps bytes together which lead to the same state in every state", func() {
		dfa := backend.DFA{
			States: []backend.State{
				{
					Transitions: []backend.Transition{
						// 'a' and 'b' are named by separate transitions but always lead to the same place.
						{ByteRange: backend.ByteRange{Low: 'a', High: 'a'}, StateIdx: 1},
						{ByteRange: backend.ByteRange{Low: 'b', High: 'b'}, StateIdx: 1},
						{ByteRange: backend.ByteRange{Low: 'c', High: 'c'}, StateIdx: 2},
					},
				},
				{},
				{},
			},
		}

		byteClasses := table.NewByteClasses(dfa)

		// One class for {'a', 'b'}, one for 'c' and one for the rest.
		Expect(byteClasses.Count()).To(Equal(3))
		Expect(byteClasses.ClassByByte['a']).To(Equal(byteClasses.ClassByByte['b']))
		Expect(byteClasses.ClassByByte['a']).ToNot(Equal(byteClasses.ClassByByte['c']))
	})

	It("splits a class as soon as a single state disagrees about it", func() {
		dfa := backend.DFA{
			States: []backend.State{
				{
					// The first state treats 'a' and 'b' alike.
					Transitions: []backend.Transition{
						{ByteRange: backend.ByteRange{Low: 'a', High: 'b'}, StateIdx: 1},
					},
				},
				{
					// The second state does not, which has to split them apart for good.
					Transitions: []backend.Transition{
						{ByteRange: backend.ByteRange{Low: 'b', High: 'b'}, StateIdx: 1},
					},
				},
			},
		}

		byteClasses := table.NewByteClasses(dfa)

		Expect(byteClasses.Count()).To(Equal(3))
		Expect(byteClasses.ClassByByte['a']).ToNot(Equal(byteClasses.ClassByByte['b']))
		Expect(byteClasses.ClassByByte['a']).ToNot(Equal(byteClasses.ClassByByte['c']))
	})

	It("handles a byte range which ends at the highest byte value", func() {
		// A range ending at 0xFF is the case where iterating with a byte wide loop counter would wrap around instead
		// of terminating.
		dfa := backend.DFA{
			States: []backend.State{
				{
					Transitions: []backend.Transition{
						{ByteRange: backend.ByteRange{Low: 0x80, High: 0xFF}, StateIdx: 1},
					},
				},
				{},
			},
		}

		byteClasses := table.NewByteClasses(dfa)

		Expect(byteClasses.Count()).To(Equal(2))
		Expect(byteClasses.ClassByByte[0xFF]).To(Equal(byteClasses.ClassByByte[0x80]))
		Expect(byteClasses.ClassByByte[0xFF]).ToNot(Equal(byteClasses.ClassByByte[0x7F]))
	})

	It("hands out dense class indices", func() {
		byteClasses := table.NewByteClasses(rulesToDFA(
			dsl.Rule("identifier", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('a', 'z')))),
			dsl.Rule("number", dsl.OneOrMore(dsl.CharClass(dsl.CharRange('0', '9')))),
		))

		// Every class index from 0 to Count - 1 must be in use, otherwise a row would carry columns no byte can
		// ever select.
		used := make([]bool, byteClasses.Count())
		for byteValue := range table.ByteValueCount {
			classIdx := byteClasses.ClassByByte[byteValue]
			Expect(classIdx).To(BeNumerically(">=", 0))
			Expect(classIdx).To(BeNumerically("<", byteClasses.Count()))
			used[classIdx] = true
		}
		for classIdx, isUsed := range used {
			Expect(isUsed).To(BeTrue(), fmt.Sprintf("class %d is never selected by any byte", classIdx))
		}
	})

	It("collapses the byte values of a real scanner into far fewer classes", func() {
		byteClasses := table.NewByteClasses(golrSpecDFA())

		// This is the whole point of the byte classes, so it is worth pinning down that they actually collapse the
		// possible byte values into a much smaller number of columns.
		Expect(byteClasses.Count()).To(BeNumerically("<", table.ByteValueCount/2))
	})
})
