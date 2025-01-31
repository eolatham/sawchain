package helper

import "sigs.k8s.io/controller-runtime/pkg/client"

type UpdateOption interface {
	ApplyToUpdate(opts UpdateOptions) UpdateOptions
}

type UpdateOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewUpdateOptions(opts []UpdateOption) UpdateOptions {
	options := UpdateOptions{}
	for _, opt := range opts {
		options = opt.ApplyToUpdate(options)
	}
	return options
}

func (o UpdateOptions) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts = o.Timeout.ApplyToUpdate(opts)
	opts = o.Interval.ApplyToUpdate(opts)
	opts = o.Template.ApplyToUpdate(opts)
	opts = o.Bindings.ApplyToUpdate(opts)
	return opts
}

// Update updates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to update the resource if given a template and optional bindings.
// Stores the state of the updated resource in the given struct.
func (h *Helper) Update(obj client.Object, opts ...UpdateOption) {
	// options := NewUpdateOptions(append([]UpdateOption{h.Options}, opts...))
	// TODO
}
