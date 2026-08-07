package backendtest_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBackendTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backend Test Suite")
}
