package golr_test

import (
	"testing"

	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIelr1(t *testing.T) {
	// we want to use fmt.Stringer for displaying data in failed tests, so that we have an easier time understanding
	// packed data.
	format.UseStringerRepresentation = true

	RegisterFailHandler(Fail)
	RunSpecs(t, "Parsergen: Cores: IELR(1) Go Suite")
}
