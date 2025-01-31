package helper

import "sigs.k8s.io/controller-runtime/pkg/client"

type AssertResourceOption interface {
	ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions
}

type AssertResourceOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewAssertResourceOptions(opts []AssertResourceOption) AssertResourceOptions {
	options := AssertResourceOptions{}
	for _, opt := range opts {
		options = opt.ApplyToAssertResource(options)
	}
	return options
}

func (o AssertResourceOptions) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts = o.Timeout.ApplyToAssertResource(opts)
	opts = o.Interval.ApplyToAssertResource(opts)
	opts = o.Template.ApplyToAssertResource(opts)
	opts = o.Bindings.ApplyToAssertResource(opts)
	return opts
}

// AssertResource asserts that the specified resource exists with the expected state within the timeout.
// Uses Chainsaw to make assertions if given a template and optional bindings.
// Stores the state of the found resource in the given struct.
func (h *Helper) AssertResource(obj client.Object, opts ...AssertResourceOption) {
	// options := NewAssertResourceOptions(append([]AssertResourceOption{h.Options}, opts...))
	// TODO
}
