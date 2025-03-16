package chainsaw_test

import (
	"github.com/kyverno/chainsaw/pkg/apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	. "github.com/eolatham/sawchain/internal/chainsaw"
)

var _ = Describe("Chainsaw", func() {
	Describe("BindingsFromMap", func() {
		DescribeTable("converting maps to bindings",
			func(bindingsMap map[string]any) {
				// Test BindingsFromMap
				bindings := BindingsFromMap(bindingsMap)
				// Check bindings
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
	})

	Describe("ParseTemplate", func() {
		type testCase struct {
			templateContent string
			expectedObjs    []unstructured.Unstructured
			expectedErrs    []string
		}

		DescribeTable("parsing templates into unstructured objects",
			func(tc testCase) {
				// Test ParseTemplate
				objs, err := ParseTemplate(tc.templateContent)
				// Check error
				if len(tc.expectedErrs) > 0 {
					Expect(err).To(HaveOccurred())
					for _, expectedErr := range tc.expectedErrs {
						Expect(err.Error()).To(ContainSubstring(expectedErr))
					}
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
				// Check objects
				Expect(objs).To(ConsistOf(tc.expectedObjs))
			},
			// Single resource tests
			Entry("should parse a single ConfigMap", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
  key2: value2`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			Entry("should parse a single Secret", testCase{
				templateContent: `apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
type: Opaque
data:
  username: dXNlcm5hbWU=
  password: cGFzc3dvcmQ=`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name":      "test-secret",
								"namespace": "default",
							},
							"type": "Opaque",
							"data": map[string]interface{}{
								"username": "dXNlcm5hbWU=",
								"password": "cGFzc3dvcmQ=",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			// Multi-resource tests
			Entry("should parse multiple resources with different kinds", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
type: Opaque
data:
  username: dXNlcm5hbWU=`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key1": "value1",
							},
						},
					},
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name":      "test-secret",
								"namespace": "default",
							},
							"type": "Opaque",
							"data": map[string]interface{}{
								"username": "dXNlcm5hbWU=",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			Entry("should parse multiple resources of the same kind", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config-1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config-2
  namespace: default
data:
  key2: value2`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config-1",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key1": "value1",
							},
						},
					},
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config-2",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key2": "value2",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			Entry("should parse multiple resources with template expressions", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: ($binding1)
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
type: Opaque
data:
  password: ($binding2)`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key1": "($binding1)",
							},
						},
					},
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name":      "test-secret",
								"namespace": "default",
							},
							"type": "Opaque",
							"data": map[string]interface{}{
								"password": "($binding2)",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			Entry("should parse multiple resources with different namespaces", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: other-namespace
data:
  key1: value1`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key1": "value1",
							},
						},
					},
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "test-config",
								"namespace": "other-namespace",
							},
							"data": map[string]interface{}{
								"key1": "value1",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			// Edge cases
			Entry("should handle empty template", testCase{
				templateContent: ``,
				expectedObjs:    []unstructured.Unstructured{},
				expectedErrs:    nil,
			}),
			Entry("should handle template with only comments", testCase{
				templateContent: `# This is just a comment
# Another comment line`,
				expectedObjs: []unstructured.Unstructured{},
				expectedErrs: nil,
			}),
			Entry("should parse resource with missing required fields", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret`,
				expectedObjs: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "test-config",
							},
						},
					},
					{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Secret",
							"metadata": map[string]interface{}{
								"name": "test-secret",
							},
						},
					},
				},
				expectedErrs: nil,
			}),
			// Error cases
			Entry("should fail on empty documents", testCase{
				templateContent: `---
---
---`,
				expectedObjs: nil,
				expectedErrs: []string{
					"failed to parse template: Object 'Kind' is missing",
				},
			}),
			Entry("should fail on invalid YAML", testCase{
				templateContent: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
  key2: value2
 badindent: fail
---
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default`,
				expectedObjs: nil,
				expectedErrs: []string{
					"failed to parse template:",
					"did not find expected key",
				},
			}),
			Entry("should fail on non-YAML content", testCase{
				templateContent: `This is not YAML content at all.`,
				expectedObjs:    nil,
				expectedErrs: []string{
					"failed to parse template:",
					"cannot unmarshal",
				},
			}),
			Entry("should fail on malformed YAML with missing colon", testCase{
				templateContent: `apiVersion v1
kind: ConfigMap
metadata:
  name: test-config`,
				expectedObjs: nil,
				expectedErrs: []string{
					"failed to parse template:",
					"mapping values are not allowed in this context",
				},
			}),
		)
	})
})
