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

// MatchYAMLMatcher is a Gomega matcher that checks if
// a client.Object matches a Chainsaw resource template.
type MatchYAMLMatcher struct {
	c               client.Client
	templateContent string
	bindings        chainsaw.Bindings
	expected        unstructured.Unstructured
	matchError      error
}

// Match implements the Gomega matcher interface.
func (m *MatchYAMLMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, errors.New("MatchYAMLMatcher expects a client.Object but got nil")
	}
	obj, ok := utilities.AsObject(actual)
	if !ok {
		return false, fmt.Errorf("MatchYAMLMatcher expects a client.Object but got %T", actual)
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
func (m *MatchYAMLMatcher) FailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nto match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NegatedFailureMessage implements the Gomega matcher interface.
func (m *MatchYAMLMatcher) NegatedFailureMessage(actual interface{}) string {
	baseMessage := fmt.Sprintf("Expected\n\t%#v\nnot to match template\n\t%#v", actual, m.templateContent)
	if m.matchError != nil {
		return fmt.Sprintf("%s: %v", baseMessage, m.matchError)
	}
	return baseMessage
}

// NewMatchYAMLMatcher creates a new MatchYAMLMatcher.
func NewMatchYAMLMatcher(
	c client.Client,
	templateContent string,
	bindings map[string]any,
) (types.GomegaMatcher, error) {
	expected, err := chainsaw.ParseTemplateSingle(templateContent)
	if err != nil {
		return nil, err
	}
	matcher := &MatchYAMLMatcher{
		c:               c,
		templateContent: templateContent,
		bindings:        chainsaw.BindingsFromMap(bindings),
		expected:        expected,
	}
	return matcher, nil
}

// StatusConditionMatcher is a Gomega matcher that checks if
// a client.Object has a specific status condition.
type StatusConditionMatcher struct {
	ConditionType  string
	ExpectedStatus string
}

// Match implements the Gomega matcher interface.
func (m *StatusConditionMatcher) Match(actual interface{}) (bool, error) {
	// TODO: implement
	return false, fmt.Errorf("not implemented")
}

// FailureMessage implements the Gomega matcher interface.
func (m *StatusConditionMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto have status condition %s=%s",
		actual, m.ConditionType, m.ExpectedStatus)
}

// NegatedFailureMessage implements the Gomega matcher interface.
func (m *StatusConditionMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to have status condition %s=%s",
		actual, m.ConditionType, m.ExpectedStatus)
}

// NewStatusConditionMatcher creates a new StatusConditionMatcher.
func NewStatusConditionMatcher(conditionType, expectedStatus string) types.GomegaMatcher {
	return &StatusConditionMatcher{
		ConditionType:  conditionType,
		ExpectedStatus: expectedStatus,
	}
}
