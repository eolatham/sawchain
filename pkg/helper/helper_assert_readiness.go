package helper

import "sigs.k8s.io/controller-runtime/pkg/client"

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

// AssertReadiness asserts that the specified resource is reconciled and becomes (or stays) ready within the timeout.
// Uses Chainsaw to identify the resource if given a template and optional bindings.
// Stores the state of the found resource in the given struct.
func (h *Helper) AssertReadiness(obj client.Object, opts ...AssertReadinessOption) {
	// options := NewAssertReadinessOptions(append([]AssertReadinessOption{h.Options}, opts...))
	// TODO
}
