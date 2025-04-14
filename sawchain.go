package sawchain

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/eolatham/sawchain/internal/chainsaw"
	"github.com/eolatham/sawchain/internal/matchers"
	"github.com/eolatham/sawchain/internal/options"
	"github.com/eolatham/sawchain/internal/util"
)

const (
	errInvalidArgs        = "invalid arguments"
	errInvalidTemplate    = "invalid template/bindings"
	errObjectInsufficient = "single object insufficient for multi-resource template"
	errObjectsWrongLength = "objects slice length must match template resource count"

	errCacheNotSynced = "client cache not synced within timeout"
	errFailedSave     = "failed to save state to object"
	errFailedWrite    = "failed to write file"

	errFailedCreateWithTemplate = "failed to create with template"
	errFailedCreateWithObject   = "failed to create with object"
	errFailedUpdateWithTemplate = "failed to update with template"
	errFailedUpdateWithObject   = "failed to update with object"
	errFailedDeleteWithTemplate = "failed to delete with template"
	errFailedDeleteWithObject   = "failed to delete with object"

	errNilOpts             = "internal error: parsed options is nil"
	errFailedReadTemplate  = "internal error: failed to read template file"
	errFailedMarshalObject = "internal error: failed to marshal object"
	errCreatedMatcherIsNil = "internal error: created matcher is nil"
)

// Sawchain provides utilities for K8s YAML-driven testing—backed by Chainsaw. It includes helpers to
// reliably create/update/delete test resources, Gomega-friendly APIs to simplify assertions, and more.
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

// CREATE/UPDATE/DELETE

