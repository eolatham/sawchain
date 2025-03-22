package matchers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain/internal/testutil"
)

// Variables must be assigned inline to beat static Entry parsing!
var (
	standardScheme         = testutil.NewStandardScheme()
	standardClient         = fake.NewClientBuilder().WithScheme(standardScheme).Build()
	schemeWithTestResource = testutil.NewStandardSchemeWithTestResource()
	clientWithTestResource = fake.NewClientBuilder().WithScheme(schemeWithTestResource).Build()
)

func TestMatchers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Matchers Suite")
}
