package dfa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDfa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scannergen: Cores: Subset: DFA Suite")
}
