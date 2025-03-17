package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain/internal/matchers"
)

// Helper function to create a ConfigMap
func createConfigMap(name, namespace string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

// Helper function to create an unstructured ConfigMap
func createUnstructuredConfigMap(name, namespace string, data map[string]string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("v1")
	obj.SetKind("ConfigMap")
	obj.SetName(name)
	obj.SetNamespace(namespace)

	dataMap := make(map[string]interface{})
	for k, v := range data {
		dataMap[k] = v
	}
	obj.Object["data"] = dataMap

	return obj
}

// Helper function to create a resource with status conditions
func createResourceWithConditions(apiVersion, kind, name, namespace string, conditions []metav1.Condition) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	status := map[string]interface{}{}
	conditionsData := make([]interface{}, len(conditions))

	for i, condition := range conditions {
		conditionMap := map[string]string{
			"type":               condition.Type,
			"status":             string(condition.Status),
			"reason":             condition.Reason,
			"message":            condition.Message,
			"lastTransitionTime": condition.LastTransitionTime.String(),
		}
		conditionsData[i] = conditionMap
	}

	status["conditions"] = conditionsData
	obj.Object["status"] = status

	return obj
}

var _ = Describe("Matchers", func() {
	var (
		fakeClient client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
	})

	Describe("NewChainsawMatcher", func() {
		type testCase struct {
			description     string
			actual          interface{}
			templateContent string
			bindings        map[string]any
			shouldMatch     bool
			expectedErr     string
		}

		DescribeTable("matching resources against templates",
			func(tc testCase) {
				matcher := matchers.NewChainsawMatcher(fakeClient, tc.templateContent, tc.bindings)

				match, err := matcher.Match(tc.actual)

				if tc.expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(match).To(Equal(tc.shouldMatch))
				}

				// Test failure messages
				if !tc.shouldMatch && tc.expectedErr == "" {
					Expect(matcher.FailureMessage(tc.actual)).To(ContainSubstring("to match template"))
				}

				if tc.shouldMatch && tc.expectedErr == "" {
					Expect(matcher.NegatedFailureMessage(tc.actual)).To(ContainSubstring("not to match template"))
				}
			},

			// Success cases with typed objects
			Entry("should match identical ConfigMap", testCase{
				description: "Exact match with typed ConfigMap",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
  key2: value2
`,
				bindings:    map[string]any{},
				shouldMatch: true,
			}),

			Entry("should match ConfigMap with subset of fields", testCase{
				description: "Partial match with typed ConfigMap",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
`,
				bindings:    map[string]any{},
				shouldMatch: true,
			}),

			Entry("should match ConfigMap with bindings", testCase{
				description: "Match with bindings in typed ConfigMap",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "bound-value",
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: ($value)
`,
				bindings: map[string]any{
					"value": "bound-value",
				},
				shouldMatch: true,
			}),

			// Success cases with unstructured objects
			Entry("should match unstructured ConfigMap", testCase{
				description: "Exact match with unstructured ConfigMap",
				actual: createUnstructuredConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
  key2: value2
`,
				bindings:    map[string]any{},
				shouldMatch: true,
			}),

			Entry("should match unstructured ConfigMap with bindings", testCase{
				description: "Match with bindings in unstructured ConfigMap",
				actual: createUnstructuredConfigMap("test-config", "default", map[string]string{
					"key1": "bound-value",
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: ($value)
`,
				bindings: map[string]any{
					"value": "bound-value",
				},
				shouldMatch: true,
			}),

			// Failure cases
			Entry("should not match when values differ", testCase{
				description: "Mismatch in values",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "wrong-value",
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: expected-value
`,
				bindings:    map[string]any{},
				shouldMatch: false,
			}),

			Entry("should not match when required fields missing", testCase{
				description: "Missing required fields",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key2": "value2",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
data:
  key1: value1
`,
				bindings:    map[string]any{},
				shouldMatch: false,
			}),

			// Edge cases
			Entry("should handle template with only metadata", testCase{
				description: "Template with only metadata",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: default
`,
				bindings:    map[string]any{},
				shouldMatch: true,
			}),

			// Error cases
			Entry("should error on nil input", testCase{
				description:     "Nil input",
				actual:          nil,
				templateContent: `apiVersion: v1\nkind: ConfigMap`,
				bindings:        map[string]any{},
				expectedErr:     "chainsawMatcher expects a client.Object but got nil",
			}),

			Entry("should error on non-object input", testCase{
				description:     "Non-object input",
				actual:          "not an object",
				templateContent: `apiVersion: v1\nkind: ConfigMap`,
				bindings:        map[string]any{},
				expectedErr:     "chainsawMatcher expects a client.Object but got string",
			}),

			Entry("should error on invalid template", testCase{
				description: "Invalid template",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
				}),
				templateContent: `invalid: yaml: content`,
				bindings:        map[string]any{},
				expectedErr:     "failed to parse template",
			}),

			// TODO: confirm how Chainsaw handles missing bindings in assertions
			// from the test failure, it seems that Chainsaw ignores the missing binding in this case
			FEntry("should error on missing binding", testCase{
				description: "Missing binding",
				actual: createConfigMap("test-config", "default", map[string]string{
					"key1": "value1",
				}),
				templateContent: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($missing)
`,
				bindings:    map[string]any{},
				expectedErr: "variable not defined: $missing",
			}),
		)
	})

	Describe("NewStatusConditionMatcher", func() {
		type testCase struct {
			description    string
			actual         interface{}
			conditionType  string
			expectedStatus string
			shouldMatch    bool
			expectedErr    string
		}

		DescribeTable("matching resources against status conditions",
			func(tc testCase) {
				matcher := matchers.NewStatusConditionMatcher(fakeClient, tc.conditionType, tc.expectedStatus)

				match, err := matcher.Match(tc.actual)

				if tc.expectedErr != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedErr))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(match).To(Equal(tc.shouldMatch))
				}

				// Test failure messages
				if !tc.shouldMatch && tc.expectedErr == "" {
					Expect(matcher.FailureMessage(tc.actual)).To(ContainSubstring("to match template"))
				}

				if tc.shouldMatch && tc.expectedErr == "" {
					Expect(matcher.NegatedFailureMessage(tc.actual)).To(ContainSubstring("not to match template"))
				}
			},

			// Success cases with unstructured objects
			// TODO: add a test case for an actual typed object
			Entry("should match when condition status is True", testCase{
				description: "Condition status is True",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    true,
			}),

			Entry("should match when condition status is False", testCase{
				description: "Condition status is False",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionFalse,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "False",
				shouldMatch:    true,
			}),

			Entry("should match when condition status is Unknown", testCase{
				description: "Condition status is Unknown",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionUnknown,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "Unknown",
				shouldMatch:    true,
			}),

			Entry("should match with multiple conditions", testCase{
				description: "Multiple conditions with match",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Available",
						Status: metav1.ConditionTrue,
					},
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
					{
						Type:   "Progressing",
						Status: metav1.ConditionFalse,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    true,
			}),

			// Failure cases
			Entry("should not match when condition status differs", testCase{
				description: "Condition status differs",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionFalse,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    false,
			}),

			Entry("should not match when condition type not found", testCase{
				description: "Condition type not found",
				actual: createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
					{
						Type:   "Available",
						Status: metav1.ConditionTrue,
					},
				}),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    false,
			}),

			// Edge cases
			Entry("should handle empty conditions array", testCase{
				description:    "Empty conditions array",
				actual:         createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{}),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    false,
			}),

			Entry("should handle missing status field", testCase{
				description: "Missing status field",
				actual: func() *unstructured.Unstructured {
					obj := &unstructured.Unstructured{}
					obj.SetAPIVersion("example.com/v1")
					obj.SetKind("TestResource")
					obj.SetName("test-resource")
					obj.SetNamespace("default")
					return obj
				}(),
				conditionType:  "Ready",
				expectedStatus: "True",
				shouldMatch:    false,
			}),

			// Error cases
			Entry("should error on nil input", testCase{
				description:    "Nil input",
				actual:         nil,
				conditionType:  "Ready",
				expectedStatus: "True",
				expectedErr:    "chainsawMatcher expects a client.Object but got nil",
			}),

			Entry("should error on non-object input", testCase{
				description:    "Non-object input",
				actual:         "not an object",
				conditionType:  "Ready",
				expectedStatus: "True",
				expectedErr:    "chainsawMatcher expects a client.Object but got string",
			}),
		)
	})

	// Test template generation for status condition matcher
	Describe("StatusConditionMatcher template generation", func() {
		It("should generate correct template for status condition", func() {
			// Create a resource with status conditions
			resource := createResourceWithConditions("example.com/v1", "TestResource", "test-resource", "default", []metav1.Condition{
				{
					Type:   "Ready",
					Status: metav1.ConditionTrue,
				},
			})

			// Register the GVK in the scheme
			gvk := schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "TestResource",
			}
			scheme.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})

			// Create a client with the updated scheme
			client := fake.NewClientBuilder().WithScheme(scheme).Build()

			// Create a matcher
			matcher := matchers.NewStatusConditionMatcher(client, "Ready", "True")

			// Match should succeed
			match, err := matcher.Match(resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(match).To(BeTrue())

			// We can't directly access the private fields, but we can check the failure message
			// which should contain the template content
			failureMessage := matcher.FailureMessage(resource)
			Expect(failureMessage).To(ContainSubstring("apiVersion: example.com/v1"))
			Expect(failureMessage).To(ContainSubstring("kind: TestResource"))
			Expect(failureMessage).To(ContainSubstring("status:"))
			Expect(failureMessage).To(ContainSubstring("conditions"))
		})
	})
})
