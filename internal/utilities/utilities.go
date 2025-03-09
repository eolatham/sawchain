package utilities

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: test

// MergeMaps merges the given maps into one.
func MergeMaps(maps ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// IsExistingFile checks if the given path exists and is a file.
func IsExistingFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ReadFileContent reads a file and returns its content as a string.
func ReadFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// AsDuration attempts to convert the given value into a time.Duration.
func AsDuration(v interface{}) (time.Duration, bool) {
	// Check if it's already a time.Duration
	if d, ok := v.(time.Duration); ok {
		return d, true
	}

	// Check if it's a string that can be parsed as a time.Duration
	if str, ok := v.(string); ok {
		if d, err := time.ParseDuration(str); err == nil {
			return d, true
		}
	}

	return 0, false
}

// AsMapStringAny attempts to convert the given value into a map with string keys.
func AsMapStringAny(v interface{}) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}

	// Use reflection to check if it's a map with string keys
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := make(map[string]any)
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key().Interface().(string)
			result[key] = iter.Value().Interface()
		}
		return result, true
	}

	return nil, false
}

// AsObject attempts to convert the given value into a client.Object.
func AsObject(v interface{}) (client.Object, bool) {
	if obj, ok := v.(client.Object); ok {
		return obj, true
	}
	return nil, false
}

// AsSliceOfObjects attempts to convert the given value into a slice of client.Object.
func AsSliceOfObjects(v interface{}) ([]client.Object, bool) {
	items, ok := v.([]interface{})
	if !ok {
		return nil, false
	}

	var objs []client.Object
	for _, item := range items {
		if obj, ok := AsObject(item); ok {
			objs = append(objs, obj)
		} else {
			return nil, false
		}
	}

	return objs, true
}

// IsUnstructured checks if the given object's value
// is of type unstructured.Unstructured.
func IsUnstructured(obj client.Object) bool {
	switch obj.(type) {
	case *unstructured.Unstructured:
		return true
	default:
		return false
	}
}

// TypedFromUnstructured uses the client scheme to convert
// the given unstructured object to a typed object.
func TypedFromUnstructured(
	c client.Client,
	obj unstructured.Unstructured,
) (client.Object, error) {
	// Get scheme
	scheme := c.Scheme()
	if scheme == nil {
		return nil, fmt.Errorf("client scheme is not set")
	}
	// Create typed object
	gvk := obj.GroupVersionKind()
	runtimeObj, err := scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("failed to create object for GVK %v: %w", gvk, err)
	}
	// Convert unstructured object
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, runtimeObj); err != nil {
		return nil, fmt.Errorf("failed to convert unstructured object to typed: %w", err)
	}
	// Return as client.Object
	clientObj, ok := runtimeObj.(client.Object)
	if !ok {
		return nil, fmt.Errorf("object of type %T does not implement client.Object", runtimeObj)
	}
	return clientObj, nil
}

// UnstructuredFromObject uses the client scheme to convert
// the given object to an unstructured object.
func UnstructuredFromObject(
	c client.Client,
	obj client.Object,
) (*unstructured.Unstructured, error) {
	// Get scheme
	scheme := c.Scheme()
	if scheme == nil {
		return nil, fmt.Errorf("client scheme is not set")
	}
	// Convert object
	unstructuredObj := &unstructured.Unstructured{}
	err := c.Scheme().Convert(obj, unstructuredObj, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to unstructured: %w", err)
	}
	// Set GVK
	gvk := obj.GetObjectKind().GroupVersionKind()
	unstructuredObj.SetGroupVersionKind(gvk)
	return unstructuredObj, nil
}
