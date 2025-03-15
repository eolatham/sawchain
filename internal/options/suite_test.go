package options_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	tempDir         string
	templateFile    string
	templateContent string
)

func TestOptions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Options Suite")
}

var _ = BeforeSuite(func() {
	// Create a temporary directory for template files
	var err error
	tempDir, err = os.MkdirTemp("", "options-test-")
	Expect(err).NotTo(HaveOccurred())

	// Create a template file
	templateContent = "This is a test template"
	templateFile = filepath.Join(tempDir, "template.yaml")
	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	// Clean up the temporary directory
	os.RemoveAll(tempDir)
})
