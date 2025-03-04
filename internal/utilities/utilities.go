package utilities

import (
	"reflect"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: test
func MergeMaps(maps ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// TODO: test
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

// TODO: test
func AsClientObject(v interface{}) (client.Object, bool) {
	if obj, ok := v.(client.Object); ok {
		return obj, true
	}
	return nil, false
}

// TODO: test
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
