# Sawchain

Go library for K8s YAML-driven testing—backed by [Chainsaw](https://github.com/kyverno/chainsaw)

## TODO

* finish testing
* finish documentation
* add branding
* add license
* add examples
  * Crossplane render test
  * KubeVela dry-run test
  * K8s helper function integration test (using fakeclient)
  * K8s operator end-to-end test (using envtest)
  * Helm install smoke test (using envtest)

## Usage

### Initialization

```go
sc := sawchain.New(
  t,                             // testing.TB for internal assertions
  testClient,                    // K8s client for internal API calls
  map[string]any{"foo", "bar"},  // Global bindings to apply to all template operations
  "10s",                         // Default Eventually timeout
  "1s",                          // Default Eventually polling interval
)
```

### Setup/Cleanup Utilities

Helpers to reliably create/update/delete test K8s resources

#### Create Resources

```go
// Create resources and wait for client cache to sync
sc.Create(ctx, obj)             // Create resource with obj
sc.Create(ctx, template)        // Create resource(s) with template, don't save state
sc.Create(ctx, obj, template)   // Create resource with single-document template, save state to obj
sc.Create(ctx, objs)            // Create resources with objs
sc.Create(ctx, objs, template)  // Create resources with multi-document template, save state to objs
```

#### Update Resources

```go
// Update resources and wait for client cache to sync
sc.Update(ctx, obj)             // Update resource with obj
sc.Update(ctx, template)        // Update resource(s) with template, don't save state
sc.Update(ctx, obj, template)   // Update resource with single-document template, save state to obj
sc.Update(ctx, objs)            // Update resources with objs
sc.Update(ctx, objs, template)  // Update resources with multi-document template, save state to objs
```

#### Delete Resources

```go
// Delete resources and wait for client cache to sync
sc.Delete(ctx, obj)             // Delete resource with obj
sc.Delete(ctx, template)        // Delete resource(s) with template, don't save state
sc.Delete(ctx, obj, template)   // Delete resource with single-document template, save metadata to obj
sc.Delete(ctx, objs)            // Delete resources with objs
sc.Delete(ctx, objs, template)  // Delete resources with multi-document template, save metadata to objs
```

### Assertion Utilities

[Gomega](https://github.com/onsi/gomega)-friendly APIs to simplify assertions on K8s resources

#### Assert Existence

```go
// Get resources from the cluster
var err error
err = sc.Get(ctx, obj)             // Get resource using obj, save state to obj
err = sc.Get(ctx, template)        // Get resource(s) using template, don't save state
err = sc.Get(ctx, obj, template)   // Get resource using single-document template, save state to obj
err = sc.Get(ctx, objs)            // Get resources using objs, save state to objs
err = sc.Get(ctx, objs, template)  // Get resources using multi-document template, save state to objs

// Assert existence immediately
Expect(sc.Get(ctx, template)).To(Succeed())

// Assert existence eventually
Eventually(sc.GetFunc(ctx, template)).Should(Succeed())
```

#### Assert State

```go
// Get resources from the cluster and return state for matching
var fetched client.Object
fetched = sc.FetchSingle(ctx, obj)            // Fetch resource using obj, save state to obj
fetched = sc.FetchSingle(ctx, template)       // Fetch resource using single-document template, don't save state
fetched = sc.FetchSingle(ctx, obj, template)  // Fetch resource using single-document template, save state to obj
var fetchedList []client.Object
fetchedList = sc.FetchMultiple(ctx, objs)            // Fetch resources using objs, save state to objs
fetchedList = sc.FetchMultiple(ctx, template)        // Fetch resources using multi-document template, don't save state
fetchedList = sc.FetchMultiple(ctx, objs, template)  // Fetch resources using multi-document template, save state to objs

// Assert state immediately
Expect(sc.FetchSingle(ctx, template)).To(HaveField("Foo", "Bar"))
Expect(sc.FetchMultiple(ctx, template)).To(ConsistOf(HaveField("Foo", "Bar")))

// Assert state eventually
Eventually(sc.FetchSingleFunc(ctx, template)).Should(HaveField("Foo", "Bar"))
Eventually(sc.FetchMultipleFunc(ctx, template)).Should(ConsistOf(HaveField("Foo", "Bar")))

// Custom matchers (single resource only)
Expect(obj).To(sc.MatchYAML(template))                    // Assert client.Object matches Chainsaw template
Expect(obj).To(sc.HaveStatusCondition("Type", "Status"))  // Assert client.Object has specific status condition
```

#### Assert (Almost) Anything

```go
// Check cluster for a matching resource
var err error
err = sc.Check(ctx, template)        // Execute Chainsaw check with template
err = sc.Check(ctx, obj, template)   // Execute Chainsaw check with single-document template, save first match to obj
err = sc.Check(ctx, objs, template)  // Execute Chainsaw check with each document in template, save first matches to objs

// Assert match found immediately
Expect(sc.Check(ctx, template)).To(Succeed())

// Assert match found eventually
Eventually(sc.CheckFunc(ctx, template)).Should(Succeed())
```

### Templating Utilities

Helpers to easily render Chainsaw templates into objects, strings, or files

```go
sc.RenderToObject(obj, template, bindings)
sc.RenderToObjects(objs, template, bindings)
s := sc.RenderToString(template, bindings)
sc.RenderToFile(filepath, template, bindings)
```

### Notes

* Sawchain accepts [client.Object](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#Object) inputs (typed or unstructured) and maintains object state in the original input format, relying on the client [scheme](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime#Scheme) to perform internal type conversions when needed.
* Fetch operations attempt to return typed objects when no input objects are provided.
* Template documents used in create and update operations must contain complete resource definitions.
* Template documents used in get and delete operations must contain valid resource keys.
* Template documents used in render operations must contain complete resource definitions when rendering into typed objects.
