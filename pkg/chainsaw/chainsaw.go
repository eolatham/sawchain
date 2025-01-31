package chainsaw

import (
	"context"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/client"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
)

// CheckResources checks if the resources in the template file match the resources in the cluster.
func CheckResources(c client.Client, ctx context.Context, templatePath string, bindingsMap map[string]any) error {
	// Load Chainsaw template resources
	resources, err := resource.Load(templatePath, true)
	if err != nil {
		return fmt.Errorf("failed to load template file %s: %w", templatePath, err)
	}

	// Convert map to Chainsaw bindings
	bindingsObj := apis.NewBindings()
	for k, v := range bindingsMap {
		bindingsObj = bindings.RegisterBinding(ctx, bindingsObj, k, v)
	}

	// For each resource, execute Chainsaw assertion
	for _, resource := range resources {
		if err := assert(c, ctx, bindingsObj, resource); err != nil {
			return fmt.Errorf("failed to execute assertion: %w", err)
		}
	}
	return nil
}
