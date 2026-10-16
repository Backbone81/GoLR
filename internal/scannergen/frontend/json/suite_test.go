package json_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJSON(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scannergen: Frontend: JSON Suite")
}
