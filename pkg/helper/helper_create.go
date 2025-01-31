package helper

import "sigs.k8s.io/controller-runtime/pkg/client"

type CreateOption interface {
	ApplyToCreate(opts CreateOptions) CreateOptions
}

type CreateOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewCreateOptions(opts []CreateOption) CreateOptions {
	options := CreateOptions{}
	for _, opt := range opts {
		options = opt.ApplyToCreate(options)
	}
	return options
}

func (o CreateOptions) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts = o.Timeout.ApplyToCreate(opts)
	opts = o.Interval.ApplyToCreate(opts)
	opts = o.Template.ApplyToCreate(opts)
	opts = o.Bindings.ApplyToCreate(opts)
	return opts
}

// Create creates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to create the resource if given a template and optional bindings.
// Stores the state of the created resource in the given struct.
func (h *Helper) Create(obj client.Object, opts ...CreateOption) {
	// options := NewCreateOptions(append([]CreateOption{h.Options}, opts...))
	// TODO
}
