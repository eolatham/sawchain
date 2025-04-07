package sawchain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/eolatham/sawchain/internal/chainsaw"
	"github.com/eolatham/sawchain/internal/matchers"
	"github.com/eolatham/sawchain/internal/options"
	"github.com/eolatham/sawchain/internal/util"
)

const (
	errInvalidArgs          = "invalid arguments"
	errInvalidTemplate      = "invalid template/bindings"
	errInvalidObjectsLength = "invalid objects slice: length must match template resource count"

	errCacheNotSynced = "client cache not synced within timeout"
	errFailedSave     = "failed to save state to object"

	errFailedCreateWithTemplate = "failed to create with template"
	errFailedCreateWithObject   = "failed to create with object"
	errFailedUpdateWithTemplate = "failed to update with template"
	errFailedUpdateWithObject   = "failed to update with object"
	errFailedDeleteWithTemplate = "failed to delete with template"
	errFailedDeleteWithObject   = "failed to delete with object"

	errNilOpts             = "internal error: parsed options is nil"
	errCreatedMatcherIsNil = "internal error: created matcher is nil"
)

// Sawchain provides utilities for K8s YAML-driven testing—backed by Chainsaw. It includes helpers
// to reliably create/update/delete test resources and Gomega-friendly APIs to simplify assertions.
// Use New to create a Sawchain instance.
type Sawchain struct {
	t    testing.TB
	g    gomega.Gomega
	c    client.Client
	opts options.Options
}

// New creates a new Sawchain instance with the provided global settings.
//
// Invalid input will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after t and c:
//
//   - Bindings (map[string]any): Optional. Global bindings to be used in all Chainsaw template
//     operations. If multiple maps are provided, they will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to 5s. Default timeout for eventual
//     assertions. If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to 1s. Default polling interval for
//     eventual assertions. If provided, must be after timeout.
//
// # Examples
//
// Create a Sawchain instance with the default settings:
//
//	sc := sawchain.New(t, k8sClient)
//
// Create a Sawchain instance with global bindings:
//
//	sc := sawchain.New(t, k8sClient, map[string]any{"namespace", "test"})
//
// Create a Sawchain instance with custom timeout and interval settings:
//
//	sc := sawchain.New(t, k8sClient, "10s", "2s")
func New(t testing.TB, c client.Client, args ...interface{}) *Sawchain {
	// Create Gomega
	g := gomega.NewWithT(t)
	// Check client
	g.Expect(c).NotTo(gomega.BeNil(), "client must not be nil")
	// Parse options
	opts, err := options.ParseAndRequireGlobal(&options.Options{
		Timeout:  time.Second * 5,
		Interval: time.Second,
	}, args...)
	g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// Instantiate Sawchain
	return &Sawchain{t: t, g: g, c: c, opts: *opts}
}

// HELPER FUNCTIONS

func (s *Sawchain) id(obj client.Object) string {
	return util.GetResourceID(obj, s.c.Scheme())
}

func (s *Sawchain) get(ctx context.Context, obj client.Object) error {
	return s.c.Get(ctx, client.ObjectKeyFromObject(obj), obj)
}

func (s *Sawchain) getF(ctx context.Context, obj client.Object) func() error {
	return func() error { return s.get(ctx, obj) }
}

func (s *Sawchain) checkResourceVersion(ctx context.Context, obj client.Object, minResourceVersion string) error {
	if err := s.get(ctx, obj); err != nil {
		return err
	}
	actualResourceVersion := obj.GetResourceVersion()
	if actualResourceVersion < minResourceVersion {
		return fmt.Errorf("%s: insufficient resource version: expected at least %s but got %s",
			s.id(obj), minResourceVersion, actualResourceVersion)
	}
	return nil
}

func (s *Sawchain) checkResourceVersionF(ctx context.Context, obj client.Object, minResourceVersion string) func() error {
	return func() error { return s.checkResourceVersion(ctx, obj, minResourceVersion) }
}

