package link

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetOption interface {
	ApplyToGet(opts GetOptions) GetOptions
}

type GetOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
}

func NewGetOptions(opts []GetOption) GetOptions {
	options := GetOptions{}
	for _, opt := range opts {
		options = opt.ApplyToGet(options)
	}
	return options
}

func (o GetOptions) ApplyToGet(opts GetOptions) GetOptions {
	opts = o.TemplateContent.ApplyToGet(opts)
	opts = o.TemplateFile.ApplyToGet(opts)
	opts = o.Bindings.ApplyToGet(opts)
	return opts
}

func (h *Link) Get(ctx context.Context, obj client.Object, opts ...GetOption) func() error {
	// Process options
	options := NewGetOptions(append([]GetOption{h.Options}, opts...))
	h.validateOptions(options)
	// TODO
	return func() error {
		return h.Client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	}
}
