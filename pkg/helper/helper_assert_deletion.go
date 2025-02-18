package helper

import (
	g "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AssertDeletionOption interface {
	ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions
}

type AssertDeletionOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewAssertDeletionOptions(opts []AssertDeletionOption) AssertDeletionOptions {
	options := AssertDeletionOptions{}
	for _, opt := range opts {
		options = opt.ApplyToAssertDeletion(options)
	}
	return options
}

func (o AssertDeletionOptions) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts = o.Timeout.ApplyToAssertDeletion(opts)
	opts = o.Interval.ApplyToAssertDeletion(opts)
	return opts
}

// TODO: revise docstring
// TODO: add tests
// AssertDeletion asserts that the specified resource is deleted within the timeout.
// Uses Chainsaw to identify the resource if given a template and optional bindings.
// Stores the key of the deleted resource in the given struct.
func (h *Helper) AssertDeletion(obj client.Object, opts ...AssertDeletionOption) {
	// Merge options
	options := NewAssertDeletionOptions(append([]AssertDeletionOption{h.Options}, opts...))
	// Parse template
	if options.Template != "" {
		h.parse(obj, options.Template, options.Bindings)
	}
	// Wait for deletion
	h.validateForCrud(obj)
	isDeleted := func() bool {
		if err := h.get(obj); apierrors.IsNotFound(err) {
			return true
		} else if err != nil {
			g.Expect(err).NotTo(g.HaveOccurred(), "Unexpected error checking resource deletion")
		}
		return false
	}
	g.Eventually(isDeleted, options.Timeout, options.Interval).
		Should(g.BeTrue(), "Resource not deleted within timeout")
}
