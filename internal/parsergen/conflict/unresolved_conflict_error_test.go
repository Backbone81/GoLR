package conflict_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/internal/parsergen/conflict"
)

var _ = Describe("UnresolvedConflictErrors", func() {
	It("should find every unresolved conflict error of a wrapped join", func() {
		first := newUnresolvedConflictError(1)
		second := newUnresolvedConflictError(2)
		err := fmt.Errorf("building the parser: %w", errors.Join(first, errors.New("unrelated"), second))

		Expect(conflict.UnresolvedConflictErrors(err)).To(Equal([]conflict.UnresolvedConflictError{first, second}))
	})

	It("should find a single unresolved conflict error which is not joined", func() {
		err := newUnresolvedConflictError(1)

		Expect(conflict.UnresolvedConflictErrors(err)).To(Equal([]conflict.UnresolvedConflictError{err}))
	})

	It("should find nothing in an error without unresolved conflicts", func() {
		Expect(conflict.UnresolvedConflictErrors(errors.New("unrelated"))).To(BeEmpty())
		Expect(conflict.UnresolvedConflictErrors(nil)).To(BeEmpty())
	})
})

// newUnresolvedConflictError creates an unresolved conflict error which is told apart from others by its state index.
func newUnresolvedConflictError(stateIdx int) conflict.UnresolvedConflictError {
	return conflict.UnresolvedConflictError{
		Conflict: conflict.Conflict{StateIdx: stateIdx},
		Report:   conflict.ConflictReport{StateIdx: stateIdx},
	}
}
