package utilities

import (
	"os"
	"reflect"
	"time"

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

// AsClientObject attempts to convert the given value into a client.Object.
func AsClientObject(v interface{}) (client.Object, bool) {
	if obj, ok := v.(client.Object); ok {
		return obj, true
	}
	return nil, false
}

// AsSliceOfClientObjects attempts to convert the given value into a slice of client.Object.
func AsSliceOfClientObjects(v interface{}) ([]client.Object, bool) {
	items, ok := v.([]interface{})
	if !ok {
		return nil, false
	}

	var objs []client.Object
	for _, item := range items {
		if obj, ok := AsClientObject(item); ok {
			objs = append(objs, obj)
		} else {
			return nil, false
		}
	}

	return objs, true
}

// AsMapStringAny attempts to convert the given value into a map[string]any.
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
