package test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain"
)

var _ = Describe("Crossplane Render", func() {
	var (
		sc *sawchain.Sawchain

		yamlDir            = "yaml"
		xrPath             = filepath.Join(yamlDir, "xr.yaml")
		compositionPath    = filepath.Join(yamlDir, "composition.yaml")
		functionsPath      = filepath.Join(yamlDir, "functions.yaml")
		extraResourcesPath = filepath.Join(yamlDir, "extra-resources.yaml")
		expectedXRPath     = filepath.Join(yamlDir, "expected-output-xr.yaml")
		expectedObjectPath = filepath.Join(yamlDir, "expected-output-object.yaml")
	)

	BeforeEach(func() {
		// Create Sawchain instance with fake client
		fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
		sc = sawchain.New(GinkgoTB(), fakeClient)
	})

	It("should render resources correctly using function-go-templating", func() {
		// Read the expected outputs for comparison
		expectedXRBytes, err := os.ReadFile(expectedXRPath)
		Expect(err).NotTo(HaveOccurred())
		expectedXR := string(expectedXRBytes)

		expectedObjectBytes, err := os.ReadFile(expectedObjectPath)
		Expect(err).NotTo(HaveOccurred())
		expectedObject := string(expectedObjectBytes)

		// Run crossplane render
		output, err := runCrossplaneRender(xrPath, compositionPath, functionsPath, extraResourcesPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).NotTo(BeEmpty())

		// Parse the render output into unstructured objects
		renderedObjects, err := unstructuredObjectsFromYaml(output)
		Expect(err).NotTo(HaveOccurred())

		// We expect two rendered resources (the XR and the Object)
		Expect(renderedObjects).To(HaveLen(2), "Expected two rendered resources (XR and Object)")

		// XR should be first
		Expect(renderedObjects[0]).To(sc.MatchYAML(expectedXR), "XR should match expected output")

		// Object (ConfigMap) should be second
		Expect(renderedObjects[1]).To(sc.MatchYAML(expectedObject), "ConfigMap Object should match expected output")
	})
})
