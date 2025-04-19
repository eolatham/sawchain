package example

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestVelaDryRun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vela Dry-Run Suite")
}

var _ = BeforeSuite(func() {
	// Enable better failure output for Sawchain's matchers
	format.UseStringerRepresentation = true
})
