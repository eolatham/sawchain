package matchers

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

// MatchYAMLMatcher is a Gomega matcher that checks if
// a client.Object matches a Chainsaw resource template.
type MatchYAMLMatcher struct {
	TemplateContent  string
	TemplateBindings map[string]any
}

// Match implements the Gomega matcher interface.
func (m *MatchYAMLMatcher) Match(actual interface{}) (bool, error) {
	// TODO: implement
	return false, fmt.Errorf("not implemented")
}

// FailureMessage implements the Gomega matcher interface.
func (m *MatchYAMLMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto match template\n\t%#v", actual, m.TemplateContent)
}

// NegatedFailureMessage implements the Gomega matcher interface.
func (m *MatchYAMLMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to match template\n\t%#v", actual, m.TemplateContent)
}

// NewMatchYAMLMatcher creates a new MatchYAMLMatcher.
func NewMatchYAMLMatcher(
	templateContent string,
	templateBindings map[string]any,
) types.GomegaMatcher {
	return &MatchYAMLMatcher{
		TemplateContent:  templateContent,
		TemplateBindings: templateBindings,
	}
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
