package chainsaw

import (
	"context"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/engine/checks"
	operrors "github.com/kyverno/chainsaw/pkg/engine/operations/errors"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/utilities"
)

const (
	errExpectedSingleResource = "expected template to contain a single resource; found %d"
)

// TODO: implement new functions here and remove previous versions

func Check()

// ListCandidates lists all resources from the cluster that may match the expected resource.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/internal.Read.
func ListCandidates(
	c client.Client,
	ctx context.Context,
	expected client.Object,
) ([]unstructured.Unstructured, error) {
	var results []unstructured.Unstructured
	gvk := expected.GetObjectKind().GroupVersionKind()
	useGet := expected.GetName() != ""
	if useGet {
		var actual unstructured.Unstructured
		actual.SetGroupVersionKind(gvk)
		if err := c.Get(ctx, client.ObjectKeyFromObject(expected), &actual); err != nil {
			return nil, err
		}
		results = append(results, actual)
	} else {
		var list unstructured.UnstructuredList
		list.SetGroupVersionKind(gvk)
		var listOptions []client.ListOption
		if expected.GetNamespace() != "" {
			listOptions = append(listOptions, client.InNamespace(expected.GetNamespace()))
		}
		if len(expected.GetLabels()) != 0 {
			listOptions = append(listOptions, client.MatchingLabels(expected.GetLabels()))
		}
		if err := c.List(ctx, &list, listOptions...); err != nil {
			return nil, err
		}
		results = append(results, list.Items...)
	}
	return results, nil
}

// Match tests each candidate against the expectations defined in the resource
// and returns the first match or an error if no match is found.
func Match(
	ctx context.Context,
	resource unstructured.Unstructured,
	bindings apis.Bindings,
	candidates ...unstructured.Unstructured,
) (unstructured.Unstructured, error) {
	var errs []error

	// Execute resource check for each candidate
	for _, candidate := range candidates {
		fieldErrs, err := checks.Check(ctx, compilers, candidate.UnstructuredContent(), bindings,
			ptr.To(v1alpha1.NewCheck(resource.UnstructuredContent())))
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(fieldErrs) != 0 {
			errs = append(errs,
				operrors.ResourceError(compilers, resource, candidate, true, bindings, fieldErrs),
			)
		} else {
			// Match found
			return candidate, nil
		}
	}

	// No matches found
	return unstructured.Unstructured{}, multierr.Combine(errs...)
}

// UnstructuredObjectsFromTemplate loads the template and returns unstructured objects
// ready for Chainsaw matching.
func UnstructuredObjectsFromTemplate(templateContent string) ([]unstructured.Unstructured, error) {
	resources, err := resource.Parse([]byte(templateContent), true)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	return resources, nil
}

// UnstructuredObjectsFromTemplate loads the template and returns an unstructured object
// ready for Chainsaw matching. Expects the template to contain a single resource.
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

// StructuredObjectsFromTemplate parses the template and returns typed objects
// ready for K8s operations.
func StructuredObjectsFromTemplate(
	c client.Client,
	ctx context.Context,
	templateContent string,
	bindingsMap map[string]any,
) ([]client.Object, error) {
	// Load unstructured objects
	unstructuredObjs, err := UnstructuredObjectsFromTemplate(templateContent)
	if err != nil {
		return nil, err
	}

	// Convert bindings
	bindings := bindingsFromMap(ctx, bindingsMap)

	// Convert unstructured objects
	var structuredObjs []client.Object
	for _, unstructuredObj := range unstructuredObjs {
		// Merge templated fields
		unstructuredObj, err := templating.TemplateAndMerge(
			ctx, apis.DefaultCompilers, unstructuredObj, bindings,
			v1alpha1.NewProjection(unstructuredObj.UnstructuredContent()))
		if err != nil {
			return nil, err
		}
		// Convert to typed
		structuredObj, err := utilities.ToTyped(c, unstructuredObj)
		if err != nil {
			return nil, err
		}
		structuredObjs = append(structuredObjs, structuredObj)
	}

	return structuredObjs, nil
}

// StructuredObjectFromTemplate parses the template and returns a typed object
// ready for K8s operations. Expects the template to contain a single resource.
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
