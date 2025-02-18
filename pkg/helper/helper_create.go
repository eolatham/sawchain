package helper

import (
	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

// TODO: revise docstring
// TODO: add tests
// Create creates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to create the resource if given a template and optional bindings.
// Stores the state of the created resource in the given struct.
func (h *Helper) Create(obj client.Object, opts ...CreateOption) {
	// Merge options
	options := NewCreateOptions(append([]CreateOption{h.Options}, opts...))
	// Parse template
	if options.Template != "" {
		h.parse(obj, options.Template, options.Bindings)
	}
	// Create resource
	h.validateForCrud(obj)
	g.Expect(h.Client.Create(h.Context, obj)).
		To(g.Succeed(), "Failed to create resource")
	// Wait for cache to sync
	g.Eventually(h.getFunc(obj), options.Timeout, options.Interval).
		Should(g.Succeed(), "Cache not synced within timeout")
}