// Create creates resources with objects, a manifest, or a Chainsaw template, and ensures client Get
// operations for all resources succeed within a configurable duration before returning.
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
//   - Object (client.Object): Typed or unstructured object for reading/writing the state of a single
//     resource. If provided without a template, resource state will be read from the object for creation.
//     If provided with a template, resource state will be read from the template and written to the object.
//     State will be maintained in the original input format, which may require internal type conversions
//     using the client scheme.
//
//   - Objects ([]client.Object): Slice of typed or unstructured objects for reading/writing the states of
//     multiple resources. If provided without a template, resource states will be read from the objects for
//     creation. If provided with a template, resource states will be read from the template and written to
//     the objects. States will be maintained in the original input format, which may require internal type
//     conversions using the client scheme.
//
//   - Template (string): File path or content of a static manifest or Chainsaw template containing complete
//     resource definitions to be read for creation. If provided with an object, must contain exactly one
//     resource definition matching the type of the object. If provided with a slice of objects, must
//     contain resource definitions exactly matching the count, order, and types of the objects.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
//   - Timeout (string or time.Duration): Duration within which client Get operations for all resources
//     should succeed after creation. If provided, must be before interval. Defaults to Sawchain's
//     global timeout value.
//
//   - Interval (string or time.Duration): Polling interval for checking the resources after creation.
//     If provided, must be after timeout. Defaults to Sawchain's global interval value.
//
// A template, an object, or a slice of objects must be provided. However, an object and a slice of objects
// may not be provided together. All other arguments are optional.
//
// # Examples
//
// Create a single resource with an object:
//
//	sc.Create(ctx, obj)
//
// Create multiple resources with objects:
//
//	sc.Create(ctx, []client.Object{obj1, obj2, obj3})
//
// Create resources with a manifest file and override duration settings:
//
//	sc.Create(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Create a single resource with a Chainsaw template and bindings:
//
//	sc.Create(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Create a single resource with a Chainsaw template and save the resource's state to an object:
//
//	sc.Create(ctx, configMap, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Create multiple resources with a Chainsaw template and bindings:
//
//	sc.Create(ctx, `
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
// Create multiple resources with a Chainsaw template and save the resources' states to objects:
//
//	sc.Create(ctx, []client.Object{configMap, secret}, `
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
func (s *Sawchain) Create(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventual(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Object != nil {
			s.g.Expect(unstructuredObjs).To(gomega.HaveLen(1), errObjectInsufficient)
		} else if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errObjectsWrongLength)
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
		if opts.Object != nil {
			s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObjs[0], opts.Object)).To(gomega.Succeed(), errFailedSave)
		} else if opts.Objects != nil {
			for i, unstructuredObj := range unstructuredObjs {
				s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Objects[i])).To(gomega.Succeed(), errFailedSave)
			}
		}
	} else if opts.Object != nil {
		// Create resource
		s.g.Expect(s.c.Create(ctx, opts.Object)).To(gomega.Succeed(), errFailedCreateWithObject)

		// Wait for cache to sync
		s.g.Eventually(s.getF(ctx, opts.Object), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
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

// TODO: test
// Update updates resources with objects, a manifest, or a Chainsaw template, and ensures client Get
// operations for all resources reflect the updates within a configurable duration before returning.
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
//   - Object (client.Object): Typed or unstructured object for reading/writing the state of a single
//     resource. If provided without a template, resource state will be read from the object for update.
//     If provided with a template, resource state will be read from the template and written to the object.
//     State will be maintained in the original input format, which may require internal type conversions
//     using the client scheme.
//
//   - Objects ([]client.Object): Slice of typed or unstructured objects for reading/writing the states of
//     multiple resources. If provided without a template, resource states will be read from the objects for
//     update. If provided with a template, resource states will be read from the template and written to
//     the objects. States will be maintained in the original input format, which may require internal type
//     conversions using the client scheme.
//
//   - Template (string): File path or content of a static manifest or Chainsaw template containing complete
//     resource definitions to be read for update. If provided with an object, must contain exactly one
//     resource definition matching the type of the object. If provided with a slice of objects, must
//     contain resource definitions exactly matching the count, order, and types of the objects.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
//   - Timeout (string or time.Duration): Duration within which client Get operations for all resources
//     should reflect the updates. If provided, must be before interval. Defaults to Sawchain's
//     global timeout value.
//
//   - Interval (string or time.Duration): Polling interval for checking the resources after updating.
//     If provided, must be after timeout. Defaults to Sawchain's global interval value.
//
// A template, an object, or a slice of objects must be provided. However, an object and a slice of objects
// may not be provided together. All other arguments are optional.
//
// # Examples
//
// Update a single resource with an object:
//
//	sc.UpdateResourceAndWait(ctx, obj)
//
// Update multiple resources with objects:
//
//	sc.Update(ctx, []client.Object{obj1, obj2, obj3})
//
// Update resources with a manifest file and override duration settings:
//
//	sc.Update(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Update a single resource with a Chainsaw template and bindings:
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
// Update a single resource with a Chainsaw template and save the resource's updated state to an object:
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
//
// Update multiple resources with a Chainsaw template and bindings:
//
//	sc.Update(ctx, `
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
// Update multiple resources with a Chainsaw template and save the resources' updated states to objects:
//
//	sc.Update(ctx, []client.Object{configMap, secret}, `
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
func (s *Sawchain) Update(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventual(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Object != nil {
			s.g.Expect(unstructuredObjs).To(gomega.HaveLen(1), errObjectInsufficient)
		} else if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errObjectsWrongLength)
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
		if opts.Object != nil {
			s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObjs[0], opts.Object)).To(gomega.Succeed(), errFailedSave)
		} else if opts.Objects != nil {
			for i, unstructuredObj := range unstructuredObjs {
				s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, opts.Objects[i])).To(gomega.Succeed(), errFailedSave)
			}
		}
	} else if opts.Object != nil {
		// Update resource
		s.g.Expect(s.c.Update(ctx, opts.Object)).To(gomega.Succeed(), errFailedUpdateWithObject)

		// Wait for cache to sync
		updatedResourceVersion := opts.Object.GetResourceVersion()
		s.g.Eventually(s.checkResourceVersionF(ctx, opts.Object, updatedResourceVersion),
			opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
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

// TODO: test
// Delete deletes resources with objects, a manifest, or a Chainsaw template, and ensures client Get
// operations for all resources reflect the deletion (resources not found) within a configurable
// duration before returning.
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
//   - Object (client.Object): Typed or unstructured object representing a single resource to be deleted.
//     If provided with a template, the template will take precedence and the object will be ignored.
//
//   - Objects ([]client.Object): Slice of typed or unstructured objects representing multiple resources to
//     be deleted. If provided with a template, the template will take precedence and the objects will be
//     ignored.
//
//   - Template (string): File path or content of a static manifest or Chainsaw template containing the
//     identifying metadata of the resources to be deleted. Takes precedence over objects.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
//   - Timeout (string or time.Duration): Duration within which client Get operations for all resources
//     should reflect deletion. If provided, must be before interval. Defaults to Sawchain's global
//     timeout value.
//
//   - Interval (string or time.Duration): Polling interval for checking the resources after deletion.
//     If provided, must be after timeout. Defaults to Sawchain's global interval value.
//
// A template, an object, or a slice of objects must be provided. However, an object and a slice of objects
// may not be provided together. All other arguments are optional.
//
// # Examples
//
// Delete a single resource with an object:
//
//	sc.DeleteResourceAndWait(ctx, obj)
//
// Delete multiple resources with objects:
//
//	sc.Delete(ctx, []client.Object{obj1, obj2, obj3})
//
// Delete resources with a manifest file and override duration settings:
//
//	sc.Delete(ctx, "path/to/manifest.yaml", "10s", "2s")
//
// Delete a single resource with a Chainsaw template and bindings:
//
//	sc.DeleteResourceAndWait(ctx, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Delete multiple resources with a Chainsaw template and bindings:
//
//	sc.Delete(ctx, `
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
func (s *Sawchain) Delete(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireEventual(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)

	if opts.Template != "" {
		// Render template
		unstructuredObjs, err := chainsaw.RenderTemplate(ctx, opts.Template, chainsaw.BindingsFromMap(opts.Bindings))
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)

		// Validate objects length
		if opts.Objects != nil {
			s.g.Expect(opts.Objects).To(gomega.HaveLen(len(unstructuredObjs)), errObjectsWrongLength)
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
	} else if opts.Object != nil {
		// Delete resource
		s.g.Expect(s.c.Delete(ctx, opts.Object)).To(gomega.Succeed(), errFailedDeleteWithObject)

		// Wait for cache to sync
		s.g.Eventually(s.checkNotFoundF(ctx, opts.Object), opts.Timeout, opts.Interval).Should(gomega.Succeed(), errCacheNotSynced)
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

// GET

// TODO: test
// TODO: document
func (s *Sawchain) Get(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediate(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: test
// TODO: document
func (s *Sawchain) GetFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediate(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// FETCH

// TODO: test
// TODO: document
func (s *Sawchain) FetchSingle(ctx context.Context, args ...interface{}) client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: test
// TODO: document
func (s *Sawchain) FetchMultiple(ctx context.Context, args ...interface{}) []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: test
// TODO: document
func (s *Sawchain) FetchSingleFunc(ctx context.Context, args ...interface{}) func() client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: test
// TODO: document
func (s *Sawchain) FetchMultipleFunc(ctx context.Context, args ...interface{}) func() []client.Object {
	// Parse options
	opts, err := options.ParseAndRequireImmediateMulti(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// CHECK

// TODO: test
// TODO: document
func (s *Sawchain) Check(ctx context.Context, args ...interface{}) error {
	// Parse options
	opts, err := options.ParseAndRequireImmediate(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// TODO: test
// TODO: document
func (s *Sawchain) CheckFunc(ctx context.Context, args ...interface{}) func() error {
	// Parse options
	opts, err := options.ParseAndRequireImmediateSingle(&s.opts, args...)
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidArgs)
	s.g.Expect(opts).NotTo(gomega.BeNil(), errNilOpts)
	// TODO: implement
	return nil
}

// MATCH

// TODO: test
// MatchYAML returns a Gomega matcher that tests if a client.Object matches a static manifest or Chainsaw
// template, including full support for Chainsaw JMESPath assertions.
//
// The returned matcher may rely on the client scheme for internal type conversions.
//
// Invalid input will result in immediate test failure.
//
// For better failure output, it's recommended to enable Gomega's format.UseStringerRepresentation.
//
// # Arguments
//
//   - Template (string): File path or content of a static manifest or Chainsaw template to match against.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
// # Examples
//
// Match an object against a static manifest file:
//
//	g.Expect(configMap).To(sc.MatchYAML("path/to/manifest.yaml"))
//
// Match an object against a template using bindings:
//
//	g.Expect(configMap).To(sc.MatchYAML(`
//	  apiVersion: v1
//	  kind: ConfigMap
//	  data:
//	    key1: ($value1)
//	    key2: ($value2)
//	`, map[string]any{"value1": "foo", "value2": "bar"}))
//
// Match a Deployment's replica count using a JMESPath assertion:
//
//	g.Expect(deployment).To(sc.MatchYAML(`
//	  apiVersion: apps/v1
//	  kind: Deployment
//	  spec:
//	    (replicas > `1` && replicas < `4`): true
//	`))
//
// For more assertion examples, go to https://kyverno.github.io/chainsaw/.
func (s *Sawchain) MatchYAML(template string, bindings ...map[string]any) types.GomegaMatcher {
	if util.IsExistingFile(template) {
		var err error
		template, err = util.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errFailedReadTemplate)
	}
	matcher := matchers.NewChainsawMatcher(s.c, template, util.MergeMaps(bindings...))
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}

// TODO: test
// HaveStatusCondition returns a Gomega matcher that uses an internal Chainsaw assertion to test if a
// client.Object has a specific status condition.
//
// The returned matcher may rely on the client scheme for internal type conversions.
//
// Invalid input will result in immediate test failure.
//
// For better failure output, it's recommended to enable Gomega's format.UseStringerRepresentation.
//
// # Arguments
//
//   - ConditionType (string): The type of the status condition to check for.
//
//   - ExpectedStatus (string): The expected status value of the condition.
//
// # Examples
//
// Check if a Deployment has condition Available=True:
//
//	g.Expect(deployment).To(sc.HaveStatusCondition("Available", "True"))
//
// Check if a Pod has condition Initialized=False:
//
//	g.Expect(pod).To(sc.HaveStatusCondition("Initialized", "False"))
//
// Check if a custom resource has condition Ready=True:
//
//	g.Expect(myCustomResource).To(sc.HaveStatusCondition("Ready", "True"))
func (s *Sawchain) HaveStatusCondition(conditionType, expectedStatus string) types.GomegaMatcher {
	matcher := matchers.NewStatusConditionMatcher(s.c, conditionType, expectedStatus)
	s.g.Expect(matcher).NotTo(gomega.BeNil(), errCreatedMatcherIsNil)
	return matcher
}

// RENDER

// TODO: test
// RenderToObject renders a Chainsaw template with optional bindings into an object.
//
// This can also be used as a convenience method for unmarshaling static manifests.
//
// Invalid input will result in immediate test failure.
//
// # Arguments
//
//   - Object (client.Object): Typed or unstructured object to render into. If the object is typed, the
//     client scheme will be used for conversion.
//
//   - Template (string): File path or content of a static manifest or Chainsaw template to render. Must
//     contain exactly one complete resource definition matching the type of the provided object.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
// # Examples
//
// Render a resource from a template using bindings:
//
//	sc.RenderToObject(configMap, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: ($name)
//	    namespace: ($namespace)
//	  data:
//	    key: value
//	`, map[string]any{"name": "test-cm", "namespace": "default"})
//
// Render a resource from a template file using bindings:
//
//	sc.RenderToObject(secret, "path/to/template.yaml",
//	  map[string]any{"name": "test-secret", "namespace": "default"})
//
// Unmarshal a static manifest into an object:
//
//	sc.RenderToObject(deployment, `
//	  apiVersion: apps/v1
//	  kind: Deployment
//	  metadata:
//	    name: nginx
//	    namespace: default
//	  spec:
//	    replicas: 3
//	    selector:
//	      matchLabels:
//	        app: nginx
//	    template:
//	      metadata:
//	        labels:
//	          app: nginx
//	      spec:
//	        containers:
//	        - name: nginx
//	          image: nginx:latest
//	          ports:
//	          - containerPort: 80
//	`)
func (s *Sawchain) RenderToObject(obj client.Object, template string, bindings ...map[string]any) {
	if util.IsExistingFile(template) {
		var err error
		template, err = util.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errFailedReadTemplate)
	}
	unstructuredObj, err := chainsaw.RenderTemplateSingle(context.TODO(), template, chainsaw.BindingsFromMap(util.MergeMaps(bindings...)))
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)
	s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, obj)).To(gomega.Succeed(), errFailedSave)
}

// TODO: test
// RenderToObjects renders a Chainsaw template with optional bindings into a slice of objects.
//
// This can also be used as a convenience method for unmarshaling static manifests.
//
// Invalid input will result in immediate test failure.
//
// # Arguments
//
//   - Objects ([]client.Object): Slice of typed or unstructured objects to render into. If any objects
//     are typed, the client scheme will be used for conversions.
//
//   - Template (string): File path or content of a static manifest or Chainsaw template to render. Must
//     contain complete resource definitions exactly matching the count, order, and types of the provided
//     objects.
//
//   - Bindings (map[string]any): Bindings to be applied to a Chainsaw template (if provided) in addition to
//     (or overriding) Sawchain's global bindings. If multiple maps are provided, they will be merged in
//     natural order.
//
// # Examples
//
// Render multiple resources from a template using bindings:
//
//	sc.RenderToObjects([]client.Object{configMap, secret}, `
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
// Render multiple resources from a template file using bindings:
//
//	sc.RenderToObjects([]client.Object{deployment, service}, "path/to/template.yaml",
//	  map[string]any{"prefix": "test", "namespace": "default"})
//
// Unmarshal a static multi-resource manifest into objects:
//
//	sc.RenderToObjects([]client.Object{configMap, service}, `
//	  apiVersion: v1
//	  kind: ConfigMap
//	  metadata:
//	    name: app-config
//	    namespace: default
//	  data:
//	    key: value
//	  ---
//	  apiVersion: v1
//	  kind: Service
//	  metadata:
//	    name: app-service
//	    namespace: default
//	  spec:
//	    selector:
//	      app: myapp
//	    ports:
//	    - port: 80
//	      targetPort: 8080
//	`)
func (s *Sawchain) RenderToObjects(objs []client.Object, template string, bindings ...map[string]any) {
	if util.IsExistingFile(template) {
		var err error
		template, err = util.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errFailedReadTemplate)
	}
	unstructuredObjs, err := chainsaw.RenderTemplate(context.TODO(), template, chainsaw.BindingsFromMap(util.MergeMaps(bindings...)))
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)
	s.g.Expect(objs).To(gomega.HaveLen(len(unstructuredObjs)), errObjectsWrongLength)
	for i, unstructuredObj := range unstructuredObjs {
		s.g.Expect(util.CopyUnstructuredToObject(s.c, unstructuredObj, objs[i])).To(gomega.Succeed(), errFailedSave)
	}
}

// TODO: test
// RenderToString renders a Chainsaw template with optional bindings into a YAML string.
//
// Invalid input and marshaling errors will result in immediate test failure.
//
// # Arguments
//
//   - Template (string): File path or content of a Chainsaw template to render.
//
//   - Bindings (map[string]any): Bindings to be applied to the template in addition to (or overriding)
//     Sawchain's global bindings. If multiple maps are provided, they will be merged in natural order.
//
// # Examples
//
// Render resources from a template using bindings:
//
//	yaml := sc.RenderToString(`
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
// Render resources from a template file using bindings:
//
//	yaml := sc.RenderToString("path/to/template.yaml",
//	  map[string]any{"prefix": "test", "namespace": "default"})
func (s *Sawchain) RenderToString(template string, bindings ...map[string]any) string {
	if util.IsExistingFile(template) {
		var err error
		template, err = util.ReadFileContent(template)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errFailedReadTemplate)
	}
	objs, err := chainsaw.RenderTemplate(context.TODO(), template, chainsaw.BindingsFromMap(util.MergeMaps(bindings...)))
	s.g.Expect(err).NotTo(gomega.HaveOccurred(), errInvalidTemplate)
	var buf bytes.Buffer
	for i, obj := range objs {
		y, err := yaml.Marshal(obj.Object)
		s.g.Expect(err).NotTo(gomega.HaveOccurred(), errFailedMarshalObject)
		if i > 0 {
			buf.WriteString("---\n")
		}
		buf.Write(y)
		buf.WriteString("\n")
	}
	return buf.String()
}

// TODO: test
// RenderToFile renders a Chainsaw template with optional bindings and writes it to a file.
//
// Invalid input, marshaling errors, and I/O errors will result in immediate test failure.
//
// # Arguments
//
//   - Filepath (string): The file path where the rendered YAML will be written.
//
//   - Template (string): File path or content of a Chainsaw template to render.
//
//   - Bindings (map[string]any): Bindings to be applied to the template in addition to (or overriding)
//     Sawchain's global bindings. If multiple maps are provided, they will be merged in natural order.
//
// # Examples
//
// Render resources from a template to a file:
//
//	sc.RenderToFile("output.yaml", `
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
// Render resources from a template file to another file:
//
//	sc.RenderToFile("output.yaml", "path/to/template.yaml",
//	  map[string]any{"prefix": "test", "namespace": "default"})
func (s *Sawchain) RenderToFile(filepath, template string, bindings ...map[string]any) {
	rendered := s.RenderToString(template, bindings...)
	s.g.Expect(os.WriteFile(filepath, []byte(rendered), 0644)).To(gomega.Succeed(), errFailedWrite)
}
