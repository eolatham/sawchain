package utilities_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// Values must be assigned inline to beat static Entry parsing!
var (
	tempDir        = createTempDir()
	emptyScheme    = createEmptyScheme()
	standardScheme = createStandardScheme()
)

func TestUtilities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utilities Suite")
}

var _ = AfterSuite(func() {
	Expect(os.RemoveAll(tempDir)).To(Succeed())
})

func createTempDir() string {
	tempDir, err := os.MkdirTemp("", "utilities-test-")
	if err != nil {
		panic(err)
	}
	return tempDir
}

func createEmptyScheme() *runtime.Scheme {
	return runtime.NewScheme()
}

func createStandardScheme() *runtime.Scheme {
	s := createEmptyScheme()
	if err := scheme.AddToScheme(s); err != nil {
		panic(err)
	}
	return s
}
