# Sawchain

Go library for K8s YAML-driven testing—backed by [Chainsaw](https://github.com/kyverno/chainsaw)

## TODO

* Documentation
  * Docstrings
  * README
* Testing
  * All possible argument combinations
  * Structured and unstructured objects
* Implementation
* Example test suites

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
sc.CreateResourceAndWait(ctx, obj)            // Create resource with obj
sc.CreateResourceAndWait(ctx, template)       // Create resource with single-document template, don't save state
sc.CreateResourceAndWait(ctx, obj, template)  // Create resource with single-document template, save state to obj
sc.CreateResourcesAndWait(ctx, objs)            // Create resources with objs
sc.CreateResourcesAndWait(ctx, template)        // Create resources with multi-document template, don't save state
sc.CreateResourcesAndWait(ctx, objs, template)  // Create resources with multi-document template, save state to objs
```

#### Update Resources

```go
// Update resources and wait for client cache to sync
sc.UpdateResourceAndWait(ctx, obj)            // Update resource with obj
sc.UpdateResourceAndWait(ctx, template)       // Update resource with single-document template, don't save state
sc.UpdateResourceAndWait(ctx, obj, template)  // Update resource with single-document template, save state to obj
sc.UpdateResourcesAndWait(ctx, objs)            // Update resources with objs
sc.UpdateResourcesAndWait(ctx, template)        // Update resources with multi-document template, don't save state
sc.UpdateResourcesAndWait(ctx, objs, template)  // Update resources with multi-document template, save state to objs
```

#### Delete Resources

```go
// Delete resources and wait for client cache to sync
sc.DeleteResourceAndWait(ctx, obj)            // Delete resource with obj
sc.DeleteResourceAndWait(ctx, template)       // Delete resource with single-document template, don't save metadata
sc.DeleteResourceAndWait(ctx, obj, template)  // Delete resource with single-document template, save metadata to obj
sc.DeleteResourcesAndWait(ctx, objs)            // Delete resources with objs
sc.DeleteResourcesAndWait(ctx, template)        // Delete resources with multi-document template, don't save metadata
sc.DeleteResourcesAndWait(ctx, objs, template)  // Delete resources with multi-document template, save metadata to objs
```

### Assertion Utilities

[Gomega](https://github.com/onsi/gomega)-friendly APIs to simplify assertions on K8s resources

#### Assert Existence

```go
// Get resources from the cluster
var err error
err = sc.GetResource(ctx, obj)            // Get resource using obj, save state to obj
err = sc.GetResource(ctx, template)       // Get resource using single-document template, don't save state
err = sc.GetResource(ctx, obj, template)  // Get resource using single-document template, save state to obj
err = sc.GetResources(ctx, objs)            // Get resources using objs, save state to objs
err = sc.GetResources(ctx, template)        // Get resources using multi-document template, don't save state
err = sc.GetResources(ctx, objs, template)  // Get resources using multi-document template, save state to objs

// Assert existence immediately
Expect(sc.GetResource(ctx, template)).To(Succeed())
Expect(sc.GetResources(ctx, template)).To(Succeed())

// Assert existence eventually
Eventually(sc.GetResourceFunc(ctx, template)).Should(Succeed())
Eventually(sc.GetResourcesFunc(ctx, template)).Should(Succeed())
```

#### Assert State

```go
// Get resources from the cluster and return state for matching
var fetched client.Object
fetched = sc.FetchResource(ctx, obj)            // Fetch resource using obj, save state to obj
fetched = sc.FetchResource(ctx, template)       // Fetch resource using single-document template, don't save state
fetched = sc.FetchResource(ctx, obj, template)  // Fetch resource using single-document template, save state to obj
var fetchedList []client.Object
fetchedList = sc.FetchResources(ctx, objs)            // Fetch resources using objs, save state to objs
fetchedList = sc.FetchResources(ctx, template)        // Fetch resources using multi-document template, don't save state
fetchedList = sc.FetchResources(ctx, objs, template)  // Fetch resources using multi-document template, save state to objs

// Assert state immediately
Expect(sc.FetchResource(ctx, template)).To(HaveField("Foo", "Bar"))
Expect(sc.FetchResources(ctx, template)).To(ConsistOf(HaveField("Foo", "Bar")))

// Assert state eventually
Eventually(sc.FetchResourceFunc(ctx, template)).Should(HaveField("Foo", "Bar"))
Eventually(sc.FetchResourcesFunc(ctx, template)).Should(ConsistOf(HaveField("Foo", "Bar")))

// Custom matchers (single resource only)
Expect(obj).To(sc.MatchYAML(template))                    // Assert client.Object (structured or not) matches Chainsaw template
Expect(obj).To(sc.HaveStatusCondition("Type", "Status"))  // Assert client.Object (structured or not) has specific status condition
```

#### Assert (Almost) Anything

```go
// Check cluster for a matching resource
var err error
err = sc.CheckResource(ctx, obj)            // Check for exact match with obj
err = sc.CheckResource(ctx, template)       // Execute Chainsaw check with single-document template
err = sc.CheckResource(ctx, obj, template)  // Execute Chainsaw check with single-document template, save first match to obj
err = sc.CheckResources(ctx, objs)            // Check for exact match with each object in objs
err = sc.CheckResources(ctx, template)        // Execute Chainsaw check with each document in template
err = sc.CheckResources(ctx, objs, template)  // Execute Chainsaw check with each document in template, save first matches to objs

// Assert match found immediately
Expect(sc.CheckResource(ctx, template)).To(Succeed())
Expect(sc.CheckResources(ctx, template)).To(Succeed())

// Assert match found eventually
Eventually(sc.CheckResourcesFunc(ctx, template)).Should(Succeed())
Eventually(sc.CheckResourceFunc(ctx, template)).Should(Succeed())
```

### Notes

* Template documents used in create and update operations must contain complete resource definitions
* Template documents used in get and delete operations must contain valid resource keys
