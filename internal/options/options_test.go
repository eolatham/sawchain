package options_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/options"
)

var _ = Describe("Options", func() {
	Describe("ParseAndRequireGlobal", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring global options",
			func(tc testCase) {
				result, err := options.ParseAndRequireGlobal(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments
			Entry("valid timeout and interval as time.Duration", testCase{
				defaults: nil,
				args:     []interface{}{5 * time.Second, 1 * time.Second},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Bindings: map[string]any{},
				},
			}),

			Entry("valid timeout and interval as strings", testCase{
				defaults: nil,
				args:     []interface{}{"5s", "1s"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Bindings: map[string]any{},
				},
			}),

			Entry("valid timeout and interval with bindings", testCase{
				defaults: nil,
				args:     []interface{}{"5s", "1s", map[string]any{"key": "value"}},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Applying defaults
			Entry("use defaults when no args provided", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
			}),

			Entry("override default timeout only", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"5s"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
			}),

			Entry("override both default durations", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"5s", "1s"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
			}),

			// Merging bindings
			Entry("merge bindings with defaults", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{map[string]any{"new": "value", "shared": "override"}},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value", "new": "value", "shared": "override"},
				},
			}),

			Entry("merge multiple bindings arguments with default bindings", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value", "first": "value", "second": "value", "shared": "second"},
				},
			}),

			Entry("merge multiple bindings maps", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Bindings: map[string]any{"first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing timeout", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Timeout",
			}),

			Entry("missing interval", testCase{
				defaults:      nil,
				args:          []interface{}{"5s"},
				expectedError: "required argument(s) not provided: Interval",
			}),

			// Invalid arguments
			Entry("too many duration arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", "2s"},
				expectedError: "too many duration arguments provided",
			}),

			Entry("unexpected argument type", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", 5},
				expectedError: "unexpected argument type: int",
			}),

			// Disallowed arguments
			Entry("disallowed template argument", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", "template"},
				expectedError: "unexpected argument type: string",
			}),

			Entry("disallowed Object argument", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
						Data: map[string]string{
							"key": "value",
						},
					},
				},
				expectedError: "unexpected argument type: *v1.ConfigMap",
			}),

			Entry("disallowed Objects argument", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
				},
				expectedError: "unexpected argument type: []client.Object",
			}),
		)
	})

	Describe("ParseAndRequireImmediateSingle", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate single options",
			func(tc testCase) {
				result, err := options.ParseAndRequireImmediateSingle(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments - Template
			Entry("valid template content", testCase{
				defaults: nil,
				args:     []interface{}{"template content"},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template file", testCase{
				defaults: nil,
				args:     []interface{}{templateFilePath},
				expected: &options.Options{
					Template: templateFileContent,
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template with bindings", testCase{
				defaults: nil,
				args:     []interface{}{"template content", map[string]any{"key": "value"}},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Valid arguments - Object
			Entry("valid typed object", testCase{
				defaults: nil,
				args: []interface{}{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
				},
				expected: &options.Options{
					Object: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid unstructured object", testCase{
				defaults: nil,
				args: []interface{}{
					&unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "unstructured-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key": "value",
							},
						},
					},
				},
				expected: &options.Options{
					Object: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name":      "unstructured-config",
								"namespace": "default",
							},
							"data": map[string]interface{}{
								"key": "value",
							},
						},
					},
					Bindings: map[string]any{},
				},
			}),

			// Valid arguments - Template and Object together
			Entry("valid template and object together", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
				},
				expected: &options.Options{
					Template: "template content",
					Object: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, object, and bindings together", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
					map[string]any{"key": "value"},
				},
				expected: &options.Options{
					Template: "template content",
					Object: &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Default bindings
			Entry("use default bindings with template", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"template content"},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value"},
				},
			}),

			// Merging bindings
			Entry("merge bindings with defaults", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"new": "value", "shared": "override"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value", "new": "value", "shared": "override"},
				},
			}),

			Entry("merge multiple bindings maps", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing template or object", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string) or Object (client.Object)",
			}),

			// Invalid arguments
			Entry("multiple template arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"template1", "template2"},
				expectedError: "multiple template arguments provided",
			}),

			Entry("multiple object arguments", testCase{
				defaults: nil,
				args: []interface{}{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "config1",
							Namespace: "default",
						},
					},
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "config2",
							Namespace: "default",
						},
					},
				},
				expectedError: "multiple client.Object arguments provided",
			}),

			Entry("unexpected argument type", testCase{
				defaults:      nil,
				args:          []interface{}{5},
				expectedError: "unexpected argument type: int",
			}),

			// Disallowed arguments
			Entry("disallowed timeout argument", testCase{
				defaults:      nil,
				args:          []interface{}{5 * time.Second},
				expectedError: "unexpected argument type: time.Duration",
			}),

			Entry("disallowed objects argument", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "config1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "config2",
								Namespace: "default",
							},
						},
					},
				},
				expectedError: "unexpected argument type: []client.Object",
			}),
		)
	})

	Describe("ParseAndRequireImmediateMultiple", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate multiple options",
			func(tc testCase) {
				result, err := options.ParseAndRequireImmediateMultiple(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments - Template
			Entry("valid template content", testCase{
				defaults: nil,
				args:     []interface{}{"template content"},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template file", testCase{
				defaults: nil,
				args:     []interface{}{templateFilePath},
				expected: &options.Options{
					Template: templateFileContent,
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template with bindings", testCase{
				defaults: nil,
				args:     []interface{}{"template content", map[string]any{"key": "value"}},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Valid arguments - Objects
			Entry("valid objects slice", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid objects slice with mixed types", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config",
								Namespace: "default",
							},
						},
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]interface{}{
									"name":      "test-secret",
									"namespace": "default",
								},
							},
						},
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config",
								Namespace: "default",
							},
						},
						&unstructured.Unstructured{
							Object: map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]interface{}{
									"name":      "test-secret",
									"namespace": "default",
								},
							},
						},
					},
					Bindings: map[string]any{},
				},
			}),

			// Valid arguments - Template and Objects together
			Entry("valid template and objects together", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
				},
				expected: &options.Options{
					Template: "template content",
					Objects: []client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, objects, and bindings together", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
					map[string]any{"key": "value"},
				},
				expected: &options.Options{
					Template: "template content",
					Objects: []client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-1",
								Namespace: "default",
							},
						},
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-config-2",
								Namespace: "default",
							},
						},
					},
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Default bindings
			Entry("use default bindings with template", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"template content"},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value"},
				},
			}),

			// Merging bindings
			Entry("merge bindings with defaults", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"new": "value", "shared": "override"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value", "new": "value", "shared": "override"},
				},
			}),

			Entry("merge multiple bindings maps", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing template or objects", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string) or Objects ([]client.Object)",
			}),

			// Invalid arguments
			Entry("multiple template arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"template1", "template2"},
				expectedError: "multiple template arguments provided",
			}),

			Entry("multiple objects arguments", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "config1",
								Namespace: "default",
							},
						},
					},
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "config2",
								Namespace: "default",
							},
						},
					},
				},
				expectedError: "multiple []client.Object arguments provided",
			}),

			Entry("unexpected argument type", testCase{
				defaults:      nil,
				args:          []interface{}{5},
				expectedError: "unexpected argument type: int",
			}),

			// Disallowed arguments
			Entry("disallowed timeout argument", testCase{
				defaults:      nil,
				args:          []interface{}{5 * time.Second},
				expectedError: "unexpected argument type: time.Duration",
			}),

			Entry("disallowed single object argument", testCase{
				defaults: nil,
				args: []interface{}{
					&corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-config",
							Namespace: "default",
						},
					},
				},
				expectedError: "unexpected argument type: *v1.ConfigMap",
			}),
		)
	})
	// TODO: test ParseAndRequireEventualSingle
	// TODO: test ParseAndRequireEventualMultiple
})
