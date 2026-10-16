package yaml_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestYAML(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scannergen: Frontend: YAML Suite")
}
