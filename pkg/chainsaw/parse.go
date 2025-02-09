package chainsaw

import (
	"context"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ParseResources parses the resources in the template file and returns them as structured objects.
func ParseResources(c client.Client, ctx context.Context, templatePath string, bindingsMap map[string]any) ([]client.Object, error) {
	// Load Chainsaw template resources
	resources, err := resource.Load(templatePath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load template file %s: %w", templatePath, err)
	}

	// Convert map to Chainsaw bindings
	bindingsObj := apis.NewBindings()
	for k, v := range bindingsMap {
		bindingsObj = bindings.RegisterBinding(ctx, bindingsObj, k, v)
	}

	// Parse template resources
	var objects []client.Object
	for _, resource := range resources {
		// Parse and merge templated fields into unstructured object
		compilers := apis.DefaultCompilers
		template := v1alpha1.NewProjection(resource.UnstructuredContent())
		merged, err := templating.TemplateAndMerge(ctx, compilers, resource, bindingsObj, template)
		if err != nil {
			return nil, err
		}

		// Get GVK from unstructured object
		gvk := merged.GroupVersionKind()

		// Create new instance of the correct type
		scheme := c.Scheme()
		if scheme == nil {
			return nil, fmt.Errorf("client scheme is not set")
		}
		typed, err := scheme.New(gvk)
		if err != nil {
			return nil, fmt.Errorf("failed to create object for GVK %v: %w", gvk, err)
		}

		// Convert unstructured object to typed one
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(merged.Object, typed); err != nil {
			return nil, fmt.Errorf("failed to convert unstructured to typed object: %w", err)
		}

		// Convert to client.Object (which all K8s types implement)
		obj, ok := typed.(client.Object)
		if !ok {
			return nil, fmt.Errorf("object of type %T does not implement client.Object", typed)
		}

		// Add to list of objects
		objects = append(objects, obj)
	}
	return objects, nil
}
