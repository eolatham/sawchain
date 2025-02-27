package link

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetObjectOption interface {
	ApplyToGetObject(opts GetObjectOptions) GetObjectOptions
}

type GetObjectOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
}

func NewGetObjectOptions(opts []GetObjectOption) GetObjectOptions {
	options := GetObjectOptions{}
	for _, opt := range opts {
		options = opt.ApplyToGetObject(options)
	}
	return options
}

func (o GetObjectOptions) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts = o.TemplateContent.ApplyToGetObject(opts)
	opts = o.TemplateFile.ApplyToGetObject(opts)
	opts = o.Bindings.ApplyToGetObject(opts)
	return opts
}

func (h *Link) GetObject(ctx context.Context, obj client.Object, opts ...GetObjectOption) func() client.Object {
	// Process options
	options := NewGetObjectOptions(append([]GetObjectOption{h.Options}, opts...))
	h.validateOptions(options)
	// Parse template
	if options.TemplateContent != "" {
		h.parseTemplate(ctx, obj, options.TemplateContent, options.Bindings)
	} else if options.TemplateFile != "" {
		h.parseTemplateFile(ctx, obj, options.TemplateFile, options.Bindings)
	}
	// Validate object
	h.validateObject(obj)
	// Return get function
	return func() client.Object {
		h.Client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		return obj
	}
}
