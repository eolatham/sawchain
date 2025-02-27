package link

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CheckOption interface {
	ApplyToCheck(opts CheckOptions) CheckOptions
}

type CheckOptions struct {
	TemplateContent TemplateContent
	TemplateFile    TemplateFile
	Bindings        Bindings
}

func NewCheckOptions(opts []CheckOption) CheckOptions {
	options := CheckOptions{}
	for _, opt := range opts {
		options = opt.ApplyToCheck(options)
	}
	return options
}

func (o CheckOptions) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts = o.TemplateContent.ApplyToCheck(opts)
	opts = o.TemplateFile.ApplyToCheck(opts)
	opts = o.Bindings.ApplyToCheck(opts)
	return opts
}

func (h *Link) Check(ctx context.Context, obj client.Object, opts ...CheckOption) error {
	// Process options
	options := NewCheckOptions(append([]CheckOption{h.Options}, opts...))
	h.validateOptions(options)
	h.requireTemplate(options)
	// TODO
	return nil
}
