package helper

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HelperOption interface {
	ApplyToHelper(opts HelperOptions) HelperOptions
}

type HelperOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
}

func NewHelperOptions(opts []HelperOption) HelperOptions {
	options := HelperOptions{}
	for _, opt := range opts {
		options = opt.ApplyToHelper(options)
	}
	return options
}

func (o HelperOptions) ApplyToHelper(opts HelperOptions) HelperOptions {
	opts = o.Timeout.ApplyToHelper(opts)
	opts = o.Interval.ApplyToHelper(opts)
	opts = o.Bindings.ApplyToHelper(opts)
	return opts
}

func (o HelperOptions) ApplyToCreate(opts CreateOptions) CreateOptions {
	opts = o.Timeout.ApplyToCreate(opts)
	opts = o.Interval.ApplyToCreate(opts)
	opts = o.Bindings.ApplyToCreate(opts)
	return opts
}

func (o HelperOptions) ApplyToUpdate(opts UpdateOptions) UpdateOptions {
	opts = o.Timeout.ApplyToUpdate(opts)
	opts = o.Interval.ApplyToUpdate(opts)
	opts = o.Bindings.ApplyToUpdate(opts)
	return opts
}

func (o HelperOptions) ApplyToDelete(opts DeleteOptions) DeleteOptions {
	opts = o.Timeout.ApplyToDelete(opts)
	opts = o.Interval.ApplyToDelete(opts)
	return opts
}

func (o HelperOptions) ApplyToAssertReadiness(opts AssertReadinessOptions) AssertReadinessOptions {
	opts = o.Timeout.ApplyToAssertReadiness(opts)
	opts = o.Interval.ApplyToAssertReadiness(opts)
	return opts
}

func (o HelperOptions) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts = o.Timeout.ApplyToAssertResource(opts)
	opts = o.Interval.ApplyToAssertResource(opts)
	opts = o.Bindings.ApplyToAssertResource(opts)
	return opts
}

func (o HelperOptions) ApplyToAssertDeletion(opts AssertDeletionOptions) AssertDeletionOptions {
	opts = o.Timeout.ApplyToAssertDeletion(opts)
	opts = o.Interval.ApplyToAssertDeletion(opts)
	return opts
}

type Helper struct {
	Client  client.Client
	Context context.Context
	Options HelperOptions
}

func NewHelper(c client.Client, ctx context.Context, opts ...HelperOption) Helper {
	return Helper{
		Client:  c,
		Context: ctx,
		Options: NewHelperOptions(opts),
	}
}
