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

// TODO: consider returning matched object on success
// CheckResource checks if the resource in the template file matches a resource in the cluster.
func CheckResource(c client.Client, ctx context.Context, templatePath string, bindingsMap map[string]any) error {
	// Load resource
	r, err := loadTemplateResource(templatePath)
	if err != nil {
		return err
	}

	// Convert bindings
	bindings := bindingsFromMap(ctx, bindingsMap)

	// Check resource
	if err := check(c, ctx, bindings, r); err != nil {
		return fmt.Errorf("failed to execute check: %w", err)
	}
	return nil
}

// check is equivalent to a Chainsaw assert operation without polling.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/assert.Exec.
func check(c client.Client, ctx context.Context, bindings apis.Bindings, obj unstructured.Unstructured) error {
	// Use default compilers
	compilers := apis.DefaultCompilers

	// Parse template
	if bindings == nil {
		bindings = apis.NewBindings()
	}
	if err := templating.ResourceRef(ctx, compilers, &obj, bindings); err != nil {
		return err
	}

	// Execute non-resource check
	var errs []error
	if obj.GetAPIVersion() == "" || obj.GetKind() == "" {
		fieldErrs, err := checks.Check(ctx, compilers, nil, bindings, ptr.To(v1alpha1.NewCheck(obj.UnstructuredContent())))
		if err != nil {
			return err
		}
		if len(fieldErrs) != 0 {
			return multierr.Combine(fieldErrs.ToAggregate().Errors()...)
		}
		return nil
	}

	// Search for resource candidates
	candidates, err := read(c, ctx, &obj)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return errors.New("actual resource not found")
		}
		return err
	}
	if len(candidates) == 0 {
		return errors.New("no actual resource found")
	}

	// Execute resource check for each candidate
	for _, candidate := range candidates {
		fieldErrs, err := checks.Check(ctx, compilers, candidate.UnstructuredContent(), bindings, ptr.To(v1alpha1.NewCheck(obj.UnstructuredContent())))
		if err != nil {
			return err
		}
		if len(fieldErrs) != 0 {
			errs = append(errs, operrors.ResourceError(compilers, obj, candidate, true, bindings, fieldErrs))
		} else {
			// Match found
			return nil
		}
	}

	// No matches found
	return multierr.Combine(errs...)
}

// read attempts to get all resources from the cluster that match the expected resource.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/internal.Read.
func read(c client.Client, ctx context.Context, expected client.Object) ([]unstructured.Unstructured, error) {
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
