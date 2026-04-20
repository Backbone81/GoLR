package nfa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNfa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scannergen: Core: Subset: NFA Suite")
}
