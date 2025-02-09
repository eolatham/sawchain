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

var _ = DescribeTable("ParseResources",
	func(
		templateContent string,
		bindingsMap map[string]any,
		expectedResources []client.Object,
		expectedErrors []string,
	) {
		// Create a temporary template file
		templatePath := filepath.Join(GinkgoT().TempDir(), "template.yaml")
		err := os.WriteFile(templatePath, []byte(templateContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Test ParseResources
		resources, err := ParseResources(k8sClient, ctx, templatePath, bindingsMap)
		if len(expectedErrors) == 0 {
			Expect(err).NotTo(HaveOccurred())
			Expect(resources).To(HaveLen(len(expectedResources)))
			for i, resource := range resources {
				Expect(resource).To(Equal(expectedResources[i]))
			}
		} else {
			Expect(err).To(HaveOccurred())
			for _, substring := range expectedErrors {
				Expect(err.Error()).To(ContainSubstring(substring))
			}
		}
	},
	// Single resource template
	Entry("should parse single resource template",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-single
  namespace: default
data:
  key: value
`,
		nil,
		[]client.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-single",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			},
		},
		nil,
	),
	// Multiple resources template
	Entry("should parse multiple resources template",
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-multi-1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: Secret
metadata:
  name: test-multi-2
  namespace: default
data:
  key2: dmFsdWUy
`,
		nil,
		[]client.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multi-1",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "value1",
				},
			},
			&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multi-2",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"key2": []byte("value2"),
				},
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
		[]client.Object{
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
