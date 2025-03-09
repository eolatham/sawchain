package chainsaw_test

import (
	"github.com/kyverno/chainsaw/pkg/apis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/eolatham/sawchain/internal/chainsaw"
)

var _ = DescribeTable("BindingsFromMap",
	func(bindingsMap map[string]any) {
		// Test BindingsFromMap
		bindings := BindingsFromMap(bindingsMap)

		// Verify bindings
		if len(bindingsMap) == 0 {
			Expect(bindings).To(Equal(apis.NewBindings()))
		} else {
			for name, expectedValue := range bindingsMap {
				binding, err := bindings.Get("$" + name)
				Expect(err).NotTo(HaveOccurred(), "Expected binding %s not found", name)
				actualValue, err := binding.Value()
				Expect(err).NotTo(HaveOccurred(), "Failed to extract value for binding %s", name)
				if expectedValue == nil {
					Expect(actualValue).To(BeNil())
				} else {
					Expect(actualValue).To(Equal(expectedValue))
				}
			}
		}
	},
	// Empty map
	Entry("should handle empty map",
		map[string]any{},
	),
	// Single binding
	Entry("should convert single binding",
		map[string]any{
			"key": "value",
		},
	),
	// Multiple bindings
	Entry("should convert multiple bindings",
		map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		},
	),
	// Different value types
	Entry("should handle different value types",
		map[string]any{
			"string": "text",
			"int":    123,
			"bool":   true,
			"float":  3.14,
			"slice":  []string{"a", "b"},
			"map":    map[string]string{"k": "v"},
			"nilVal": nil,
		},
	),
)
