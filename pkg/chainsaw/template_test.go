package chainsaw

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kyverno/chainsaw/pkg/apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = DescribeTable("loadTemplateResource",
	func(
		templateContent string,
		expectedResource unstructured.Unstructured,
		expectedErrs []string,
	) {
		// Create a temporary template file
		templatePath := filepath.Join(GinkgoT().TempDir(), "template.yaml")
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Test loadTemplateResource
		resource, err := loadTemplateResource(templatePath)
		Expect(resource).To(Equal(expectedResource))
		if len(expectedErrs) == 0 {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			for _, substring := range expectedErrs {
				Expect(err.Error()).To(ContainSubstring(substring))
			}
		}
	},
	// Basic template
	Entry("should load single resource template",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
data:
  key: value
`,
		unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"data": map[string]interface{}{
					"key": "value",
				},
			},
		},
		nil,
	),
	// Empty template
	Entry("should fail with empty template",
		"",
		unstructured.Unstructured{},
		[]string{
			"failed to load template file",
			"found no resource",
		},
	),
	// Invalid YAML
	Entry("should fail with invalid YAML",
		`
invalid: yaml: content
  - not: valid
    kubernetes: resource
`,
		unstructured.Unstructured{},
		[]string{
			"failed to load template file",
			"yaml: line 2: mapping values are not allowed in this context",
		},
	),
	// Multiple resources
	Entry("should fail when template contains multiple resources",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-2
`,
		unstructured.Unstructured{},
		[]string{
			"expected template file",
			"to contain a single resource; found 2",
		},
	),
)

var _ = DescribeTable("bindingsFromMap",
	func(bindingsMap map[string]any) {
		// Test bindingsFromMap
		bindings := bindingsFromMap(context.Background(), bindingsMap)

		// Verify bindings
		if len(bindingsMap) == 0 {
			Expect(bindings).To(Equal(apis.NewBindings()))
		} else {
			for name, expectedValue := range bindingsMap {
				binding, err := bindings.Get("$" + name)
				Expect(err).NotTo(HaveOccurred(), "Expected binding %s not found", name)
				actualValue, err := binding.Value()
				Expect(err).NotTo(HaveOccurred(), "Failed to extract value for binding %s", name)
				if expectedValue == nil {
					Expect(actualValue).To(BeNil())
				} else {
					Expect(actualValue).To(Equal(expectedValue))
				}
			}
		}
	},
	// Empty map
	Entry("should handle empty map",
		map[string]any{},
	),
	// Single binding
	Entry("should convert single binding",
		map[string]any{
			"key": "value",
		},
	),
	// Multiple bindings
	Entry("should convert multiple bindings",
		map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	),
	// Different value types
	Entry("should handle different value types",
		map[string]any{
			"string": "text",
			"int":    123,
			"bool":   true,
			"float":  3.14,
			"slice":  []string{"a", "b"},
			"map":    map[string]string{"k": "v"},
			"nilVal": nil,
		},
	),
)
