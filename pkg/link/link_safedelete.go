package link

import (
	"context"

	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SafeDeleteOption interface {
	ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions
}

type SafeDeleteOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
	Timeout         Timeout
	Interval        Interval
}

func NewSafeDeleteOptions(opts []SafeDeleteOption) SafeDeleteOptions {
	options := SafeDeleteOptions{}
	for _, opt := range opts {
		options = opt.ApplyToSafeDelete(options)
	}
	return options
}

func (o SafeDeleteOptions) ApplyToDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts = o.TemplateContent.ApplyToSafeDelete(opts)
	opts = o.TemplateFile.ApplyToSafeDelete(opts)
	opts = o.Bindings.ApplyToSafeDelete(opts)
	opts = o.Timeout.ApplyToSafeDelete(opts)
	opts = o.Interval.ApplyToSafeDelete(opts)
	return opts
}

// TODO: revise docstring
// TODO: add tests
// Delete deletes the specified resource and ensures the client cache is synced within the timeout.
// Uses Chainsaw to delete the resource if given a template and optional bindings.
// Stores the state of the deleted resource in the given struct.
func (h *Link) SafeDelete(ctx context.Context, obj client.Object, opts ...SafeDeleteOption) {
	// Merge options
	options := NewSafeDeleteOptions(append([]SafeDeleteOption{h.Options}, opts...))
	// Validate options
	h.validateOptions(options)
	// Parse template
	if options.TemplateContent != "" {
		h.parseTemplate(ctx, obj, options.TemplateContent, options.Bindings)
	} else if options.TemplateFile != "" {
		h.parseTemplateFile(ctx, obj, options.TemplateFile, options.Bindings)
	}
	// Delete resource
	h.validateObject(obj)
	g.Expect(client.IgnoreNotFound(h.Client.Delete(ctx, obj))).
		To(g.Succeed(), "Failed to delete resource")
	// Wait for cache for sync
	g.Eventually(h.Get(ctx, obj), options.Timeout, options.Interval).
		ShouldNot(g.Succeed(), "Cache not synced within timeout")
}