func (s *Sawchain) checkNotFound(ctx context.Context, obj client.Object) error {
	err := s.get(ctx, obj)
	if err == nil {
		return fmt.Errorf("%s: expected resource not to be found", s.id(obj))
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}

func (s *Sawchain) checkNotFoundF(ctx context.Context, obj client.Object) func() error {
	return func() error { return s.checkNotFound(ctx, obj) }
}

// CREATE OPERATIONS

// CreateResourceAndWait creates a resource with an object, manifest, or Chainsaw template, and ensures
// client Get operations for the resource succeed within a configurable duration before returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the resource immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Object (client.Object): Required if a template is not provided. Typed or unstructured object for
//     reading/writing resource state. If provided without a template, resource state will be read from
//     the object for creation. If provided with a template, resource state will be read from the
//     template and written to the object. State will be maintained in the original input format,
//     which may require internal type conversions using the client scheme.
//
//   - Template (string): Required if an object is not provided. File path or content of a static
//     manifest or Chainsaw template containing a single complete resource definition. If provided,
//     resource state will be read from the template for creation.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for the resource should succeed after creation.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resource after creation. If provided, must be after timeout.
//
// # Examples
//
// Create a resource with an object:
//
//	sc.CreateResourceAndWait(ctx, obj)
//
// Create a resource with a manifest file and override duration settings:
//
//	sc.CreateResourceAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Create a resource with a Chainsaw template and bindings:
//
//	sc.CreateResourceAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Create a resource with a Chainsaw template and save the resource's state to an object:
//
//	sc.CreateResourceAndWait(ctx, configMap, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
func (s *Sawchain) CreateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObj, err := chainsaw.RenderTemplateSingle(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Create resource
		s.g.Expect(s.c.Create(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedCreateWithTemplate)

		// Wait for cache to sync
		s.g.Eventually(s.getF(ctx, &unstructuredObj), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)

		// Save object
		if opts.Object != nil {
			s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Object)).To(gomega.Succeed(), errFailedSave)
		}
	} else {
		// Create resource
		s.g.Expect(s.c.Create(ctx, opts.Object)).To(gomega.Succeed(), errFailedCreateWithObject)

		// Wait for cache to sync
		s.g.Eventually(s.getF(ctx, opts.Object), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// CreateResourcesAndWait creates resources with objects, a manifest, or a Chainsaw template containing
// multiple resources, and ensures client Get operations for all resources succeed within a configurable
// duration before returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the resources immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Objects ([]client.Object): Required if a template is not provided. Slice of typed or unstructured
//     objects for reading/writing resource states. If provided without a template, resource states will be
//     read from the objects for creation. If provided with a template, resource states will be read from the
//     template and written to the objects. States will be maintained in the original input format, which may
//     require internal type conversions using the client scheme.
//
//   - Template (string): Required if objects are not provided. File path or content of a static
//     manifest or Chainsaw template containing multiple complete resource definitions. If provided,
//     resource states will be read from the template for creation.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for all resources should succeed after creation.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resources after creation. If provided, must be after timeout.
//
// # Examples
//
// Create resources with objects:
//
//	sc.CreateResourcesAndWait(ctx, []client.Object{obj1, obj2, obj3})
//
// Create resources with a manifest file and override duration settings:
//
//	sc.CreateResourcesAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Create resources with a Chainsaw template and bindings:
//
//	sc.CreateResourcesAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: (join('-', [$prefix, 'cm']))
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	  ---
//	  apiVersion: v1
//	  kind: Secret
//	  metadata:
//	    name: (join('-', [$prefix, 'secret']))
//	    namespace: ($namespace)
//	  type: Opaque
//	  stringData:
//	    username: admin
//	    password: secret
//	`, map[string]any{"prefix": "test", "namespace": "default"})
//
// Create resources with a Chainsaw template and save the resources' states to objects:
//
//	sc.CreateResourcesAndWait(ctx, []client.Object{configMap, secret}, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: (join('-', [$prefix, 'cm']))
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	  ---
//	  apiVersion: v1
//	  kind: Secret
//	  metadata:
//	    name: (join('-', [$prefix, 'secret']))
//	    namespace: ($namespace)
//	  type: Opaque
//	  stringData:
//	    username: admin
//	    password: secret
//	`, map[string]any{"prefix": "test", "namespace": "default"})
func (s *Sawchain) CreateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errInvalidObjectsLength)
		}

		// Create resources
		for _, unstructuredObj := range unstructuredObjs {
			s.g.Expect(s.c.Create(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedCreateWithTemplate)
		}

		// Wait for cache to sync
		getAll := func() error {
			for i := range unstructuredObjs {
				// Use index to update object in outer scope
				if err := s.get(ctx, &unstructuredObjs[i]); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(getAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)

		// Save objects
		if opts.Objects != nil {
			for i, unstructuredObj := range unstructuredObjs {
				s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Objects[i])).To(gomega.Succeed(), errFailedSave)
			}
		}
	} else {
		// Create resources
		for _, obj := range opts.Objects {
			s.g.Expect(s.c.Create(ctx, obj)).To(gomega.Succeed(), errFailedCreateWithObject)
		}

		// Wait for cache to sync
		getAll := func() error {
			for _, obj := range opts.Objects {
				if err := s.get(ctx, obj); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(getAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// UPDATE OPERATIONS

// UpdateResourceAndWait updates a resource with an object, manifest, or Chainsaw template, and ensures
// client Get operations for the resource reflect the update within a configurable duration before
// returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the updated resource immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Object (client.Object): Required if a template is not provided. Typed or unstructured object for
//     reading/writing resource state. If provided without a template, resource state will be read from
//     the object for update. If provided with a template, resource state will be read from the
//     template and written to the object. State will be maintained in the original input format,
//     which may require internal type conversions using the client scheme.
//
//   - Template (string): Required if an object is not provided. File path or content of a static
//     manifest or Chainsaw template containing a single complete resource definition. If provided,
//     resource state will be read from the template for update.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for the resource should reflect the update.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resource after update. If provided, must be after timeout.
//
// # Examples
//
// Update a resource with an object:
//
//	sc.UpdateResourceAndWait(ctx, obj)
//
// Update a resource with a manifest file and override duration settings:
//
//	sc.UpdateResourceAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Update a resource with a Chainsaw template and bindings:
//
//	sc.UpdateResourceAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: updated-value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Update a resource with a Chainsaw template and save the resource's updated state to an object:
//
//	sc.UpdateResourceAndWait(ctx, configMap, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: updated-value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
func (s *Sawchain) UpdateResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObj, err := chainsaw.RenderTemplateSingle(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Update resource
		s.g.Expect(s.c.Update(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedUpdateWithTemplate)

		// Wait for cache to sync
		updatedResourceVersion := unstructuredObj.GetResourceVersion()
		s.g.Eventually(s.checkResourceVersionF(ctx, &unstructuredObj, updatedResourceVersion),
			opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)

		// Save object
		if opts.Object != nil {
			s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Object)).To(gomega.Succeed(), errFailedSave)
		}
	} else {
		// Update resource
		s.g.Expect(s.c.Update(ctx, opts.Object)).To(gomega.Succeed(), errFailedUpdateWithObject)

		// Wait for cache to sync
		updatedResourceVersion := opts.Object.GetResourceVersion()
		s.g.Eventually(s.checkResourceVersionF(ctx, opts.Object, updatedResourceVersion),
			opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// UpdateResourcesAndWait updates resources with objects, a manifest, or a Chainsaw template containing
// multiple resources, and ensures client Get operations for all resources reflect the updates within a
// configurable duration before returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the updated resources immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Objects ([]client.Object): Required if a template is not provided. Slice of typed or unstructured
//     objects for reading/writing resource states. If provided without a template, resource states will be
//     read from the objects for update. If provided with a template, resource states will be read from the
//     template and written to the objects. States will be maintained in the original input format, which may
//     require internal type conversions using the client scheme.
//
//   - Template (string): Required if objects are not provided. File path or content of a static
//     manifest or Chainsaw template containing multiple complete resource definitions. If provided,
//     resource states will be read from the template for update.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for all resources should reflect the updates.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resources after update. If provided, must be after timeout.
//
// # Examples
//
// Update resources with objects:
//
//	sc.UpdateResourcesAndWait(ctx, []client.Object{obj1, obj2, obj3})
//
// Update resources with a manifest file and override duration settings:
//
//	sc.UpdateResourcesAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Update resources with a Chainsaw template and bindings:
//
//	sc.UpdateResourcesAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: (join('-', [$prefix, 'cm']))
//	    namespace: ($namespace)
//	  data:
//	    key: updated-value
//	  ---
//	  apiVersion: v1
//	  kind: Secret
//	  metadata:
//	    name: (join('-', [$prefix, 'secret']))
//	    namespace: ($namespace)
//	  type: Opaque
//	  stringData:
//	    username: admin
//	    password: updated-secret
//	`, map[string]any{"prefix": "test", "namespace": "default"})
//
// Update resources with a Chainsaw template and save the resources' updated states to objects:
//
//	sc.UpdateResourcesAndWait(ctx, []client.Object{configMap, secret}, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: (join('-', [$prefix, 'cm']))
//	    namespace: ($namespace)
//	  data:
//	    key: updated-value
//	  ---
//	  apiVersion: v1
//	  kind: Secret
//	  metadata:
//	    name: (join('-', [$prefix, 'secret']))
//	    namespace: ($namespace)
//	  type: Opaque
//	  stringData:
//	    username: admin
//	    password: updated-secret
//	`, map[string]any{"prefix": "test", "namespace": "default"})
func (s *Sawchain) UpdateResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errInvalidObjectsLength)
		}

		// Update resources
		for _, unstructuredObj := range unstructuredObjs {
			s.g.Expect(s.c.Update(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedUpdateWithTemplate)
		}

		// Wait for cache to sync
		updatedResourceVersions := make([]string, len(unstructuredObjs))
		for i := range unstructuredObjs {
			updatedResourceVersions[i] = unstructuredObjs[i].GetResourceVersion()
		}
		checkAll := func() error {
			for i := range unstructuredObjs {
				// Use index to update object in outer scope
				if err := s.checkResourceVersion(ctx, &unstructuredObjs[i], updatedResourceVersions[i]); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(checkAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)

		// Save objects
		if opts.Objects != nil {
			for i, unstructuredObj := range unstructuredObjs {
				s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Objects[i])).To(gomega.Succeed(), errFailedSave)
			}
		}
	} else {
		// Update resources
		for _, obj := range opts.Objects {
			s.g.Expect(s.c.Update(ctx, obj)).To(gomega.Succeed(), errFailedUpdateWithObject)
		}

		// Wait for cache to sync
		updatedResourceVersions := make([]string, len(opts.Objects))
		for i := range opts.Objects {
			updatedResourceVersions[i] = opts.Objects[i].GetResourceVersion()
		}
		checkAll := func() error {
			for i := range opts.Objects {
				if err := s.checkResourceVersion(ctx, opts.Objects[i], updatedResourceVersions[i]); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(checkAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// DELETE OPERATIONS

// DeleteResourceAndWait deletes a resource with an object, manifest, or Chainsaw template, and ensures
// client Get operations for the resource reflect the deletion (resource not found) within a configurable
// duration before returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the resource's absence immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Object (client.Object): Required if a template is not provided. Typed or unstructured object
//     representing the resource to be deleted. If provided with a template, the template will take
//     precedence and the object will be ignored.
//
//   - Template (string): Required if an object is not provided. File path or content of a static manifest
//     or Chainsaw template containing the identifying metadata of the resource to be deleted. Takes
//     precedence over object.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for the resource should reflect deletion.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resource after deletion. If provided, must be after timeout.
//
// # Examples
//
// Delete a resource with an object:
//
//	sc.DeleteResourceAndWait(ctx, obj)
//
// Delete a resource with a manifest file and override duration settings:
//
//	sc.DeleteResourceAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Delete a resource with a Chainsaw template and bindings:
//
//	sc.DeleteResourceAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
func (s *Sawchain) DeleteResourceAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObj, err := chainsaw.RenderTemplateSingle(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Delete resource
		s.g.Expect(s.c.Delete(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedDeleteWithTemplate)

		// Wait for cache to sync
		s.g.Eventually(s.checkNotFoundF(ctx, &unstructuredObj), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	} else {
		// Delete resource
		s.g.Expect(s.c.Delete(ctx, opts.Object)).To(gomega.Succeed(), errFailedDeleteWithObject)

		// Wait for cache to sync
		s.g.Eventually(s.checkNotFoundF(ctx, opts.Object), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// DeleteResourcesAndWait deletes resources with objects, a manifest, or a Chainsaw template containing
// multiple resources, and ensures client Get operations for all resources reflect the deletion (resources
// not found) within a configurable duration before returning.
//
// If testing with a cached client, this ensures the client cache is synced and it is safe to make
// assertions on the resources' absence immediately after execution.
//
// Invalid input, client errors, and timeout errors will result in immediate test failure.
//
// # Arguments
//
// The following arguments may be provided in any order (unless noted otherwise) after the context:
//
//   - Objects ([]client.Object): Required if a template is not provided. Slice of typed or unstructured
//     objects representing the resources to be deleted. If provided with a template, the template will take
//     precedence and the objects will be ignored.
//
//   - Template (string): Required if objects are not provided. File path or content of a static manifest
//     or Chainsaw template containing the identifying metadata of the resources to be deleted. Takes
//     precedence over objects.
//
//   - Bindings (map[string]any): Optional. Bindings to be applied to a Chainsaw template (if provided)
//     in addition to (or overriding) Sawchain's global bindings. If multiple maps are provided, they
//     will be merged in natural order.
//
//   - Timeout (string or time.Duration): Optional. Defaults to Sawchain's global timeout value.
//     Duration within which client Get operations for all resources should reflect deletion.
//     If provided, must be before interval.
//
//   - Interval (string or time.Duration): Optional. Defaults to Sawchain's global interval value.
//     Polling interval for checking the resources after deletion. If provided, must be after timeout.
//
// # Examples
//
// Delete resources with objects:
//
//	sc.DeleteResourcesAndWait(ctx, []client.Object{obj1, obj2, obj3})
//
// Delete resources with a manifest file and override duration settings:
//
//	sc.DeleteResourcesAndWait(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Delete resources with a Chainsaw template and bindings:
//
//	sc.DeleteResourcesAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: (join('-', [$prefix, 'cm']))
//	    namespace: ($namespace)
//	  ---
//	  apiVersion: v1
//	  kind: Secret
//	  metadata:
//	    name: (join('-', [$prefix, 'secret']))
//	    namespace: ($namespace)
//	`, map[string]any{"prefix": "test", "namespace": "default"})
func (s *Sawchain) DeleteResourcesAndWait(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventualMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errInvalidObjectsLength)
		}

		// Delete resources
		for _, unstructuredObj := range unstructuredObjs {
			s.g.Expect(s.c.Delete(ctx, &unstructuredObj)).To(gomega.Succeed(), errFailedDeleteWithTemplate)
		}

		// Wait for cache to sync
		checkAll := func() error {
			for i := range unstructuredObjs {
				// Use index to update object in outer scope
				if err := s.checkNotFound(ctx, &unstructuredObjs[i]); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(checkAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	} else {
		// Delete resources
		for _, obj := range opts.Objects {
			s.g.Expect(s.c.Delete(ctx, obj)).To(gomega.Succeed(), errFailedDeleteWithObject)
		}

		// Wait for cache to sync
		checkAll := func() error {
			for _, obj := range opts.Objects {
				if err := s.checkNotFound(ctx, obj); err != nil {
					return err
				}
			}
			return nil
		}
		s.g.Eventually(checkAll, opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
	}

	return nil
}

// GET OPERATIONS

// TODO: document
// GetResource gets a resource from the cluster.
func (s *Sawchain) GetResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// GetResources gets multiple resources from the cluster.
func (s *Sawchain) GetResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// GetResourceFunc returns a function that gets a resource for use with Eventually.
func (s *Sawchain) GetResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// GetResourcesFunc returns a function that gets multiple resources for use with Eventually.
func (s *Sawchain) GetResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FETCH OPERATIONS

// TODO: document
// FetchResource fetches a resource from the cluster.
func (s *Sawchain) FetchResource(ctx context.Context, args ...interface{}) client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// FetchResources fetches multiple resources from the cluster.
func (s *Sawchain) FetchResources(ctx context.Context, args ...interface{}) []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// FetchResourceFunc returns a function that fetches a resource for use with Eventually.
func (s *Sawchain) FetchResourceFunc(ctx context.Context, args ...interface{}) func() client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// FetchResourcesFunc returns a function that fetches multiple resources for use with Eventually.
func (s *Sawchain) FetchResourcesFunc(ctx context.Context, args ...interface{}) func() []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CHECK OPERATIONS

// TODO: document
// CheckResource checks if a resource matches the expected state.
func (s *Sawchain) CheckResource(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// CheckResources checks if multiple resources match the expected state.
func (s *Sawchain) CheckResources(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// CheckResourceFunc returns a function that checks a resource for use with Eventually.
func (s *Sawchain) CheckResourceFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: document
// CheckResourcesFunc returns a function that checks multiple resources for use with Eventually.
func (s *Sawchain) CheckResourcesFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CUSTOM MATCHERS

// TODO: document
// TODO: recommend enabling format.UseStringerRepresentation for better failure output
// Match returns a matcher that checks if a client.Object matches a Chainsaw template.
func (s *Sawchain) MatchYAML(template string, bindings ...map[string]any) types.GomegaMatcher {
	if util.IsExistingFile(template) {
		var err error
		template, err = util.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), "internal error: failed to read template file")
	}
	matcher := matchers.NewChainsawMatcher(s.c, template, util.MergeMaps(bindings...))
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}

// TODO: document
// TODO: recommend enabling format.UseStringerRepresentation for better failure output
// HaveStatusCondition returns a matcher that checks if a client.Object has a specific status condition.
func (s *Sawchain) HaveStatusCondition(conditionType, expectedStatus string) types.GomegaMatcher {
	matcher := matchers.NewStatusConditionMatcher(s.c, conditionType, expectedStatus)
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}
