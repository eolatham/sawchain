package example

import (
	"github.com/eolatham/sawchain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TODO: implement real example
var _ = Describe("Vela Dry-Run", func() {
	It("temp", func() {
		sc := sawchain.New(GinkgoTB(), fake.NewClientBuilder().Build())
		app := sc.RenderToString("yaml/application.tpl.yaml", map[string]any{
			"port":        8080,
			"annotations": map[string]string{"foo": "bar", "bar": "baz"},
		})
		Expect(app).NotTo(BeEmpty())
	})
})
