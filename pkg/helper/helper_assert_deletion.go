package helper

import "sigs.k8s.io/controller-runtime/pkg/client"

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

// AssertDeletion asserts that the specified resource is deleted within the timeout.
// Uses Chainsaw to identify the resource if given a template and optional bindings.
// Stores the key of the deleted resource in the given struct.
func (h *Helper) AssertDeletion(obj client.Object, opts ...AssertDeletionOption) {
	// options := NewAssertDeletionOptions(append([]AssertDeletionOption{h.Options}, opts...))
	// TODO
}
