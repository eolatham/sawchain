package helper

import (
	"os"

	g "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"k8s-test-helper/pkg/chainsaw"
)

type AssertReadinessOption interface {
	ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions
}

type AssertReadinessOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewAssertReadinessOptions(opts []AssertReadinessOption) AssertReadinessOptions {
	options := AssertReadinessOptions{}
	for _, opt := range opts {
		options = opt.ApplyToAssertReadiness(options)
	}
	return options
}

func (o AssertReadinessOptions) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts = o.Timeout.ApplyToAssertReadiness(opts)
	opts = o.Interval.ApplyToAssertReadiness(opts)
	return opts
}

// TODO: revise docstring
// TODO: add tests
// AssertReadiness asserts that the specified resource is reconciled and becomes (or stays) ready within the timeout.
// Uses Chainsaw to identify the resource if given a template and optional bindings.
// Stores the state of the found resource in the given struct.
func (h *Helper) AssertReadiness(obj client.Object, opts ...AssertReadinessOption) {
	// Merge options
	options := NewAssertReadinessOptions(append([]AssertReadinessOption{h.Options}, opts...))
	// Parse template
	if options.Template != "" {
		h.parse(obj, options.Template, options.Bindings)
	}
	// Validate object
	h.validateForCrud(obj)
	// Marshal identifying metadata
	minimalObj := toMinimalObject(obj)
	yamlData, err := yaml.Marshal(minimalObj)
	g.Expect(err).NotTo(g.HaveOccurred(), "Internal error: failed to marshal object")
	g.Expect(yamlData).NotTo(g.BeEmpty(), "Internal error: marshaled object data is empty")
	// Append status assertion
	chainsawAssertion := `
status:
  (conditions[?type == 'Ready']):
    - status: "True"
`
	finalYaml := string(yamlData) + chainsawAssertion
	// Create temporary file
	tempFile, err := os.CreateTemp("", "assert-ready-*.yaml")
	g.Expect(err).NotTo(g.HaveOccurred(), "Internal error: failed to create temporary file")
	g.Expect(tempFile).NotTo(g.BeNil(), "Internal error: created temporary file is nil")
	defer os.Remove(tempFile.Name())
	_, err = tempFile.Write([]byte(finalYaml))
	g.Expect(err).NotTo(g.HaveOccurred(), "Internal error: failed to write to temporary file")
	g.Expect(tempFile.Close()).To(g.Succeed(), "Internal error: failed to close temporary file")
	// Use Chainsaw to make assertion
	check := func() error {
		var err error
		// Save match to object variable
		obj, err = chainsaw.CheckResource(h.Client, h.Context, tempFile.Name(), nil)
		return err
	}
	g.Eventually(check, options.Timeout, options.Interval).
		Should(g.Succeed(), "Chainsaw assertion never succeeded")
}

// minimalObject represents the essential fields for identifying a Kubernetes object.
type minimalObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

// toMinimalObject extracts only the identifying fields from a client.Object.
func toMinimalObject(obj client.Object) minimalObject {
	return minimalObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
	}
}
