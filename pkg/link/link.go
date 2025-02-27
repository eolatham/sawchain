package link

import (
	"context"
	"os"

	g "github.com/onsi/gomega"
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
	Client  client.Client
	Options LinkOptions
}

func NewLink(c client.Client, opts ...LinkOption) Link {
	return Link{Client: c, Options: NewLinkOptions(opts)}
}

// validateOptions executes assertions common for all operations on input options.
func (h *Link) validateOptions(options interface{}) {
	g.Expect(options).NotTo(g.And(
		g.HaveField("TemplateContent", g.Not(g.BeEmpty())),
		g.HaveField("TemplateFile", g.Not(g.BeEmpty()))),
		"Invalid options: TemplateContent and TemplateFile are mutually exclusive")
}

// requireTemplate asserts that either TemplateContent or TemplateFile is provided (mutually exclusive).
func (h *Link) requireTemplate(options interface{}) {
	g.Expect(options).To(g.Or(
		g.And(g.HaveField("TemplateContent", g.Not(g.BeEmpty())), g.HaveField("TemplateFile", g.BeEmpty())),
		g.And(g.HaveField("TemplateContent", g.BeEmpty()), g.HaveField("TemplateFile", g.Not(g.BeEmpty())))),
		"Invalid options: expected either TemplateContent or TemplateFile to be provided (mutually exclusive)")
}

// parseTemplate parses the template and saves its structured content to the object.
func (h *Link) parseTemplate(
	ctx context.Context,
	obj client.Object,
	template TemplateContent,
	bindings Bindings,
) {
	g.Expect(obj).NotTo(g.BeNil(), "Given object must be an actual pointer in order to save state")
	var err error
	obj, err = chainsaw.ParseResource(h.Client, ctx, string(template), bindings)
	g.Expect(err).NotTo(g.HaveOccurred(), "Failed to parse template")
	g.Expect(obj).NotTo(g.BeNil(), "Parsed object is nil")
}

// parseTemplateFile parses the template from the file
// and saves its structured content to the object.
func (h *Link) parseTemplateFile(
	ctx context.Context,
	obj client.Object,
	templateFile TemplateFile,
	bindings Bindings,
) {
	content, err := os.ReadFile(string(templateFile))
	g.Expect(err).NotTo(g.HaveOccurred(), "Failed to read template file")
	g.Expect(content).NotTo(g.BeEmpty(), "Template file content is empty")
	h.parseTemplate(ctx, obj, TemplateContent(string(content)), bindings)
}

// validateObject asserts that the object is not nil and that it has a name.
func (h *Link) validateObject(obj client.Object) {
	g.Expect(obj).NotTo(g.BeNil(), "Object must not be nil")
	g.Expect(obj.GetName()).NotTo(g.BeEmpty(), "Object must have a name")
}
