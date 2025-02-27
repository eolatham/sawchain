package link

import (
	"context"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SafeCreateOption interface {
	ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions
}

type SafeCreateOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
	Timeout         Timeout
	Interval        Interval
}

func NewSafeCreateOptions(opts []SafeCreateOption) SafeCreateOptions {
	options := SafeCreateOptions{}
	for _, opt := range opts {
		options = opt.ApplyToSafeCreate(options)
	}
	return options
}

func (o SafeCreateOptions) ApplyToCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts = o.TemplateContent.ApplyToSafeCreate(opts)
	opts = o.TemplateFile.ApplyToSafeCreate(opts)
	opts = o.Bindings.ApplyToSafeCreate(opts)
	opts = o.Timeout.ApplyToSafeCreate(opts)
	opts = o.Interval.ApplyToSafeCreate(opts)
	return opts
}

// Create creates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to create the resource if given a template and optional bindings.
// Stores the state of the created resource in the given struct.
func (h *Link) SafeCreate(ctx context.Context, obj client.Object, opts ...SafeCreateOption) {
	// Process options
	options := NewSafeCreateOptions(append([]SafeCreateOption{h.Options}, opts...))
	h.validateOptions(options)
	// Parse template
	if options.TemplateContent != "" {
		h.parseTemplate(ctx, obj, options.TemplateContent, options.Bindings)
	} else if options.TemplateFile != "" {
		h.parseTemplateFile(ctx, obj, options.TemplateFile, options.Bindings)
	}
	// Create resource
	h.validateObject(obj)
	h.Gomega.Expect(h.Client.Create(ctx, obj)).
		To(gomega.Succeed(), "Failed to create resource")
	// Wait for cache to sync
	h.Gomega.Eventually(h.Get(ctx, obj), options.Timeout, options.Interval).
		Should(gomega.Succeed(), "Cache not synced within timeout")
}
