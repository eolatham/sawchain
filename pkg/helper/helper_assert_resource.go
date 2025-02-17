package helper

import (
	g "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s-test-helper/pkg/chainsaw"
)

type AssertResourceOption interface {
	ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions
}

type AssertResourceOptions struct {
	Timeout  Timeout
	Interval Interval
	Bindings Bindings
	Template Template
}

func NewAssertResourceOptions(opts []AssertResourceOption) AssertResourceOptions {
	options := AssertResourceOptions{}
	for _, opt := range opts {
		options = opt.ApplyToAssertResource(options)
	}
	return options
}

func (o AssertResourceOptions) ApplyToAssertResource(opts AssertResourceOptions) AssertResourceOptions {
	opts = o.Timeout.ApplyToAssertResource(opts)
	opts = o.Interval.ApplyToAssertResource(opts)
	opts = o.Template.ApplyToAssertResource(opts)
	opts = o.Bindings.ApplyToAssertResource(opts)
	return opts
}

// TODO: test
// AssertResource asserts that the specified resource exists with the expected state within the timeout.
// Uses Chainsaw to make assertions if given a template and optional bindings.
// Otherwise, asserts an exact match (all-inclusive) for the given object.
// Stores the state of the found resource in the given struct.
func (h *Helper) AssertResource(obj client.Object, opts ...AssertResourceOption) {
	// Merge options
	options := NewAssertResourceOptions(append([]AssertResourceOption{h.Options}, opts...))
	// Execute assertion
	if options.Template != "" {
		// Chainsaw assertion
		h.TB.Logf(
			"Executing Chainsaw assertion with template %s and bindings %+v",
			options.Template, options.Bindings,
		)
		check := func() error {
			var err error
			// Save match to object variable
			obj, err = chainsaw.CheckResource(
				h.Client, h.Context, string(options.Template), options.Bindings,
			)
			return err
		}
		g.Eventually(check, options.Timeout, options.Interval).
			Should(g.Succeed(), "Chainsaw assertion never succeeded")
	} else {
		// Non-Chainsaw assertion (exact match)
		h.TB.Logf("Asserting exact match for object %+v", obj)
		h.validateForCrud(obj)
		expectedObj := obj.DeepCopyObject().(client.Object)
		getObj := func() client.Object {
			if err := h.get(obj); apierrors.IsNotFound(err) {
				// Ignore not found errors
				return nil
			} else if err != nil {
				// Fail on other errors
				g.Expect(err).NotTo(g.HaveOccurred(), "Unexpected error checking resource")
			}
			return obj
		}
		g.Eventually(getObj, options.Timeout, options.Interval).
			Should(g.Equal(expectedObj), "Exact match never found")
	}
}
