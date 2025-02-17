package helper

import (
	"context"
	"testing"

	g "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-test-helper/pkg/chainsaw"
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
	TB      testing.TB
	Client  client.Client
	Context context.Context
	Options HelperOptions
}

func NewHelper(tb testing.TB, c client.Client, ctx context.Context, opts ...HelperOption) Helper {
	return Helper{
		TB:      tb,
		Client:  c,
		Context: ctx,
		Options: NewHelperOptions(opts),
	}
}

// get gets the specified resource.
func (h *Helper) get(obj client.Object) error {
	return h.Client.Get(h.Context, client.ObjectKeyFromObject(obj), obj)
}

// getFunc returns a function that gets the specified resource.
func (h *Helper) getFunc(obj client.Object) func() error {
	return func() error { return h.get(obj) }
}

// parse parses the template and saves its structured content to the object.
func (h *Helper) parse(obj client.Object, template Template, bindings Bindings) {
	if obj != nil {
		h.TB.Logf("Overwriting object with template content")
	}
	var err error
	obj, err = chainsaw.ParseResource(h.Client, h.Context, string(template), bindings)
	g.Expect(err).NotTo(g.HaveOccurred(), "Failed to parse template")
	g.Expect(obj).NotTo(g.BeNil(), "Parsed object is nil")
}

// validateForCrud asserts that the object is not nil and that it has a name.
func (h *Helper) validateForCrud(obj client.Object) {
	g.Expect(obj).NotTo(g.BeNil(), "Object must not be nil")
	g.Expect(obj.GetName()).NotTo(g.BeEmpty(), "Object must have a name")
}
