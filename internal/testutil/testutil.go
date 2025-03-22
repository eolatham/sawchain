package testutil

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type TestResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            TestResourceStatus `json:"status,omitempty"`
}

type TestResourceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (t *TestResource) DeepCopyObject() runtime.Object {
	return &TestResource{
		TypeMeta:   t.TypeMeta,
		ObjectMeta: t.ObjectMeta,
		Status:     t.Status,
	}
}

// CreateTempDir creates a temporary directory and returns its path.
func CreateTempDir(namePattern string) string {
	tempDir, err := os.MkdirTemp("", namePattern)
	if err != nil {
		panic(err)
	}
	return tempDir
}

// CreateTempFile creates a temporary file and returns its path.
func CreateTempFile(namePattern, content string) string {
	file, err := os.CreateTemp("", namePattern)
	if err != nil {
		panic(err)
	}
	path := file.Name()
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	return path
}

// NewEmptyScheme returns a new empty runtime.Scheme.
func NewEmptyScheme() *runtime.Scheme {
	return runtime.NewScheme()
}

// NewStandardScheme returns a new standard runtime.scheme supporting built-in APIs.
func NewStandardScheme() *runtime.Scheme {
	s := NewEmptyScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		panic(err)
	}
	return s
}

// NewStandardSchemeWithTestResource returns a new standard runtime.scheme supporting built-in APIs
// and the custom TestResource type.
func NewStandardSchemeWithTestResource() *runtime.Scheme {
	s := NewStandardScheme()
	s.AddKnownTypes(schema.GroupVersion{Group: "example.com", Version: "v1"}, &TestResource{})
	return s
}

// TODO: document, test, and make use of the functions below

func NewConfigMap(
	name, namespace string,
	data map[string]string,
) *corev1.ConfigMap {
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

func NewUnstructuredConfigMap(
	name, namespace string,
	data map[string]string,
) *unstructured.Unstructured {
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

func NewTestResource(
	name, namespace string,
	conditions []metav1.Condition,
) *TestResource {
	return &TestResource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "example.com/v1",
			Kind:       "TestResource",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: TestResourceStatus{
			Conditions: conditions,
		},
	}
}

func NewUnstructuredTestResource(
	name, namespace string,
	conditions []metav1.Condition,
) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("example.com/v1")
	obj.SetKind("TestResource")
	obj.SetName(name)
	obj.SetNamespace(namespace)

	status := map[string]interface{}{}
	conditionsData := make([]interface{}, len(conditions))

	for i, condition := range conditions {
		conditionMap := map[string]interface{}{
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
