package chainsaw

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/checks"
	operrors "github.com/kyverno/chainsaw/pkg/engine/operations/errors"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"go.uber.org/multierr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var compilers = apis.DefaultCompilers

// CheckResource checks if the resource in the template matches a resource in the cluster.
// Returns the first matching resource on success.
func CheckResource(
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

	// Check resource
	match, err := check(c, ctx, bindings, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to execute check: %w", err)
	}

	// Return match as structured object
	obj, err := convertToStruct(c, match)
	if err != nil {
		return nil, fmt.Errorf("failed to convert match to struct: %w", err)
	}
	return obj, nil
}

// check is equivalent to a Chainsaw assert operation without polling.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/assert.Exec.
// Returns the first matching resource on success.
func check(
	c client.Client,
	ctx context.Context,
	bindings apis.Bindings,
	resource unstructured.Unstructured,
) (unstructured.Unstructured, error) {
	// Parse template
	if bindings == nil {
		bindings = apis.NewBindings()
	}
	if err := templating.ResourceRef(ctx, compilers, &resource, bindings); err != nil {
		return unstructured.Unstructured{}, err
	}

	// Execute non-resource check
	var errs []error
	if resource.GetAPIVersion() == "" || resource.GetKind() == "" {
		fieldErrs, err := checks.Check(ctx, compilers, nil, bindings,
			ptr.To(v1alpha1.NewCheck(resource.UnstructuredContent())))
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(fieldErrs) != 0 {
			return unstructured.Unstructured{}, multierr.Combine(fieldErrs.ToAggregate().Errors()...)
		}
		return unstructured.Unstructured{}, nil
	}

	// Search for resource candidates
	candidates, err := read(c, ctx, &resource)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return unstructured.Unstructured{}, errors.New("actual resource not found")
		}
		return unstructured.Unstructured{}, err
	}
	if len(candidates) == 0 {
		return unstructured.Unstructured{}, errors.New("no actual resource found")
	}

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

// read attempts to get all resources from the cluster that match the expected resource.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/internal.Read.
func read(
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
