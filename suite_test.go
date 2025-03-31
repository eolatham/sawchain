package sawchain_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain/internal/testutil"
)

// Variables must be assigned inline to beat static Entry parsing!
var (
	ctx = context.Background()

	fastTimeout  = 100 * time.Millisecond
	fastInterval = 5 * time.Millisecond

	standardScheme = testutil.NewStandardScheme()
	standardClient = fake.NewClientBuilder().WithScheme(standardScheme).Build()

	schemeWithTestResource = testutil.NewStandardSchemeWithTestResource()
	clientWithTestResource = fake.NewClientBuilder().WithScheme(schemeWithTestResource).Build()
)

func TestSawchain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sawchain Suite")
}
