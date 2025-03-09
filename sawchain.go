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

const (
	errInvalidArgs = "invalid arguments"
	errNilOpts     = "internal error: parsed options is nil"
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
	opts, err := options.ParseGlobal(&options.Options{
		Timeout:  time.Second * 10,
		Interval: time.Second,
	}, args...)
	g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// Instantiate Sawchain
	return &Sawchain{t: t, g: g, c: c, opts: *opts}
}

// CREATION OPERATIONS

// CreateResourceAndWait creates a resource and waits for client cache to sync.
func (s *Sawchain) CreateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CreateResourcesAndWait creates multiple resources and waits for client cache to sync.
func (s *Sawchain) CreateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// UPDATE OPERATIONS

// UpdateResourceAndWait updates a resource and waits for client cache to sync.
func (s *Sawchain) UpdateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// UpdateResourcesAndWait updates multiple resources and waits for client cache to sync.
func (s *Sawchain) UpdateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// DELETE OPERATIONS

// DeleteResourceAndWait deletes a resource and waits for client cache to sync.
func (s *Sawchain) DeleteResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// DeleteResourcesAndWait deletes multiple resources and waits for client cache to sync.
func (s *Sawchain) DeleteResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseEventualMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GET OPERATIONS

// GetResource gets a resource from the cluster.
func (s *Sawchain) GetResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResources gets multiple resources from the cluster.
func (s *Sawchain) GetResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResourceFunc returns a function that gets a resource for use with Eventually.
func (s *Sawchain) GetResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// GetResourcesFunc returns a function that gets multiple resources for use with Eventually.
func (s *Sawchain) GetResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FETCH OPERATIONS

// FetchResource fetches a resource from the cluster.
func (s *Sawchain) FetchResource(ctx context.Context, args ...interface{}) client.Object {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResources fetches multiple resources from the cluster.
func (s *Sawchain) FetchResources(ctx context.Context, args ...interface{}) []client.Object {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResourceFunc returns a function that fetches a resource for use with Eventually.
func (s *Sawchain) FetchResourceFunc(ctx context.Context, args ...interface{}) func() client.Object {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FetchResourcesFunc returns a function that fetches multiple resources for use with Eventually.
func (s *Sawchain) FetchResourcesFunc(ctx context.Context, args ...interface{}) func() []client.Object {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CHECK OPERATIONS

// CheckResource checks if a resource matches the expected state.
func (s *Sawchain) CheckResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResources checks if multiple resources match the expected state.
func (s *Sawchain) CheckResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResourceFunc returns a function that checks a resource for use with Eventually.
func (s *Sawchain) CheckResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CheckResourcesFunc returns a function that checks multiple resources for use with Eventually.
func (s *Sawchain) CheckResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseImmediateMultiple(&s.opts, args...)
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
	matcher, err := matchers.NewMatchYAMLMatcher(s.c, template, utilities.MergeMaps(bindings...))
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create MatchYAMLMatcher")
	s.g.Expect(matcher).NotTo(gomega.BeNil(), "internal error: created MatchYAMLMatcher is nil")
	return matcher
}

// HaveStatusCondition returns a matcher that checks if a client.Object has a specific status condition.
func (s *Sawchain) HaveStatusCondition(conditionType, expectedStatus string) types.GomegaMatcher {
	return matchers.NewStatusConditionMatcher(conditionType, expectedStatus)
}
