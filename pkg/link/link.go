package link

import (
	"context"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/pkg/chainsaw"
)

type LinkOption interface {
	ApplyToLink(opts LinkOptions) LinkOptions
}

type LinkOptions struct {
	Bindings Bindings
	Timeout  Timeout
	Interval Interval
}

func NewLinkOptions(opts []LinkOption) LinkOptions {
	options := LinkOptions{}
	for _, opt := range opts {
		options = opt.ApplyToLink(options)
	}
	return options
}

func (o LinkOptions) ApplyToLink(opts LinkOptions) LinkOptions {
	opts = o.Bindings.ApplyToLink(opts)
	opts = o.Timeout.ApplyToLink(opts)
	opts = o.Interval.ApplyToLink(opts)
	return opts
}

func (o LinkOptions) ApplyToSafeCreate(opts SafeCreateOptions) SafeCreateOptions {
	opts = o.Bindings.ApplyToSafeCreate(opts)
	opts = o.Timeout.ApplyToSafeCreate(opts)
	opts = o.Interval.ApplyToSafeCreate(opts)
	return opts
}

func (o LinkOptions) ApplyToSafeUpdate(opts SafeUpdateOptions) SafeUpdateOptions {
	opts = o.Bindings.ApplyToSafeUpdate(opts)
	opts = o.Timeout.ApplyToSafeUpdate(opts)
	opts = o.Interval.ApplyToSafeUpdate(opts)
	return opts
}

func (o LinkOptions) ApplyToSafeDelete(opts SafeDeleteOptions) SafeDeleteOptions {
	opts = o.Bindings.ApplyToSafeDelete(opts)
	opts = o.Timeout.ApplyToSafeDelete(opts)
	opts = o.Interval.ApplyToSafeDelete(opts)
	return opts
}

func (o LinkOptions) ApplyToGet(opts GetOptions) GetOptions {
	opts = o.Bindings.ApplyToGet(opts)
	return opts
}

func (o LinkOptions) ApplyToGetObject(opts GetObjectOptions) GetObjectOptions {
	opts = o.Bindings.ApplyToGetObject(opts)
	return opts
}

func (o LinkOptions) ApplyToCheck(opts CheckOptions) CheckOptions {
	opts = o.Bindings.ApplyToCheck(opts)
	return opts
}

type Link struct {
	T       testing.TB
	Gomega  gomega.Gomega
	Client  client.Client
	Options LinkOptions
}

func NewLink(t testing.TB, c client.Client, opts ...LinkOption) Link {
	return Link{
		T:       t,
		Gomega:  gomega.NewWithT(t),
		Client:  c,
		Options: NewLinkOptions(opts),
	}
}

// validateOptions executes assertions common for all operations on input options.
func (h *Link) validateOptions(options interface{}) {
	h.Gomega.Expect(options).NotTo(gomega.And(
		gomega.HaveField("TemplateContent", gomega.Not(gomega.BeEmpty())),
		gomega.HaveField("TemplateFile", gomega.Not(gomega.BeEmpty()))),
		"Invalid options: TemplateContent and TemplateFile are mutually exclusive")
}

// requireTemplate asserts that either TemplateContent or TemplateFile is provided (mutually exclusive).
func (h *Link) requireTemplate(options interface{}) {
	h.Gomega.Expect(options).To(gomega.Or(
		gomega.And(
			gomega.HaveField("TemplateContent", gomega.Not(gomega.BeEmpty())),
			gomega.HaveField("TemplateFile", gomega.BeEmpty())),
		gomega.And(
			gomega.HaveField("TemplateContent", gomega.BeEmpty()),
			gomega.HaveField("TemplateFile", gomega.Not(gomega.BeEmpty())))),
		"Invalid options: expected either TemplateContent or TemplateFile to be provided (mutually exclusive)")
}

// readTemplateFile reads the template file, asserts it is not empty, and returns its content.
func (h *Link) readTemplateFile(templateFile TemplateFile) TemplateContent {
	content, err := os.ReadFile(string(templateFile))
	h.Gomega.Expect(err).NotTo(gomega.BeNil(), "Failed to read template file")
	h.Gomega.Expect(content).NotTo(gomega.BeEmpty(), "Template file content is empty")
	return TemplateContent(string(content))
}

// parseTemplate parses the template and saves its structured content to the object.
func (h *Link) parseTemplate(
	ctx context.Context,
	obj client.Object,
	template TemplateContent,
	bindings Bindings,
) {
	h.Gomega.Expect(obj).NotTo(gomega.BeNil(), "Given object must be an actual pointer in order to save state")
	var err error
	obj, err = chainsaw.ParseResource(h.Client, ctx, string(template), bindings)
	h.Gomega.Expect(err).NotTo(gomega.BeNil(), "Failed to parse template")
	h.Gomega.Expect(obj).NotTo(gomega.BeNil(), "Parsed object is nil")
}

// parseTemplateFile parses the template from the file
// and saves its structured content to the object.
func (h *Link) parseTemplateFile(
	ctx context.Context,
	obj client.Object,
	templateFile TemplateFile,
	bindings Bindings,
) {
	templateContent := h.readTemplateFile(templateFile)
	h.parseTemplate(ctx, obj, templateContent, bindings)
}

// validateObject asserts that the object is not nil and that it has a name.
func (h *Link) validateObject(obj client.Object) {
	h.Gomega.Expect(obj).NotTo(gomega.BeNil(), "Object must not be nil")
	h.Gomega.Expect(obj.GetName()).NotTo(gomega.BeEmpty(), "Object must have a name")
}
