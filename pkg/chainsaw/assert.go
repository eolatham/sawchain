package chainsaw

import (
	"context"
	"errors"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/client"
	"github.com/kyverno/chainsaw/pkg/engine/checks"
	operrors "github.com/kyverno/chainsaw/pkg/engine/operations/errors"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"go.uber.org/multierr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
)

func assert(c client.Client, ctx context.Context, bindings apis.Bindings, obj unstructured.Unstructured) error {
	compilers := apis.DefaultCompilers

	if bindings == nil {
		bindings = apis.NewBindings()
	}
	if err := templating.ResourceRef(ctx, compilers, &obj, bindings); err != nil {
		return err
	}

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

	candidates, err := getResources(c, ctx, &obj)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return errors.New("actual resource not found")
		}
		return err
	}
	if len(candidates) == 0 {
		return errors.New("no actual resource found")
	}

	for _, candidate := range candidates {
		fieldErrs, err := checks.Check(ctx, compilers, candidate.UnstructuredContent(), bindings, ptr.To(v1alpha1.NewCheck(obj.UnstructuredContent())))
		if err != nil {
			return err
		}
		if len(fieldErrs) != 0 {
			errs = append(errs, operrors.ResourceError(compilers, obj, candidate, true, bindings, fieldErrs))
		} else {
			// Found a match, can return success
			return nil
		}
	}

	// If we get here, no matches were found
	return multierr.Combine(errs...)
}
