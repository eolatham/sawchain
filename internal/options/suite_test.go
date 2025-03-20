package options_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/eolatham/sawchain/internal/testutil"
)

const templateFileContent = "template file content"

// Variables must be assigned inline to beat static Entry parsing!
var templateFilePath = testutil.CreateTempFile("template-*.yaml", templateFileContent)

func TestOptions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Options Suite")
}

var _ = AfterSuite(func() {
	Expect(os.Remove(templateFilePath)).To(Succeed())
})
