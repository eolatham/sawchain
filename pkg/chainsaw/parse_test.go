package chainsaw

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: test case when template contains multiple resources (should fail)
var _ = DescribeTable("ParseResource",
	func(
		templateContent string,
		bindingsMap map[string]any,
		expectedObj client.Object,
		expectedErrs []string,
	) {
		// Create a temporary template file
		templatePath := filepath.Join(GinkgoT().TempDir(), "template.yaml")
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Test ParseResource
		obj, err := ParseResource(k8sClient, ctx, templatePath, bindingsMap)
		if len(expectedErrs) == 0 {
			Expect(err).NotTo(HaveOccurred())
			Expect(obj).To(Equal(expectedObj))
		} else {
			Expect(err).To(HaveOccurred())
			for _, substring := range expectedErrs {
				Expect(err.Error()).To(ContainSubstring(substring))
			}
		}
	},
	// Basic template
	Entry("should parse resource template",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
data:
  key: value
`,
		nil,
		&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string]string{
				"key": "value",
			},
		},
		nil,
	),
	// Template with bindings
	Entry("should parse template with bindings",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($name)
  namespace: default
data:
  key: ($value)
`,
		map[string]any{
			"name":  "test-bindings",
			"value": "bound-value",
		},
		&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-bindings",
				Namespace: "default",
			},
			Data: map[string]string{
				"key": "bound-value",
			},
		},
		nil,
	),
	// Invalid YAML
	Entry("should fail with invalid YAML",
		`
invalid: yaml: content
  - not: valid
    kubernetes: resource
`,
		nil,
		nil,
		[]string{
			"failed to load template file",
			"yaml: line 2: mapping values are not allowed in this context",
		},
	),
	// Missing required fields
	Entry("should fail with missing required fields",
		`
apiVersion: v1
metadata:
  name: test-missing-kind
`,
		nil,
		nil,
		[]string{
			"failed to load template file",
			"Object 'Kind' is missing",
		},
	),
	// Undefined binding
	Entry("should fail with undefined binding",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($undefined)
  namespace: default
`,
		map[string]any{
			"name": "test-invalid-binding",
		},
		nil,
		[]string{
			"variable not defined: $undefined",
		},
	),
	// Empty template
	Entry("should fail with empty template",
		"",
		nil,
		nil,
		[]string{
			"failed to load template file",
			"found no resource",
		},
	),
)
