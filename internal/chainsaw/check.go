package chainsaw

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/checks"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"go.uber.org/multierr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/utilities"
)

var compilers = apis.DefaultCompilers

// CheckResourceOld checks if the resource in the template matches a resource in the cluster.
// Returns the first matching resource on success.
func CheckResourceOld(
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

	// Return match as typed object
	obj, err := utilities.ToTyped(c, match)
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
	candidates, err := ListCandidates(c, ctx, &resource)
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
	return Match(ctx, resource, bindings, candidates...)
}
