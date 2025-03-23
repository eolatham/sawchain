package testutil_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/eolatham/sawchain/internal/testutil"
)

var _ = Describe("Testutil", func() {
	DescribeTable("CreateTempDir",
		func(namePattern string) {
			tempDirPath := testutil.CreateTempDir(namePattern)

			// Verify the directory exists and has the right pattern
			Expect(tempDirPath).To(ContainSubstring(namePattern))
			info, err := os.Stat(tempDirPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			// Clean up
			os.RemoveAll(tempDirPath)
		},
		Entry("with test pattern", "test-pattern"),
		Entry("with empty pattern", ""),
	)

	DescribeTable("CreateTempFile",
		func(namePattern, content string) {
			tempFilePath := testutil.CreateTempFile(namePattern, content)

			// Verify the file exists and has the right pattern and content
			Expect(tempFilePath).To(ContainSubstring(namePattern))
			fileContent, err := os.ReadFile(tempFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(fileContent)).To(Equal(content))

			// Clean up
			os.Remove(tempFilePath)
		},
		Entry("with test pattern and content", "test-file-pattern", "test content"),
		Entry("with empty pattern and content", "", "test content"),
		Entry("with pattern and empty content", "test-file-pattern", ""),
	)

	DescribeTable("CreateEmptyScheme",
		func() {
			scheme := testutil.NewEmptyScheme()
			Expect(scheme).NotTo(BeNil())
			Expect(scheme.AllKnownTypes()).To(HaveLen(0))
		},
		Entry("creates empty scheme"),
	)

	DescribeTable("CreateStandardScheme",
		func() {
			scheme := testutil.NewStandardScheme()
			Expect(scheme).NotTo(BeNil())
			Expect(scheme.AllKnownTypes()).NotTo(BeEmpty())
		},
		Entry("creates standard scheme"),
	)

	DescribeTable("CreateStandardSchemeWithTestResource",
		func() {
			scheme := testutil.NewStandardSchemeWithTestResource()
			Expect(scheme).NotTo(BeNil())
			Expect(len(scheme.AllKnownTypes()) > 1).To(BeTrue(), "scheme should have multiple types")
			// Verify TestResource is registered
			obj, err := scheme.New(schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "TestResource",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(obj).To(BeAssignableToTypeOf(&testutil.TestResource{}))
		},
		Entry("creates standard scheme with TestResource"),
	)

	DescribeTable("NewConfigMap",
		func(name, namespace string, data map[string]string) {
			cm := testutil.NewConfigMap(name, namespace, data)
			Expect(cm.APIVersion).To(Equal("v1"))
			Expect(cm.Kind).To(Equal("ConfigMap"))
			Expect(cm.Name).To(Equal(name))
			Expect(cm.Namespace).To(Equal(namespace))
			Expect(cm.Data).To(Equal(data))
		},
		Entry("with non-empty data", "test-cm", "default", map[string]string{"key1": "value1", "key2": "value2"}),
		Entry("with empty data", "test-cm", "default", map[string]string{}),
		Entry("with nil data", "test-cm", "default", nil),
	)

	DescribeTable("NewUnstructuredConfigMap",
		func(name, namespace string, data map[string]string) {
			unstructuredCm := testutil.NewUnstructuredConfigMap(name, namespace, data)
			Expect(unstructuredCm.GetAPIVersion()).To(Equal("v1"))
			Expect(unstructuredCm.GetKind()).To(Equal("ConfigMap"))
			Expect(unstructuredCm.GetName()).To(Equal(name))
			Expect(unstructuredCm.GetNamespace()).To(Equal(namespace))
			// Check data
			unstructuredData, found, err := unstructured.NestedMap(unstructuredCm.Object, "data")
			Expect(err).NotTo(HaveOccurred(), "failed to get data from unstructured ConfigMap")
			Expect(found).To(BeTrue(), "data not found in unstructured ConfigMap")
			Expect(unstructuredData).To(HaveLen(len(data)))
			for k, v := range data {
				Expect(unstructuredData).To(HaveKeyWithValue(k, v))
			}
		},
		Entry("with non-empty data", "test-cm", "default", map[string]string{"key1": "value1", "key2": "value2"}),
		Entry("with empty data", "test-cm", "default", map[string]string{}),
		Entry("with nil data", "test-cm", "default", nil),
	)

	DescribeTable("NewTestResource",
		func(name, namespace string, conditions []metav1.Condition) {
			tr := testutil.NewTestResource(name, namespace, conditions)
			Expect(tr.APIVersion).To(Equal("example.com/v1"))
			Expect(tr.Kind).To(Equal("TestResource"))
			Expect(tr.Name).To(Equal(name))
			Expect(tr.Namespace).To(Equal(namespace))
			Expect(tr.Status.Conditions).To(Equal(conditions))
		},
		Entry("with non-empty conditions",
			"test-resource",
			"default",
			[]metav1.Condition{
				{
					Type:    "Ready",
					Status:  metav1.ConditionTrue,
					Reason:  "TestReason",
					Message: "Test message",
				},
			},
		),
		Entry("with empty conditions", "test-resource", "default", []metav1.Condition{}),
		Entry("with nil conditions", "test-resource", "default", nil),
	)

	DescribeTable("NewUnstructuredTestResource",
		func(name, namespace string, conditions []metav1.Condition) {
			unstructuredTr := testutil.NewUnstructuredTestResource(name, namespace, conditions)
			Expect(unstructuredTr.GetAPIVersion()).To(Equal("example.com/v1"))
			Expect(unstructuredTr.GetKind()).To(Equal("TestResource"))
			Expect(unstructuredTr.GetName()).To(Equal(name))
			Expect(unstructuredTr.GetNamespace()).To(Equal(namespace))
			// Check conditions
			unstructuredConditions, found, err := unstructured.NestedSlice(unstructuredTr.Object, "status", "conditions")
			Expect(err).NotTo(HaveOccurred(), "failed to get conditions from unstructured TestResource")
			Expect(found).To(BeTrue(), "conditions not found in unstructured TestResource")
			Expect(unstructuredConditions).To(HaveLen(len(conditions)))
			for i, condition := range conditions {
				Expect(unstructuredConditions[i]).To(HaveKeyWithValue("type", condition.Type))
				Expect(unstructuredConditions[i]).To(HaveKeyWithValue("status", string(condition.Status)))
				Expect(unstructuredConditions[i]).To(HaveKeyWithValue("reason", condition.Reason))
				Expect(unstructuredConditions[i]).To(HaveKeyWithValue("message", condition.Message))
				Expect(unstructuredConditions[i]).To(HaveKeyWithValue("lastTransitionTime", condition.LastTransitionTime.String()))
			}
		},
		Entry("with non-empty conditions",
			"test-resource",
			"default",
			[]metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					Reason:             "test reason",
					Message:            "test message",
					LastTransitionTime: metav1.Now(),
				},
			},
		),
		Entry("with empty conditions", "test-resource", "default", []metav1.Condition{}),
		Entry("with nil conditions", "test-resource", "default", nil),
	)
})
