package utilities_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/eolatham/sawchain/internal/utilities"
)

var _ = Describe("Utilities", func() {
	Describe("MergeMaps", func() {
		type testCase struct {
			maps     []map[string]any
			expected map[string]any
		}

		DescribeTable("merging maps",
			func(tc testCase) {
				result := utilities.MergeMaps(tc.maps...)
				Expect(result).To(Equal(tc.expected))
			},
			Entry("no maps provided", testCase{
				maps:     []map[string]any{},
				expected: map[string]any{},
			}),
			Entry("single map provided", testCase{
				maps:     []map[string]any{{"key1": "value1", "key2": 2}},
				expected: map[string]any{"key1": "value1", "key2": 2},
			}),
			Entry("two maps with non-overlapping keys", testCase{
				maps: []map[string]any{
					{"key1": "value1", "key2": 2},
					{"key3": true, "key4": 4.5},
				},
				expected: map[string]any{"key1": "value1", "key2": 2, "key3": true, "key4": 4.5},
			}),
			Entry("two maps with overlapping keys", testCase{
				maps: []map[string]any{
					{"key1": "original", "key2": 2, "key3": "value3"},
					{"key1": "override", "key4": 4},
				},
				expected: map[string]any{"key1": "override", "key2": 2, "key3": "value3", "key4": 4},
			}),
			Entry("multiple maps with overlapping keys", testCase{
				maps: []map[string]any{
					{"key1": "first", "key2": 2},
					{"key1": "second", "key3": true},
					{"key1": "third", "key4": 4.5},
				},
				expected: map[string]any{"key1": "third", "key2": 2, "key3": true, "key4": 4.5},
			}),
			Entry("maps with nested structures", testCase{
				maps: []map[string]any{
					{
						"key1": "value1",
						"nested": map[string]any{
							"a": 1,
							"b": 2,
						},
					},
					{
						"key2": "value2",
						"nested": map[string]any{
							"c": 3,
							"d": 4,
						},
					},
				},
				expected: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"nested": map[string]any{
						"c": 3,
						"d": 4,
					},
				},
			}),
			Entry("nil maps are skipped", testCase{
				maps: []map[string]any{
					{"key1": "value1"},
					nil,
					{"key2": "value2"},
				},
				expected: map[string]any{"key1": "value1", "key2": "value2"},
			}),
			Entry("all nil maps", testCase{
				maps: []map[string]any{
					nil,
					nil,
				},
				expected: map[string]any{},
			}),
		)
	})

	Describe("IsExistingFile", func() {
		type testCase struct {
			setup    func() string
			expected bool
		}

		DescribeTable("checking if a file exists",
			func(tc testCase) {
				path := tc.setup()
				result := utilities.IsExistingFile(path)
				Expect(result).To(Equal(tc.expected))
			},
			Entry("path points to an existing file", testCase{
				setup: func() string {
					filePath := filepath.Join(tempDir, "test-file.txt")
					err := os.WriteFile(filePath, []byte("test content"), 0644)
					Expect(err).NotTo(HaveOccurred())
					return filePath
				},
				expected: true,
			}),
			Entry("path points to a directory", testCase{
				setup: func() string {
					return tempDir
				},
				expected: false,
			}),
			Entry("path does not exist", testCase{
				setup: func() string {
					return filepath.Join(tempDir, "non-existent-file.txt")
				},
				expected: false,
			}),
			Entry("path is empty", testCase{
				setup: func() string {
					return ""
				},
				expected: false,
			}),
		)
	})

	Describe("ReadFileContent", func() {
		type testCase struct {
			setup          func() string
			expectedResult string
			expectError    bool
		}

		DescribeTable("reading file content",
			func(tc testCase) {
				path := tc.setup()
				result, err := utilities.ReadFileContent(path)

				if tc.expectError {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(tc.expectedResult))
				}
			},
			Entry("reading content from an existing file", testCase{
				setup: func() string {
					content := "This is test content.\nWith multiple lines."
					filePath := filepath.Join(tempDir, "test-file.txt")
					err := os.WriteFile(filePath, []byte(content), 0644)
					Expect(err).NotTo(HaveOccurred())
					return filePath
				},
				expectedResult: "This is test content.\nWith multiple lines.",
				expectError:    false,
			}),
			Entry("reading content from an empty file", testCase{
				setup: func() string {
					filePath := filepath.Join(tempDir, "empty-file.txt")
					err := os.WriteFile(filePath, []byte(""), 0644)
					Expect(err).NotTo(HaveOccurred())
					return filePath
				},
				expectedResult: "",
				expectError:    false,
			}),
			Entry("reading content from a non-existent file", testCase{
				setup: func() string {
					return filepath.Join(tempDir, "non-existent-file.txt")
				},
				expectedResult: "",
				expectError:    true,
			}),
			Entry("reading content from a directory", testCase{
				setup: func() string {
					return tempDir
				},
				expectedResult: "",
				expectError:    true,
			}),
		)
	})

	Describe("AsDuration", func() {
		type testCase struct {
			input          interface{}
			expectedResult time.Duration
			expectedOk     bool
		}

		DescribeTable("converting to duration",
			func(tc testCase) {
				result, ok := utilities.AsDuration(tc.input)
				Expect(ok).To(Equal(tc.expectedOk))
				if tc.expectedOk {
					Expect(result).To(Equal(tc.expectedResult))
				}
			},
			Entry("input is already a time.Duration", testCase{
				input:          5 * time.Second,
				expectedResult: 5 * time.Second,
				expectedOk:     true,
			}),
			Entry("input is a valid duration string", testCase{
				input:          "10m30s",
				expectedResult: 10*time.Minute + 30*time.Second,
				expectedOk:     true,
			}),
			Entry("input is an invalid duration string", testCase{
				input:          "invalid",
				expectedResult: 0,
				expectedOk:     false,
			}),
			Entry("input is an integer", testCase{
				input:          42,
				expectedResult: 0,
				expectedOk:     false,
			}),
			Entry("input is nil", testCase{
				input:          nil,
				expectedResult: 0,
				expectedOk:     false,
			}),
		)
	})

	Describe("AsMapStringAny", func() {
		type testCase struct {
			input          interface{}
			expectedResult map[string]any
			expectedOk     bool
		}

		DescribeTable("converting to map[string]any",
			func(tc testCase) {
				result, ok := utilities.AsMapStringAny(tc.input)
				Expect(ok).To(Equal(tc.expectedOk))
				if tc.expectedOk {
					Expect(result).To(Equal(tc.expectedResult))
				}
			},
			Entry("input is already a map[string]any", testCase{
				input:          map[string]any{"key1": "value1", "key2": 42},
				expectedResult: map[string]any{"key1": "value1", "key2": 42},
				expectedOk:     true,
			}),
			Entry("input is a map[string]string", testCase{
				input:          map[string]string{"key1": "value1", "key2": "value2"},
				expectedResult: map[string]any{"key1": "value1", "key2": "value2"},
				expectedOk:     true,
			}),
			Entry("input is a map[string]int", testCase{
				input:          map[string]int{"key1": 1, "key2": 2},
				expectedResult: map[string]any{"key1": 1, "key2": 2},
				expectedOk:     true,
			}),
			Entry("input is a map with non-string keys", testCase{
				input:          map[int]string{1: "value1", 2: "value2"},
				expectedResult: nil,
				expectedOk:     false,
			}),
			Entry("input is not a map", testCase{
				input:          "not a map",
				expectedResult: nil,
				expectedOk:     false,
			}),
			Entry("input is nil", testCase{
				input:          nil,
				expectedResult: nil,
				expectedOk:     false,
			}),
		)
	})

	Describe("AsObject", func() {
		type testCase struct {
			input      interface{}
			expectedOk bool
		}

		DescribeTable("converting to client.Object",
			func(tc testCase) {
				_, ok := utilities.AsObject(tc.input)
				Expect(ok).To(Equal(tc.expectedOk))
			},
			Entry("input implements client.Object", testCase{
				input:      &corev1.ConfigMap{},
				expectedOk: true,
			}),
			Entry("input does not implement client.Object", testCase{
				input:      "not an object",
				expectedOk: false,
			}),
			Entry("input is nil", testCase{
				input:      nil,
				expectedOk: false,
			}),
		)
	})

	Describe("AsSliceOfObjects", func() {
		type testCase struct {
			input          interface{}
			expectedOk     bool
			expectedLength int
		}

		DescribeTable("converting to []client.Object",
			func(tc testCase) {
				result, ok := utilities.AsSliceOfObjects(tc.input)
				Expect(ok).To(Equal(tc.expectedOk))
				if tc.expectedOk {
					Expect(result).To(HaveLen(tc.expectedLength))
				}
			},
			Entry("input is a slice of client.Object", testCase{
				input:          []client.Object{&corev1.ConfigMap{}, &corev1.ConfigMap{}},
				expectedOk:     true,
				expectedLength: 2,
			}),
			Entry("input is an empty slice of client.Object", testCase{
				input:          []client.Object{},
				expectedOk:     true,
				expectedLength: 0,
			}),
			Entry("input is a slice of interfaces that implement client.Object", testCase{
				input:          []interface{}{&corev1.ConfigMap{}, &corev1.ConfigMap{}},
				expectedOk:     true,
				expectedLength: 2,
			}),
			Entry("input is an empty slice of interfaces", testCase{
				input:          []interface{}{},
				expectedOk:     true,
				expectedLength: 0,
			}),
			Entry("input is a slice containing non-client.Object values", testCase{
				input:          []interface{}{&corev1.ConfigMap{}, "not an object"},
				expectedOk:     false,
				expectedLength: 0,
			}),
			Entry("input is not a slice", testCase{
				input:          "not a slice",
				expectedOk:     false,
				expectedLength: 0,
			}),
			Entry("input is nil", testCase{
				input:          nil,
				expectedOk:     false,
				expectedLength: 0,
			}),
		)
	})

	Describe("IsUnstructured", func() {
		type testCase struct {
			input    client.Object
			expected bool
		}

		DescribeTable("checking if object is unstructured",
			func(tc testCase) {
				result := utilities.IsUnstructured(tc.input)
				Expect(result).To(Equal(tc.expected))
			},
			Entry("input is an Unstructured object", testCase{
				input: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "test-configmap",
							"namespace": "default",
						},
					},
				},
				expected: true,
			}),
			Entry("input is a ConfigMap object", testCase{
				input: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-configmap",
						Namespace: "default",
					},
				},
				expected: false,
			}),
		)
	})

	Describe("TypedFromUnstructured and UnstructuredFromObject", func() {
		var k8sClient client.Client

		BeforeEach(func() {
			// Create a client with the standard Kubernetes scheme
			s := runtime.NewScheme()
			err := scheme.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())
			k8sClient = fake.NewClientBuilder().WithScheme(s).Build()
		})

		Context("TypedFromUnstructured", func() {
			It("converts an unstructured ConfigMap to a typed ConfigMap", func() {
				// Create an unstructured ConfigMap
				unstructuredObj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name":      "test-cm",
							"namespace": "default",
						},
						"data": map[string]interface{}{
							"key1": "value1",
							"key2": "value2",
						},
					},
				}

				// Convert to typed object
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)

				// Verify no error occurred
				Expect(err).NotTo(HaveOccurred())

				// Verify the returned object is not nil
				Expect(typedObj).NotTo(BeNil())

				// Verify it's a ConfigMap with correct data
				cm, ok := typedObj.(*corev1.ConfigMap)
				Expect(ok).To(BeTrue(), "Expected a *corev1.ConfigMap")
				Expect(cm.Name).To(Equal("test-cm"))
				Expect(cm.Namespace).To(Equal("default"))
				Expect(cm.Data).To(HaveLen(2))
				Expect(cm.Data).To(HaveKeyWithValue("key1", "value1"))
				Expect(cm.Data).To(HaveKeyWithValue("key2", "value2"))

				// Verify GVK is set correctly
				gvk := cm.GetObjectKind().GroupVersionKind()
				Expect(gvk.Group).To(Equal(""))
				Expect(gvk.Version).To(Equal("v1"))
				Expect(gvk.Kind).To(Equal("ConfigMap"))
			})

			It("returns an error when unstructured object has no GVK", func() {
				// Create an unstructured object with no GVK
				unstructuredObj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name":      "test-unknown",
							"namespace": "default",
						},
					},
				}

				// Try to convert to typed object
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)

				// Verify error occurred
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unstructured object has no GroupVersionKind"))

				// Verify returned object is nil
				Expect(typedObj).To(BeNil())
			})

			It("returns an error for unknown GVK", func() {
				// Create an unstructured object with unknown GVK
				unstructuredObj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "unknown.group/v1",
						"kind":       "UnknownKind",
						"metadata": map[string]interface{}{
							"name":      "test-unknown",
							"namespace": "default",
						},
					},
				}

				// Try to convert to typed object
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)

				// Verify error occurred
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create object for GroupVersionKind"))
				Expect(err.Error()).To(ContainSubstring("unknown.group/v1"))
				Expect(err.Error()).To(ContainSubstring("UnknownKind"))

				// Verify returned object is nil
				Expect(typedObj).To(BeNil())
			})

			It("returns an error for invalid unstructured data", func() {
				// Create an unstructured ConfigMap with invalid data
				unstructuredObj := unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "test-cm",
							// Add an invalid field type that will cause conversion to fail
							"creationTimestamp": map[string]string{"invalid": "timestamp"},
						},
					},
				}

				// Try to convert to typed object
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)

				// Verify error occurred
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to convert unstructured object to typed"))

				// Verify returned object is nil
				Expect(typedObj).To(BeNil())
			})
		})

		Context("UnstructuredFromObject", func() {
			It("converts a typed ConfigMap to an unstructured object", func() {
				// Create a typed ConfigMap with TypeMeta set
				cm := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				}

				// Convert to unstructured
				unstructuredObj, err := utilities.UnstructuredFromObject(k8sClient, cm)

				// Verify no error occurred
				Expect(err).NotTo(HaveOccurred())

				// Verify metadata
				Expect(unstructuredObj.GetName()).To(Equal("test-cm"))
				Expect(unstructuredObj.GetNamespace()).To(Equal("default"))
				Expect(unstructuredObj.GetKind()).To(Equal("ConfigMap"))
				Expect(unstructuredObj.GetAPIVersion()).To(Equal("v1"))

				// Verify labels
				labels := unstructuredObj.GetLabels()
				Expect(labels).To(HaveLen(1))
				Expect(labels).To(HaveKeyWithValue("app", "test"))

				// Verify data
				data, found, err := unstructured.NestedStringMap(unstructuredObj.Object, "data")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue(), "Data field not found in unstructured object")
				Expect(data).To(HaveLen(2))
				Expect(data).To(HaveKeyWithValue("key1", "value1"))
				Expect(data).To(HaveKeyWithValue("key2", "value2"))
			})

			It("converts a typed ConfigMap without TypeMeta to an unstructured object", func() {
				// Create a typed ConfigMap without setting TypeMeta
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"key1": "value1",
					},
				}

				// Verify TypeMeta is empty
				Expect(cm.APIVersion).To(BeEmpty())
				Expect(cm.Kind).To(BeEmpty())

				// Convert to unstructured
				unstructuredObj, err := utilities.UnstructuredFromObject(k8sClient, cm)

				// Verify no error occurred
				Expect(err).NotTo(HaveOccurred())

				// Verify GVK was determined correctly from the scheme
				Expect(unstructuredObj.GetAPIVersion()).To(Equal("v1"))
				Expect(unstructuredObj.GetKind()).To(Equal("ConfigMap"))

				// Verify other metadata
				Expect(unstructuredObj.GetName()).To(Equal("test-cm"))
				Expect(unstructuredObj.GetNamespace()).To(Equal("default"))
			})
		})

		Context("Round-trip conversions", func() {
			It("performs a round-trip conversion with TypeMeta set", func() {
				// Start with a typed ConfigMap with TypeMeta
				originalCm := &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "ConfigMap",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: "default",
						Labels: map[string]string{
							"app":     "test",
							"version": "v1",
						},
						Annotations: map[string]string{
							"description": "Test ConfigMap",
							"created-by":  "utilities-test",
						},
					},
					Data: map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					},
				}

				// Convert to unstructured
				unstructuredObj, err := utilities.UnstructuredFromObject(k8sClient, originalCm)
				Expect(err).NotTo(HaveOccurred(), "Error converting to unstructured")

				// Convert back to typed
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)
				Expect(err).NotTo(HaveOccurred(), "Error converting back to typed")

				// Verify it's a ConfigMap
				roundTripCm, ok := typedObj.(*corev1.ConfigMap)
				Expect(ok).To(BeTrue(), "Expected a *corev1.ConfigMap")

				// Verify all metadata is preserved
				Expect(roundTripCm.Name).To(Equal(originalCm.Name))
				Expect(roundTripCm.Namespace).To(Equal(originalCm.Namespace))
				Expect(roundTripCm.Labels).To(Equal(originalCm.Labels))
				Expect(roundTripCm.Annotations).To(Equal(originalCm.Annotations))

				// Verify all data is preserved
				Expect(roundTripCm.Data).To(Equal(originalCm.Data))

				// Verify GVK is preserved
				gvk := roundTripCm.GetObjectKind().GroupVersionKind()
				Expect(gvk.Group).To(Equal(""))
				Expect(gvk.Version).To(Equal("v1"))
				Expect(gvk.Kind).To(Equal("ConfigMap"))
			})

			It("performs a round-trip conversion without TypeMeta set", func() {
				// Start with a typed ConfigMap without TypeMeta
				originalCm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cm",
						Namespace: "default",
						Labels: map[string]string{
							"app": "test",
						},
					},
					Data: map[string]string{
						"key1": "value1",
					},
				}

				// Verify TypeMeta is empty
				Expect(originalCm.APIVersion).To(BeEmpty())
				Expect(originalCm.Kind).To(BeEmpty())

				// Convert to unstructured
				unstructuredObj, err := utilities.UnstructuredFromObject(k8sClient, originalCm)
				Expect(err).NotTo(HaveOccurred(), "Error converting to unstructured")

				// Verify GVK was determined correctly
				Expect(unstructuredObj.GetAPIVersion()).To(Equal("v1"))
				Expect(unstructuredObj.GetKind()).To(Equal("ConfigMap"))

				// Convert back to typed
				typedObj, err := utilities.TypedFromUnstructured(k8sClient, unstructuredObj)
				Expect(err).NotTo(HaveOccurred(), "Error converting back to typed")

				// Verify it's a ConfigMap
				roundTripCm, ok := typedObj.(*corev1.ConfigMap)
				Expect(ok).To(BeTrue(), "Expected a *corev1.ConfigMap")

				// Verify metadata is preserved
				Expect(roundTripCm.Name).To(Equal(originalCm.Name))
				Expect(roundTripCm.Namespace).To(Equal(originalCm.Namespace))
				Expect(roundTripCm.Labels).To(Equal(originalCm.Labels))

				// Verify data is preserved
				Expect(roundTripCm.Data).To(Equal(originalCm.Data))

				// Verify GVK is set correctly after round trip
				gvk := roundTripCm.GetObjectKind().GroupVersionKind()
				Expect(gvk.Group).To(Equal(""))
				Expect(gvk.Version).To(Equal("v1"))
				Expect(gvk.Kind).To(Equal("ConfigMap"))
			})
		})
	})
})
