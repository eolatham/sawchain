package link

import (
	"context"

	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SafeUpdateOption interface {
	ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions
}

type SafeUpdateOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
	Timeout         Timeout
	Interval        Interval
}

func NewSafeUpdateOptions(opts []SafeUpdateOption) SafeUpdateOptions {
	options := SafeUpdateOptions{}
	for _, opt := range opts {
		options = opt.ApplyToSafeUpdate(options)
	}
	return options
}

func (o SafeUpdateOptions) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts = o.TemplateContent.ApplyToSafeUpdate(opts)
	opts = o.TemplateFile.ApplyToSafeUpdate(opts)
	opts = o.Bindings.ApplyToSafeUpdate(opts)
	opts = o.Timeout.ApplyToSafeUpdate(opts)
	opts = o.Interval.ApplyToSafeUpdate(opts)
	return opts
}

// TODO: revise docstring
// TODO: add tests
// Update updates the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to update the resource if given a template and optional bindings.
// Stores the state of the updated resource in the given struct.
func (h *Link) SafeUpdate(ctx context.Context, obj client.Object, opts ...SafeUpdateOption) {
	// Merge options
	options := NewSafeUpdateOptions(append([]SafeUpdateOption{h.Options}, opts...))
	// Validate options
	h.validateOptions(options)
	// Parse template
	if options.TemplateContent != "" {
		h.parseTemplate(ctx, obj, options.TemplateContent, options.Bindings)
	} else if options.TemplateFile != "" {
		h.parseTemplateFile(ctx, obj, options.TemplateFile, options.Bindings)
	}
	// Update resource
	h.validateObject(obj)
	g.Expect(h.Client.Update(ctx, obj)).To(g.Succeed(), "Failed to update resource")
	// Wait for cache to sync
	updatedObj := obj.DeepCopyObject().(client.Object)
	getCachedObj := func() client.Object {
		g.Expect(h.Get(ctx, obj)).To(g.Succeed())
		return obj
	}
	g.Eventually(getCachedObj, options.Timeout, options.Interval).
		Should(g.Equal(updatedObj), "Cache not synced within timeout")
}
