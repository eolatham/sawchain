package sawchain

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/matchers"
	"github.com/eolatham/sawchain/internal/options"
	"github.com/eolatham/sawchain/internal/utilities"
)

// TODO: test

const (
	errInvalidArgs         = "invalid arguments"
	errNilOpts             = "internal error: parsed options is nil"
	errCreateMatcherFailed = "failed to create matcher"
	errCreatedMatcherIsNil = "internal error: created matcher is nil"
)

// Sawchain provides a Chainsaw-backed testing utility for K8s.
type Sawchain struct {
	t    testing.TB
	g    gomega.Gomega
	c    client.Client
	opts options.Options
}

// New creates a new Sawchain instance.
func New(t testing.TB, c client.Client, args ...interface{}) *Sawchain {
	// Create Gomega
	g := gomega.NewGomegaWithT(t)
	// Parse options
	opts, err := options.ParseAndRequireGlobal(&options.Options{
		Timeout:  time.Second * 10,
		Interval: time.Second,
	}, args...)
	g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// Instantiate Sawchain
	return &Sawchain{t: t, g: g, c: c, opts: *opts}
}

// CREATION OPERATIONS

// CreateResourceAndWait creates a resource from an object, manifest, or Chainsaw template,
// and waits for GET operations to succeed before returning.
//
// This is especially helpful when testing with a cached client, as it ensures the client cache
// is synced and it is safe to make assertions on the resource immediately after execution.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Object (client.Object): Required if a template is not provided.
//     Typed or unstructured object for reading/writing resource state. State will be maintained in the original
//     input format, and the client scheme may be used to perform internal type conversions if needed.
//     If provided without a template, resource state will be read from the object for creation.
//     If provided with a template, resource state will be read from the template and written to the object.
//
//   - Template (string): Required if an object is not provided.
//     May be the file path or the content of a Chainsaw template (or static manifest)
//     containing a single complete resource definition.
//     If provided without an object, resource state will be read from the template for creation.
//     If provided with an object, resource state will be read from the template and written to the object.
//
//   - Bindings (map[string]any): Optional. Defaults to Sawchain's global bindings.
//     Bindings to be applied to a Chainsaw template (if provided).
//     If multiple maps are provided, they will all be used.
//     Sawchain's global bindings are always included.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     The duration within which getting the resource should succeed after creation
//     (i.e. how long to wait for the client cache to sync).
//     Must be provided before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     The polling interval for checking the resource after creation.
//     Must be provided after timeout.
//
// # Examples
//
// Create a resource with an object:
//
//	sc.CreateResourceAndWait(ctx, obj)
//
// Create a resource with a static manifest file and override duration settings:
//
//	sc.CreateResourceAndWait(ctx, "path/to/manifest.yaml", "30s", "2s")
//
// Create a resource with a Chainsaw template and bindings:
//
//	sc.CreateResourceAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: {{ .name }}
//	    namespace: {{ .namespace }}
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Create a resource with a Chainsaw template and bindings and save the resource's state to an object:
//
//	sc.CreateResourceAndWait(ctx, obj, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: {{ .name }}
//	    namespace: {{ .namespace }}
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
func (s *Sawchain) CreateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CreateResourcesAndWait creates multiple resources and waits for client cache to sync.
func (s *Sawchain) CreateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// UPDATE OPERATIONS

// UpdateResourceAndWait updates a resource and waits for client cache to sync.
func (s *Sawchain) UpdateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// UpdateResourcesAndWait updates multiple resources and waits for client cache to sync.
func (s *Sawchain) UpdateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// DELETE OPERATIONS

// DeleteResourceAndWait deletes a resource and waits for client cache to sync.
func (s *Sawchain) DeleteResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// DeleteResourcesAndWait deletes multiple resources and waits for client cache to sync.
func (s *Sawchain) DeleteResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GET OPERATIONS

// GetResource gets a resource from the cluster.
func (s *Sawchain) GetResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResources gets multiple resources from the cluster.
func (s *Sawchain) GetResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResourceFunc returns a function that gets a resource for use with Eventually.
func (s *Sawchain) GetResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResourcesFunc returns a function that gets multiple resources for use with Eventually.
func (s *Sawchain) GetResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FETCH OPERATIONS

// FetchResource fetches a resource from the cluster.
func (s *Sawchain) FetchResource(ctx context.Context, args ...interface{}) client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResources fetches multiple resources from the cluster.
func (s *Sawchain) FetchResources(ctx context.Context, args ...interface{}) []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResourceFunc returns a function that fetches a resource for use with Eventually.
func (s *Sawchain) FetchResourceFunc(ctx context.Context, args ...interface{}) func() client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResourcesFunc returns a function that fetches multiple resources for use with Eventually.
func (s *Sawchain) FetchResourcesFunc(ctx context.Context, args ...interface{}) func() []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CHECK OPERATIONS

// CheckResource checks if a resource matches the expected state.
func (s *Sawchain) CheckResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResources checks if multiple resources match the expected state.
func (s *Sawchain) CheckResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResourceFunc returns a function that checks a resource for use with Eventually.
func (s *Sawchain) CheckResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResourcesFunc returns a function that checks multiple resources for use with Eventually.
func (s *Sawchain) CheckResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CUSTOM MATCHERS

// Match returns a matcher that checks if a client.Object matches a Chainsaw resource template.
func (s *Sawchain) MatchYAML(template string, bindings ...map[string]any) types.GomegaMatcher {
	if utilities.IsExistingFile(template) {
		var err error
		template, err = utilities.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), "internal error: failed to read template file")
	}
	matcher, err := matchers.NewChainsawMatcher(s.c, template, utilities.MergeMaps(bindings...))
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errCreateMatcherFailed)
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}

// HaveStatusCondition returns a matcher that checks if a client.Object has a specific status condition.
func (s *Sawchain) HaveStatusCondition(conditionType, expectedStatus string) types.GomegaMatcher {
	matcher, err := matchers.NewStatusConditionMatcher(s.c, conditionType, expectedStatus)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errCreateMatcherFailed)
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}
