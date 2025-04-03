package chainsaw_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain/internal/testutil"
)

// Variables must be assigned inline to beat static Entry parsing!
var (
	ctx       = context.Background()
	scheme    = testutil.NewStandardScheme()
	k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
)

func TestChainsaw(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chainsaw Suite")
}
