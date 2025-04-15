package example

import (
	"path/filepath"
	"strings"

	"github.com/eolatham/sawchain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	compositionPath = "yaml/composition.yaml"
	functionsPath   = "yaml/functions.yaml"

	xrTemplatePath             = "yaml/xr.tpl.yaml"
	extraResourcesTemplatePath = "yaml/extra-resources.tpl.yaml"
	expectedOutputTemplatePath = "yaml/expected-output.tpl.yaml"
)

var _ = Describe("Crossplane Render", func() {
	DescribeTable("rendering resources with function-go-templating",
		func(environment string, expectedErrs []string) {
			// Create Sawchain with fake client and global bindings
			sc := sawchain.New(GinkgoTB(), fake.NewClientBuilder().Build(), map[string]any{"environment": environment})

			// Render input template files
			xrPath := filepath.Join(GinkgoT().TempDir(), "xr.yaml")
			sc.RenderToFile(xrPath, xrTemplatePath)

			extraResourcesPath := filepath.Join(GinkgoT().TempDir(), "extra-resources.yaml")
			sc.RenderToFile(extraResourcesPath, extraResourcesTemplatePath)

			// Render expected output
			expectedOutput := sc.RenderToString(expectedOutputTemplatePath)

			// Run crossplane render
			output, err := runCrossplaneRender(xrPath, compositionPath, functionsPath, extraResourcesPath)
			if len(expectedErrs) > 0 {
				// Verify error
				Expect(err).To(HaveOccurred())
				for _, expectedErr := range expectedErrs {
					Expect(err.Error()).To(ContainSubstring(expectedErr))
				}
			} else {
				// Verify no error
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeEmpty())

				// Parse render output into unstructured objects
				resources, err := unstructuredFromYaml(output)
				Expect(err).NotTo(HaveOccurred())

				// Verify rendered resource count
				Expect(resources).To(HaveLen(2))

				// Verify rendered resource fields
				for _, document := range strings.Split(expectedOutput, "---") {
					Expect(resources).To(ContainElement(sc.MatchYAML(document)))
				}
			}
		},
		Entry("dev environment", "dev", nil),
		Entry("prod environment", "prod", nil),
		Entry("invalid environment", "yaml: bad", []string{
			"error: cannot render composite resource: pipeline step \"render-templates\" returned a fatal result",
			"cannot decode manifest: error converting YAML to JSON: yaml: line 8: mapping values are not allowed in this context",
		}),
	)
})
