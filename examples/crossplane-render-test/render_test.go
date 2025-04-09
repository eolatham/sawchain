package example

import (
	"github.com/eolatham/sawchain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	compositionPath    = "yaml/composition.yaml"
	functionsPath      = "yaml/functions.yaml"
	expectedXrPath     = "yaml/expected-xr.yaml"
	expectedObjectPath = "yaml/expected-object.yaml"
)

// TODO: templatize input YAMLs
// TODO: add failure case
var _ = Describe("Crossplane Render", func() {
	// Create Sawchain instance with fake client
	var sc = sawchain.New(GinkgoTB(), fake.NewClientBuilder().Build())

	DescribeTable("rendering resources with function-go-templating",
		func(xrPath, extraResourcesPath, expectedConfigMapName string) {
			// Run crossplane render
			output, err := runCrossplaneRender(xrPath, compositionPath, functionsPath, extraResourcesPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			// Parse render output into unstructured objects
			resources, err := unstructuredFromYaml(output)
			Expect(err).NotTo(HaveOccurred())

			// Check length of rendered resources
			Expect(resources).To(HaveLen(2))

			// Check rendered XR status (using Sawchain's MatchYAML matcher)
			Expect(resources).To(ContainElement(sc.MatchYAML(expectedXrPath)))

			// Check rendered Object (using Sawchain's MatchYAML matcher)
			bindings := map[string]any{"expectedConfigMapName": expectedConfigMapName}
			Expect(resources).To(ContainElement(sc.MatchYAML(expectedObjectPath, bindings)))
		},
		Entry("dev environment", "yaml/xr-dev.yaml", "yaml/extra-resources-dev.yaml", "my-awesome-dev-bucket-bucket"),
		Entry("prod environment", "yaml/xr-prod.yaml", "yaml/extra-resources-prod.yaml", "my-awesome-prod-bucket-bucket"),
	)
})
