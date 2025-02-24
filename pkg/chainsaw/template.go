package chainsaw

import (
	"context"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// loadTemplateResource loads the template and returns its unstructured contents.
// Expects the template to contain a single resource.
func loadTemplateResource(templateContent string) (unstructured.Unstructured, error) {
	var r unstructured.Unstructured
	resources, err := resource.Parse([]byte(templateContent), true)
	if err != nil {
		return r, fmt.Errorf("failed to parse template: %w", err)
	}
	if len(resources) != 1 {
		return r, fmt.Errorf("expected template to contain a single resource; found %d", len(resources))
	}
	r = resources[0]
	return r, nil
}

// bindingsFromMap converts the map into an object that can applied to Chainsaw templates.
func bindingsFromMap(ctx context.Context, bindingsMap map[string]any) apis.Bindings {
	bindingsObj := apis.NewBindings()
	for k, v := range bindingsMap {
		bindingsObj = bindings.RegisterBinding(ctx, bindingsObj, k, v)
	}
	return bindingsObj
}
