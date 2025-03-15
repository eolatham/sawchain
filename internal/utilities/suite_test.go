package utilities_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var tempDir string

func TestUtilities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utilities Suite")
}

var _ = BeforeSuite(func() {
	// Create a temporary directory for test files
	var err error
	tempDir, err = os.MkdirTemp("", "utilities-test-")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Clean up the temporary directory
	os.RemoveAll(tempDir)
})
