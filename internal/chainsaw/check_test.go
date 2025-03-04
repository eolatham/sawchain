package chainsaw

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = DescribeTableSubtree("CheckResource", Ordered,
	func(
		objs []client.Object,
		templateContent string,
		bindingsMap map[string]any,
		expectedErrs []string,
	) {
		BeforeAll(func() {
			// Create the test objects in the cluster
			for _, obj := range objs {
				err := k8sClient.Create(ctx, obj)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should check resources", func() {
			match, err := CheckResource(k8sClient, ctx, templateContent, bindingsMap)
			if len(expectedErrs) == 0 {
				Expect(err).NotTo(HaveOccurred())
				Expect(match).NotTo(BeNil())
				// Clear match GVK because created objects have empty GVKs
				match.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
				// Assert match is one of the created objects
				Expect(objs).To(ContainElement(match))
			} else {
				Expect(err).To(HaveOccurred())
				Expect(match).To(BeNil())
				for _, substring := range expectedErrs {
					Expect(err.Error()).To(ContainSubstring(substring))
				}
			}
		})

		AfterAll(func() {
			// Delete the test objects from the cluster
			for _, obj := range objs {
				err := k8sClient.Delete(ctx, obj)
				Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
			}
		})
	},
	// Single candidate, match
	Entry("should successfully check single matching resource",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-single",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			},
		},
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
		nil,
	),
	// Single candidate, no match
	Entry("should fail when single resource doesn't match",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-single-nomatch",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "actual-value",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-single-nomatch
  namespace: default
data:
  key: expected-value
`,
		nil,
		[]string{
			"failed to execute check",
			"data.key: Invalid value: \"actual-value\": Expected value: \"expected-value\"",
		},
	),
	// Multiple candidates with match (using label selector)
	Entry("should successfully check when one of multiple resources matches",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multiple-1",
					Namespace: "default",
					Labels: map[string]string{
						"test": "multiple",
					},
				},
				Data: map[string]string{
					"key": "different",
				},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multiple-2",
					Namespace: "default",
					Labels: map[string]string{
						"test": "multiple",
					},
				},
				Data: map[string]string{
					"key": "value",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: default
  labels:
    test: multiple
data:
  key: value
`,
		nil,
		nil,
	),
	// Multiple candidates with no match
	Entry("should fail when none of multiple resources match",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multiple-nomatch-1",
					Namespace: "default",
					Labels: map[string]string{
						"test": "multiple-nomatch",
					},
				},
				Data: map[string]string{
					"key": "actual-value-1",
				},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multiple-nomatch-2",
					Namespace: "default",
					Labels: map[string]string{
						"test": "multiple-nomatch",
					},
				},
				Data: map[string]string{
					"key": "actual-value-2",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: default
  labels:
    test: multiple-nomatch
data:
  key: expected-value
`,
		nil,
		[]string{
			"failed to execute check",
			"data.key: Invalid value: \"actual-value-1\": Expected value: \"expected-value\"",
			"data.key: Invalid value: \"actual-value-2\": Expected value: \"expected-value\"",
		},
	),
	// No candidates
	Entry("should fail when no resources exist",
		[]client.Object{},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-nonexistent
  namespace: default
data:
  key: value
`,
		nil,
		[]string{
			"failed to execute check: actual resource not found",
		},
	),
	// Invalid template
	Entry("should fail with invalid template",
		[]client.Object{},
		`
invalid: yaml: content
  - not: valid
    kubernetes: resource
`,
		nil,
		[]string{
			"failed to parse template",
			"yaml: line 2: mapping values are not allowed in this context",
		},
	),
	// Template with bindings
	Entry("should successfully check resource with bindings",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-bindings",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "bound-value",
				},
			},
		},
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
		nil,
	),
	// Template with invalid bindings
	Entry("should fail with undefined binding",
		[]client.Object{},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  name: ($undefined)
  namespace: default
data:
  key: value
`,
		map[string]any{
			"name": "test-invalid-binding",
		},
		[]string{
			"failed to execute check",
			"variable not defined: $undefined",
		},
	),
	// Template with no resources
	Entry("should fail with empty template",
		[]client.Object{},
		"",
		nil,
		[]string{
			"expected template to contain a single resource; found 0",
		},
	),
	// Template with namespace selector
	Entry("should match resource in specific namespace",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ns",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: default
data:
  key: value
`,
		nil,
		nil,
	),
	// Template with JMESPath expression that checks the length of a field
	Entry("should check field length using JMESPath",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-jmespath-happy",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test",
					},
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: default
  labels:
    (length(app)): 4
`,
		nil,
		nil,
	),
	// Template with JMESPath expression that checks the length of a field
	Entry("should fail when field length is not equal to expected",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-jmespath-sad",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test",
					},
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: default
  labels:
    (length(app)): 100
`,
		nil,
		[]string{
			"failed to execute check",
			"metadata.labels.(length(app)): Invalid value: 4: Expected value: 100",
		},
	),
	// Template with kind selector
	Entry("should match resource by kind only",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-kind-1",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value1",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
data:
  key: value1
`,
		nil,
		nil,
	),
	// Template with kind selector that doesn't match any resources
	Entry("should fail when no resources match kind selector",
		[]client.Object{
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
			},
		},
		`
apiVersion: v1
kind: ConfigMap
data:
  key: value1
`,
		nil,
		[]string{
			"failed to execute check",
		},
	),
	// Template with multiple resources
	Entry("should fail when template contains multiple resources",
		[]client.Object{
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multi-1",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "value1",
				},
			},
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-multi-2",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"key2": []byte("value2"),
				},
			},
		},
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
		[]string{
			"expected template to contain a single resource; found 2",
		},
	),
)
