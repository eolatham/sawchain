package matchers

import (
	"context"
	"errors"
	"fmt"

	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/chainsaw"
	"github.com/eolatham/sawchain/internal/utilities"
)

// chainsawMatcher is a Gomega matcher that checks if
// a client.Object matches a Chainsaw resource template.
type chainsawMatcher struct {
	// K8s client used for type conversions
	c client.Client
	// Function to create template content
	createTemplateContent func(c client.Client, obj client.Object) (string, error)
	// Current template content
	templateContent string
	// Template bindings
	bindings chainsaw.Bindings
	// Current match error
	matchError error
}

// Match implements the Gomega matcher interface.
func (m *chainsawMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, errors.New("chainsawMatcher expects a client.Object but got nil")
	}
	obj, ok := utilities.AsObject(actual)
	if !ok {
		return false, fmt.Errorf("chainsawMatcher expects a client.Object but got %T", actual)
	}
	candidate, err := utilities.UnstructuredFromObject(m.c, obj)
	if err != nil {
		return false, err
	}
	m.templateContent, err = m.createTemplateContent(m.c, obj)
	if err != nil {
		return false, err
	}
	expected, err := chainsaw.ParseTemplateSingle(m.templateContent)
	if err != nil {
		return false, err
	}
	_, m.matchError = chainsaw.Match(context.TODO(), []unstructured.Unstructured{candidate}, expected, m.bindings)
	return m.matchError == nil, nil
}

// FailureMessage implements the Gomega matcher interface.
func (m *chainsawMatcher) FailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nto match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NegatedFailureMessage implements the Gomega matcher interface.
func (m *chainsawMatcher) NegatedFailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nnot to match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NewChainsawMatcher creates a new chainsawMatcher with static template content.
func NewChainsawMatcher(
	c client.Client,
	templateContent string,
	bindings map[string]any,
) types.GomegaMatcher {
	return &chainsawMatcher{
		c: c,
		createTemplateContent: func(c client.Client, obj client.Object) (string, error) {
			return templateContent, nil
		},
		bindings: chainsaw.BindingsFromMap(bindings),
	}
}

// NewStatusConditionMatcher creates a new chainsawMatcher that checks
// if resources have the expected status condition.
func NewStatusConditionMatcher(
	c client.Client,
	conditionType,
	expectedStatus string,
) types.GomegaMatcher {
	return &chainsawMatcher{
		c: c,
		createTemplateContent: func(c client.Client, obj client.Object) (string, error) {
			// Extract apiVersion and kind from object
			gvk, err := utilities.GetGroupVersionKind(obj, c.Scheme())
			if err != nil {
				return "", fmt.Errorf("failed to create template content: %w", err)
			}
			apiVersion := gvk.GroupVersion().String()
			kind := gvk.Kind
			// Create template content
			templateContent := fmt.Sprintf(`
apiVersion: %s
kind: %s
status:
  (conditions[?type == '%s']):
  - status: '%s'`,
				apiVersion,
				kind,
				conditionType,
				expectedStatus,
			)
			return templateContent, nil
		},
	}
}
