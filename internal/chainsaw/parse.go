package chainsaw

import (
	"context"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ParseResource parses the resource in the template and returns it as a structured object.
func ParseResource(
	c client.Client,
	ctx context.Context,
	templateContent string,
	bindingsMap map[string]any,
) (client.Object, error) {
	// Load resource
	resource, err := loadTemplateResource(templateContent)
	if err != nil {
		return nil, err
	}

	// Convert bindings
	bindings := bindingsFromMap(ctx, bindingsMap)

	// Parse and merge templated fields into unstructured object
	compilers := apis.DefaultCompilers
	template := v1alpha1.NewProjection(resource.UnstructuredContent())
	merged, err := templating.TemplateAndMerge(ctx, compilers, resource, bindings, template)
	if err != nil {
		return nil, err
	}

	// Return structured object
	return convertToStruct(c, merged)
}
