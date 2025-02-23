package link

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GetObjectOption interface {
	ApplyToGetObject(opts GetObjectOptions) GetObjectOptions
}

type GetObjectOptions struct {
	Template Template
	Bindings Bindings
}

func NewGetObjectOptions(opts []GetObjectOption) GetObjectOptions {
	options := GetObjectOptions{}
	for _, opt := range opts {
		options = opt.ApplyToGetObject(options)
	}
	return options
}

func (o GetObjectOptions) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts = o.Template.ApplyToGetObject(opts)
	opts = o.Bindings.ApplyToGetObject(opts)
	return opts
}

// TODO
func (h *Link) GetObject(ctx context.Context, obj client.Object, opts ...GetOption) func() client.Object {
	// Merge options
	// options := NewGetObjectOptions(append([]GetObjectOption{h.Options}, opts...))
	return func() client.Object {
		h.Client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
		return obj
	}
}
