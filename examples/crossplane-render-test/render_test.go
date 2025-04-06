package example

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain"
)

const yamlDir = "yaml"

func readYaml(fileName string) string {
	bytes, err := os.ReadFile(filepath.Join(yamlDir, fileName))
	Expect(err).NotTo(HaveOccurred())
	return string(bytes)
}

var _ = Describe("Crossplane Render", func() {
	var (
		// Create Sawchain instance with fake client
		sc = sawchain.New(GinkgoTB(), fake.NewClientBuilder().Build())

		// Read expectation YAMLs
		expectedXrYaml     = readYaml("expected-output-xr.yaml")
		expectedObjectYaml = readYaml("expected-output-object.yaml")
	)

	var _ = DescribeTable("rendering resources with function-go-templating",
		func(xrFileName, extraResourcesFileName, expectedConfigMapName string) {
			// Run crossplane render
			output, err := runCrossplaneRender(
				filepath.Join(yamlDir, xrFileName),
				filepath.Join(yamlDir, "composition.yaml"),
				filepath.Join(yamlDir, "functions.yaml"),
				filepath.Join(yamlDir, extraResourcesFileName),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())

			// Parse the render output into unstructured objects
			resources, err := unstructuredFromYaml(output)
			Expect(err).NotTo(HaveOccurred())

			// Check length of rendered resources
			Expect(resources).To(HaveLen(2))

			// Check rendered XR status (using Sawchain's MatchYAML matcher)
			Expect(resources).To(ContainElement(sc.MatchYAML(expectedXrYaml)))

			// Check rendered Object (using Sawchain's MatchYAML matcher)
			bindings := map[string]any{"expectedConfigMapName": expectedConfigMapName}
			Expect(resources).To(ContainElement(sc.MatchYAML(expectedObjectYaml, bindings)))
		},
		Entry("dev environment", "xr-dev.yaml", "extra-resources-dev.yaml", "my-awesome-dev-bucket-bucket"),
		Entry("prod environment", "xr-prod.yaml", "extra-resources-prod.yaml", "my-awesome-prod-bucket-bucket"),
	)
})
