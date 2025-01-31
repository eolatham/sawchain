package chainsaw

import (
	"context"

	"github.com/kyverno/chainsaw/pkg/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// getResources attempts to get all resources from the cluster that match the expected resource.
// It is based on github.com/kyverno/chainsaw/pkg/engine/operations/internal.Read.
func getResources(c client.Client, ctx context.Context, expected client.Object) ([]unstructured.Unstructured, error) {
	var results []unstructured.Unstructured
	gvk := expected.GetObjectKind().GroupVersionKind()
	useGet := expected.GetName() != ""
	if useGet {
		var actual unstructured.Unstructured
		actual.SetGroupVersionKind(gvk)
		if err := c.Get(ctx, client.Key(expected), &actual); err != nil {
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
