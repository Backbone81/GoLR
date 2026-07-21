package frontend_test

import (
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	// we want to use fmt.Stringer for displaying data in failed tests, so that we have an easier time understanding
	// packed data.
	format.UseStringerRepresentation = true

	RegisterFailHandler(Fail)
	RunSpecs(t, "Parsergen: Frontend Suite")
}
