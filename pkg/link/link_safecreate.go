package link

import (
	"context"

	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SafeCreateOption interface {
	ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions
}

type SafeCreateOptions struct {
	Template Template
	Bindings Bindings
	Timeout  Timeout
	Interval Interval
}

func NewSafeCreateOptions(opts []SafeCreateOption) SafeCreateOptions {
	options := SafeCreateOptions{}
	for _, opt := range opts {
		options = opt.ApplyToSafeCreate(options)
	}
	return options
}

func (o SafeCreateOptions) ApplyToCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts = o.Template.ApplyToSafeCreate(opts)
	opts = o.Bindings.ApplyToSafeCreate(opts)
	opts = o.Timeout.ApplyToSafeCreate(opts)
	opts = o.Interval.ApplyToSafeCreate(opts)
	return opts
}

// TODO: revise docstring
// TODO: add tests
// Create creates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to create the resource if given a template and optional bindings.
// Stores the state of the created resource in the given struct.
func (h *Link) SafeCreate(ctx context.Context, obj client.Object, opts ...SafeCreateOption) {
	// Merge options
	options := NewSafeCreateOptions(append([]SafeCreateOption{h.Options}, opts...))
	// Parse template
	if options.Template != "" {
		h.parseTemplate(ctx, obj, options.Template, options.Bindings)
	}
	// Create resource
	h.validateObject(obj)
	g.Expect(h.Client.Create(ctx, obj)).
		To(g.Succeed(), "Failed to create resource")
	// Wait for cache to sync
	g.Eventually(h.Get(ctx, obj), options.Timeout, options.Interval).
		Should(g.Succeed(), "Cache not synced within timeout")
}
