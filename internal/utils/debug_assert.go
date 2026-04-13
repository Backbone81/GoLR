package utils

import (
	"fmt"
)

// DebugAssert is for testing assertions during development and testing. The assertion is an anonymous function which
// returns an error object when the assertion did not hold. The error is then passed on to panic. If validation is
// disabled, the assertion is never checked.
//
// IMPORTANT: Only assert on checks which are logic errors in your implementation. User input should never trigger a
// debug assertion and should be handled with normal error handling.
func DebugAssert(assertion func() error) {
	if !EnableDebugAssertions {
		return
	}
	if err := assertion(); err != nil {
		panic(err.Error())
	}
}

// AssertValidIndex executes a debug assertion which checks for the given index to be not negative and within the
// maximum index provided.
func AssertValidIndex(index int, maxIndex int) {
	AssertValidIndexRange(index, 0, maxIndex)
}

// AssertValidIndexRange executes a debug assertion which checks that the given index is within the requested range.
func AssertValidIndexRange(index int, minIndex int, maxIndex int) {
	DebugAssert(func() error {
		if index < minIndex || maxIndex < index {
			return fmt.Errorf("index out of bounds [%d, %d]", minIndex, maxIndex)
		}
		return nil
	})
}
