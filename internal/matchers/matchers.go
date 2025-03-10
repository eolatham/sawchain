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

// TODO: test

// ChainsawMatcher is a Gomega matcher that checks if
// a client.Object matches a Chainsaw resource template.
type ChainsawMatcher struct {
	c               client.Client
	templateContent string
	bindings        chainsaw.Bindings
	expected        unstructured.Unstructured
	matchError      error
}

// Match implements the Gomega matcher interface.
func (m *ChainsawMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, errors.New("ChainsawMatcher expects a client.Object but got nil")
	}
	obj, ok := utilities.AsObject(actual)
	if !ok {
		return false, fmt.Errorf("ChainsawMatcher expects a client.Object but got %T", actual)
	}
	candidate, err := utilities.UnstructuredFromObject(m.c, obj)
	if err != nil {
		return false, err
	}
	m.matchError = nil
	_, m.matchError = chainsaw.Match(context.TODO(), []unstructured.Unstructured{candidate}, m.expected, m.bindings)
	return m.matchError == nil, nil
}

// FailureMessage implements the Gomega matcher interface.
func (m *ChainsawMatcher) FailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nto match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NegatedFailureMessage implements the Gomega matcher interface.
func (m *ChainsawMatcher) NegatedFailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nnot to match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NewChainsawMatcher creates a new ChainsawMatcher.
func NewChainsawMatcher(
	c client.Client,
	templateContent string,
	bindings map[string]any,
) (types.GomegaMatcher, error) {
	expected, err := chainsaw.ParseTemplateSingle(templateContent)
	if err != nil {
		return nil, err
	}
	matcher := &ChainsawMatcher{
		c:               c,
		templateContent: templateContent,
		bindings:        chainsaw.BindingsFromMap(bindings),
		expected:        expected,
	}
	return matcher, nil
}

// NewStatusConditionMatcher creates a new ChainsawMatcher that checks
// if resources have the expected status condition.
func NewStatusConditionMatcher(
	c client.Client,
	conditionType, expectedStatus string,
) (types.GomegaMatcher, error) {
	templateContent := `
status:
	(conditions[?type == $conditionType]):
	- status: ($expectedStatus)
`
	bindings := map[string]any{
		"conditionType":  conditionType,
		"expectedStatus": expectedStatus,
	}
	return NewChainsawMatcher(c, templateContent, bindings)
}
