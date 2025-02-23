package link

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetOption interface {
	ApplyToGet(opts GetOptions) GetOptions
}

type GetOptions struct {
	Template Template
	Bindings Bindings
}

func NewGetOptions(opts []GetOption) GetOptions {
	options := GetOptions{}
	for _, opt := range opts {
		options = opt.ApplyToGet(options)
	}
	return options
}

func (o GetOptions) ApplyToGet(opts GetOptions) GetOptions {
	opts = o.Template.ApplyToGet(opts)
	opts = o.Bindings.ApplyToGet(opts)
	return opts
}

// TODO
func (h *Link) Get(ctx context.Context, obj client.Object, opts ...GetOption) func() error {
	// Merge options
	// options := NewGetOptions(append([]GetOption{h.Options}, opts...))
	return func() error {
		return h.Client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	}
}
