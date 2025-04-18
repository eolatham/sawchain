package options_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/options"
	"github.com/eolatham/sawchain/internal/testutil"
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

			// Using defaults
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

			Entry("merge multiple bindings maps with defaults", testCase{
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
			Entry("interval greater than timeout", testCase{
				defaults:      nil,
				args:          []interface{}{"1s", "5s"},
				expectedError: "provided interval is greater than timeout",
			}),

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

			Entry("disallowed object argument", testCase{
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

			Entry("disallowed objects argument", testCase{
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

	Describe("ParseAndRequireEventual", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring eventual operation options",
			func(tc testCase) {
				result, err := options.ParseAndRequireEventual(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments - template
			Entry("valid durations and template content", testCase{
				defaults: nil,
				args:     []interface{}{"5s", "1s", "template content"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations and template file", testCase{
				defaults: nil,
				args:     []interface{}{"5s", "1s", templateFilePath},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: templateFileContent,
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations, template, and bindings", testCase{
				defaults: nil,
				args:     []interface{}{"5s", "1s", "template content", map[string]any{"key": "value"}},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Valid arguments - single object
			Entry("valid durations and typed object", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Object:   testutil.NewConfigMap("test-config", "default", nil),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations and unstructured object", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Object:   testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations, template, bindings, and object", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					"template content",
					map[string]any{"key": "value"},
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Object:   testutil.NewConfigMap("test-config", "default", nil),
				},
			}),

			// Valid arguments - multiple objects
			Entry("valid durations and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations and objects with mixed types", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					[]client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid durations, template, bindings, and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					"template content",
					map[string]any{"key": "value"},
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
			}),

			// Using defaults
			Entry("use defaults for durations", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"template content"},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"default": "value"},
				},
			}),

			Entry("override default timeout only", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"5s", "template content"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 2 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"default": "value"},
				},
			}),

			Entry("override both default durations", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value"},
				},
				args: []interface{}{"5s", "1s", "template content"},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
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
				args: []interface{}{
					"template content",
					map[string]any{"new": "value", "shared": "override"},
				},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"default": "value", "new": "value", "shared": "override"},
				},
			}),

			Entry("merge multiple bindings maps", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Timeout:  5 * time.Second,
					Interval: 1 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"first": "value", "second": "value", "shared": "second"},
				},
			}),

			Entry("merge multiple bindings maps with defaults", testCase{
				defaults: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Timeout:  10 * time.Second,
					Interval: 2 * time.Second,
					Template: "template content",
					Bindings: map[string]any{"default": "value", "first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing timeout", testCase{
				defaults:      nil,
				args:          []interface{}{"template content"},
				expectedError: "required argument(s) not provided: Timeout",
			}),

			Entry("missing interval", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "template content"},
				expectedError: "required argument(s) not provided: Interval",
			}),

			Entry("missing template/object/objects", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s"},
				expectedError: "required argument(s) not provided: Template (string), Object (client.Object), or Objects ([]client.Object)",
			}),

			// Invalid arguments
			Entry("interval greater than timeout", testCase{
				defaults:      nil,
				args:          []interface{}{"1s", "5s", "template content"},
				expectedError: "provided interval is greater than timeout",
			}),

			Entry("too many duration arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", "2s", "template content"},
				expectedError: "too many duration arguments provided",
			}),

			Entry("multiple template arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", "template1", "template2"},
				expectedError: "multiple template arguments provided",
			}),

			Entry("nil object", testCase{
				defaults: nil,
				args: []interface{}{
					(*corev1.ConfigMap)(nil),
				},
				expectedError: "provided client.Object is nil or has a nil underlying value",
			}),

			Entry("objects containing nil object", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-config",
								Namespace: "default",
							},
						},
						(*corev1.ConfigMap)(nil),
					},
				},
				expectedError: "provided []client.Object contains an element that is nil or has a nil underlying value",
			}),

			Entry("object and objects together", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					testutil.NewConfigMap("single-config", "default", nil),
					[]client.Object{
						testutil.NewConfigMap("multi-config", "default", nil),
					},
				},
				expectedError: "client.Object and []client.Object arguments both provided",
			}),

			Entry("multiple object arguments", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					testutil.NewConfigMap("config1", "default", nil),
					testutil.NewConfigMap("config2", "default", nil),
				},
				expectedError: "multiple client.Object arguments provided",
			}),

			Entry("multiple objects arguments", testCase{
				defaults: nil,
				args: []interface{}{
					"5s",
					"1s",
					[]client.Object{
						testutil.NewConfigMap("config1", "default", nil),
					},
					[]client.Object{
						testutil.NewConfigMap("config2", "default", nil),
					},
				},
				expectedError: "multiple []client.Object arguments provided",
			}),

			Entry("unexpected argument type", testCase{
				defaults:      nil,
				args:          []interface{}{"5s", "1s", 5},
				expectedError: "unexpected argument type: int",
			}),
		)
	})

	Describe("ParseAndRequireImmediate", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate operation options",
			func(tc testCase) {
				result, err := options.ParseAndRequireImmediate(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments - template
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

			Entry("valid template and bindings", testCase{
				defaults: nil,
				args:     []interface{}{"template content", map[string]any{"key": "value"}},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
				},
			}),

			// Valid arguments - single object
			Entry("valid typed object", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Object:   testutil.NewConfigMap("test-config", "default", nil),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid unstructured object", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
				},
				expected: &options.Options{
					Object:   testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, bindings, and object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Object:   testutil.NewConfigMap("test-config", "default", nil),
				},
			}),

			// Valid arguments - multiple objects
			Entry("valid objects", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid objects with mixed types", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, bindings, and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
			}),

			// Using defaults
			Entry("use default bindings", testCase{
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

			Entry("merge multiple bindings maps with defaults", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value", "first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing template/object/objects", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string), Object (client.Object), or Objects ([]client.Object)",
			}),

			// Invalid arguments
			Entry("nil object", testCase{
				defaults: nil,
				args: []interface{}{
					(*corev1.ConfigMap)(nil),
				},
				expectedError: "provided client.Object is nil or has a nil underlying value",
			}),

			Entry("objects containing nil object", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-config",
								Namespace: "default",
							},
						},
						(*corev1.ConfigMap)(nil),
					},
				},
				expectedError: "provided []client.Object contains an element that is nil or has a nil underlying value",
			}),

			Entry("multiple template arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"template1", "template2"},
				expectedError: "multiple template arguments provided",
			}),

			Entry("object and objects together", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewConfigMap("single-config", "default", nil),
					[]client.Object{
						testutil.NewConfigMap("multi-config", "default", nil),
					},
				},
				expectedError: "client.Object and []client.Object arguments both provided",
			}),

			Entry("multiple object arguments", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewConfigMap("config1", "default", nil),
					testutil.NewConfigMap("config2", "default", nil),
				},
				expectedError: "multiple client.Object arguments provided",
			}),

			Entry("multiple objects arguments", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						testutil.NewConfigMap("config1", "default", nil),
					},
					[]client.Object{
						testutil.NewConfigMap("config2", "default", nil),
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
		)
	})

	Describe("ParseAndRequireImmediateSingle", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate single-resource operation options",
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

			// Valid arguments
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

			Entry("valid typed object", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Object:   testutil.NewConfigMap("test-config", "default", nil),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid unstructured object", testCase{
				defaults: nil,
				args: []interface{}{
					testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
				},
				expected: &options.Options{
					Object:   testutil.NewUnstructuredConfigMap("unstructured-config", "default", map[string]string{"key": "value"}),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template and object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Template: "template content",
					Object:   testutil.NewConfigMap("test-config", "default", nil),
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, bindings, and object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Object:   testutil.NewConfigMap("test-config", "default", nil),
				},
			}),

			// Using defaults
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

			Entry("merge multiple bindings maps with defaults", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value", "first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing template or object", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string) or Object (client.Object)",
			}),

			// Invalid arguments
			Entry("nil object", testCase{
				defaults: nil,
				args: []interface{}{
					(*corev1.ConfigMap)(nil),
				},
				expectedError: "provided client.Object is nil or has a nil underlying value",
			}),

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

	Describe("ParseAndRequireImmediateMulti", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate multi-resource operation options",
			func(tc testCase) {
				result, err := options.ParseAndRequireImmediateMulti(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments
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

			Entry("valid objects", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid objects with mixed types", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
				},
				expected: &options.Options{
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-typed", "default", nil),
						testutil.NewUnstructuredConfigMap("test-config-unstructured", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Template: "template content",
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
					Bindings: map[string]any{},
				},
			}),

			Entry("valid template, bindings, and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
			}),

			// Using defaults
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

			Entry("merge multiple bindings maps with defaults", testCase{
				defaults: &options.Options{
					Bindings: map[string]any{"default": "value", "shared": "default"},
				},
				args: []interface{}{
					"template content",
					map[string]any{"first": "value", "shared": "first"},
					map[string]any{"second": "value", "shared": "second"},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"default": "value", "first": "value", "second": "value", "shared": "second"},
				},
			}),

			// Missing required arguments
			Entry("missing template or objects", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string) or Objects ([]client.Object)",
			}),

			// Invalid arguments
			Entry("objects containing nil object", testCase{
				defaults: nil,
				args: []interface{}{
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-config",
								Namespace: "default",
							},
						},
						(*corev1.ConfigMap)(nil),
					},
				},
				expectedError: "provided []client.Object contains an element that is nil or has a nil underlying value",
			}),

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

			Entry("disallowed object argument", testCase{
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

	Describe("ParseAndRequireImmediateTemplate", func() {
		type testCase struct {
			defaults      *options.Options
			args          []interface{}
			expected      *options.Options
			expectedError string
		}

		DescribeTable("parsing and requiring immediate template operation options",
			func(tc testCase) {
				result, err := options.ParseAndRequireImmediateTemplate(tc.defaults, tc.args...)
				if tc.expectedError != "" {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tc.expectedError))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expected))
				}
			},

			// Valid arguments
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

			Entry("valid template, bindings, and object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					testutil.NewConfigMap("test-config", "default", nil),
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Object:   testutil.NewConfigMap("test-config", "default", nil),
				},
			}),

			Entry("valid template, bindings, and objects", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					map[string]any{"key": "value"},
					[]client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
				expected: &options.Options{
					Template: "template content",
					Bindings: map[string]any{"key": "value"},
					Objects: []client.Object{
						testutil.NewConfigMap("test-config-1", "default", nil),
						testutil.NewConfigMap("test-config-2", "default", nil),
					},
				},
			}),

			// Using defaults
			Entry("use default bindings", testCase{
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
			Entry("missing template", testCase{
				defaults:      nil,
				args:          []interface{}{},
				expectedError: "required argument(s) not provided: Template (string)",
			}),

			// Invalid arguments
			Entry("nil object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					(*corev1.ConfigMap)(nil),
				},
				expectedError: "provided client.Object is nil or has a nil underlying value",
			}),

			Entry("objects containing nil object", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					[]client.Object{
						&corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-config",
								Namespace: "default",
							},
						},
						(*corev1.ConfigMap)(nil),
					},
				},
				expectedError: "provided []client.Object contains an element that is nil or has a nil underlying value",
			}),

			Entry("multiple template arguments", testCase{
				defaults:      nil,
				args:          []interface{}{"template1", "template2"},
				expectedError: "multiple template arguments provided",
			}),

			Entry("object and objects together", testCase{
				defaults: nil,
				args: []interface{}{
					"template content",
					testutil.NewConfigMap("single-config", "default", nil),
					[]client.Object{
						testutil.NewConfigMap("multi-config", "default", nil),
					},
				},
				expectedError: "client.Object and []client.Object arguments both provided",
			}),

			Entry("unexpected argument type", testCase{
				defaults:      nil,
				args:          []interface{}{"template content", 5},
				expectedError: "unexpected argument type: int",
			}),

			// Disallowed arguments
			Entry("disallowed timeout argument", testCase{
				defaults:      nil,
				args:          []interface{}{5 * time.Second},
				expectedError: "unexpected argument type: time.Duration",
			}),
		)
	})
})
