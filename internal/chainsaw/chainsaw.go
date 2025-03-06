package chainsaw

import (
	"context"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errExpectedSingleResource = "expected template to contain a single resource; found %d"
)

// TODO: implement new functions here and remove previous versions

func CheckResource()
func ListCandidates()
func MatchCandidates()

// UnstructuredObjectsFromTemplate loads the template and
// returns unstructured objects ready for Chainsaw matching.
func UnstructuredObjectsFromTemplate(templateContent string) ([]unstructured.Unstructured, error) {
	resources, err := resource.Parse([]byte(templateContent), true)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	return resources, nil
}

// UnstructuredObjectsFromTemplate loads the template and
// returns an unstructured object ready for Chainsaw matching.
// Expects the template to contain a single resource.
func UnstructuredObjectFromTemplate(templateContent string) (unstructured.Unstructured, error) {
	objs, err := UnstructuredObjectsFromTemplate(templateContent)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	if len(objs) != 1 {
		return unstructured.Unstructured{}, fmt.Errorf(errExpectedSingleResource, len(objs))
	}
	return objs[0], nil
}

// convertToStruct converts the unstructured resource into the appropriate client.Object struct.
func convertToStruct(c client.Client, resource unstructured.Unstructured) (client.Object, error) {
	// Get GVK from unstructured object
	gvk := resource.GroupVersionKind()

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
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, typed); err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to typed object: %w", err)
	}

	// Convert to client.Object (which all K8s types implement)
	obj, ok := typed.(client.Object)
	if !ok {
		return nil, fmt.Errorf("object of type %T does not implement client.Object", typed)
	}

	// Return structured object
	return obj, nil
}

// StructuredObjectsFromTemplate parses the template and
// returns structured objects ready for K8s operations.
func StructuredObjectsFromTemplate(
	c client.Client,
	ctx context.Context,
	templateContent string,
	bindingsMap map[string]any,
) ([]client.Object, error) {
	// Load unstructured objects from template
	unstructuredObjs, err := UnstructuredObjectsFromTemplate(templateContent)
	if err != nil {
		return nil, err
	}

	// Convert bindings
	bindings := bindingsFromMap(ctx, bindingsMap)

	// Convert unstructured objects to structured objects
	var structuredObjs []client.Object
	for _, unstructuredObj := range unstructuredObjs {
		// Parse and merge templated fields
		unstructuredObj, err := templating.TemplateAndMerge(
			ctx, apis.DefaultCompilers, unstructuredObj, bindings,
			v1alpha1.NewProjection(unstructuredObj.UnstructuredContent()))
		if err != nil {
			return nil, err
		}
		// Convert to struct
		structuredObj, err := convertToStruct(c, unstructuredObj)
		if err != nil {
			return nil, err
		}
		structuredObjs = append(structuredObjs, structuredObj)
	}

	return structuredObjs, nil
}

// StructuredObjectFromTemplate parses the template and
// returns a structured object ready for K8s operations.
// Expects the template to contain a single resource.
func StructuredObjectFromTemplate(
	c client.Client,
	ctx context.Context,
	templateContent string,
	bindingsMap map[string]any,
) (client.Object, error) {
	objs, err := StructuredObjectsFromTemplate(c, ctx, templateContent, bindingsMap)
	if err != nil {
		return nil, err
	}
	if len(objs) != 1 {
		return nil, fmt.Errorf(errExpectedSingleResource, len(objs))
	}
	return objs[0], nil
}

// BindingsFromMap converts the map into an object that can applied to Chainsaw templates.
func BindingsFromMap(ctx context.Context, bindingsMap map[string]any) apis.Bindings {
	bindingsObj := apis.NewBindings()
	for k, v := range bindingsMap {
		bindingsObj = bindings.RegisterBinding(ctx, bindingsObj, k, v)
	}
	return bindingsObj
}
