package chainsaw

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/engine/bindings"
	"github.com/kyverno/chainsaw/pkg/engine/checks"
	operrors "github.com/kyverno/chainsaw/pkg/engine/operations/errors"
	"github.com/kyverno/chainsaw/pkg/engine/templating"
	"github.com/kyverno/chainsaw/pkg/loaders/resource"
	"go.uber.org/multierr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Bindings = apis.Bindings

const errExpectedSingleResource = "expected template to contain a single resource; found %d"

var compilers = apis.DefaultCompilers

// BindingsFromMap converts the map into a Bindings object.
func BindingsFromMap(m map[string]any) Bindings {
	b := apis.NewBindings()
	for k, v := range m {
		b = bindings.RegisterBinding(context.TODO(), b, k, v)
	}
	return b
}

// ParseTemplate parses the template into unstructured objects
// (without processing template expressions).
func ParseTemplate(templateContent string) ([]unstructured.Unstructured, error) {
	objs, err := resource.Parse([]byte(templateContent), true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return objs, nil
}

// ParseTemplateSingle parses the single-resource template into an unstructured object
// (without processing template expressions).
func ParseTemplateSingle(templateContent string) (unstructured.Unstructured, error) {
	parsed, err := ParseTemplate(templateContent)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	if len(parsed) != 1 {
		return unstructured.Unstructured{}, fmt.Errorf(errExpectedSingleResource, len(parsed))
	}
	return parsed[0], nil
}

// RenderTemplate renders the template into unstructured objects
// (and processes template expressions).
func RenderTemplate(
	ctx context.Context,
	templateContent string,
	bindings Bindings,
) ([]unstructured.Unstructured, error) {
	parsed, err := ParseTemplate(templateContent)
	if err != nil {
		return nil, err
	}
	var rendered []unstructured.Unstructured
	for _, obj := range parsed {
		template := v1alpha1.NewProjection(obj.UnstructuredContent())
		obj, err := templating.TemplateAndMerge(ctx, compilers, obj, bindings, template)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, obj)
	}
	return rendered, nil
}

// RenderTemplateSingle renders the single-resource template into an unstructured object
// (and processes template expressions).
func RenderTemplateSingle(
	ctx context.Context,
	templateContent string,
	bindings Bindings,
) (unstructured.Unstructured, error) {
	rendered, err := RenderTemplate(ctx, templateContent, bindings)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	if len(rendered) != 1 {
		return unstructured.Unstructured{}, fmt.Errorf(errExpectedSingleResource, len(rendered))
	}
	return rendered[0], nil
}

// Match compares candidates with the expectation and returns the first match
// or an error if no match is found.
func Match(
	ctx context.Context,
	candidates []unstructured.Unstructured,
	expected unstructured.Unstructured,
	bindings Bindings,
) (unstructured.Unstructured, error) {
	var errs []error
	for _, candidate := range candidates {
		fieldErrs, err := checks.Check(ctx, compilers, candidate.UnstructuredContent(), bindings,
			ptr.To(v1alpha1.NewCheck(expected.UnstructuredContent())))
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(fieldErrs) != 0 {
			errs = append(errs,
				operrors.ResourceError(compilers, expected, candidate, true, bindings, fieldErrs),
			)
		} else {
			// Match found
			return candidate, nil
		}
	}
	return unstructured.Unstructured{}, multierr.Combine(errs...)
}

// listCandidates lists resources in the cluster that might match the expectation.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/internal.Read.
func listCandidates(
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

// Check is equivalent to a Chainsaw assert resource operation, except that it does not do polling
// or handle non-resource assertions. It returns the first matching resource on success.
// Based on github.com/kyverno/chainsaw/pkg/engine/operations/assert.Exec.
func Check(
	c client.Client,
	ctx context.Context,
	expected unstructured.Unstructured,
	bindings Bindings,
) (unstructured.Unstructured, error) {
	// Render resource metadata
	if bindings == nil {
		bindings = apis.NewBindings()
	}
	if err := templating.ResourceRef(ctx, compilers, &expected, bindings); err != nil {
		return unstructured.Unstructured{}, err
	}

	// List candidates
	candidates, err := listCandidates(c, ctx, &expected)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return unstructured.Unstructured{}, errors.New("actual resource not found")
		}
		return unstructured.Unstructured{}, err
	}
	if len(candidates) == 0 {
		return unstructured.Unstructured{}, errors.New("no actual resource found")
	}

	// Return first match
	return Match(ctx, candidates, expected, bindings)
}
