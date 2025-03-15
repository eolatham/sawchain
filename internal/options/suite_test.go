package options_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Values must be assigned inline to beat static Entry parsing!
var templateFilePath, templateFileContent = createTemplateFile()

func TestOptions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Options Suite")
}

var _ = AfterSuite(func() {
	Expect(os.Remove(templateFilePath)).To(Succeed())
})

func createTemplateFile() (string, string) {
	file, err := os.CreateTemp("", "template-*.yaml")
	if err != nil {
		panic(err)
	}
	path := file.Name()
	content := "template file content"
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	return path, content
}
