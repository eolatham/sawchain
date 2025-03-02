# Sawchain

Sawchain (the part of the chainsaw that cuts things) is a Kubernetes testing library that exposes the power of [Chainsaw](https://github.com/kyverno/chainsaw) YAML assertions through a [Gomega](https://github.com/onsi/gomega)-friendly interface.

## TODO

* Written documentation
  * Project description
  * Docstrings
  * README
* Link testing
  * All possible argument combinations
  * Structured and unstructured objects
* Link implementation
* Example test suites

## Link

Use Sawchain's incisive Link object to write YAML-driven Kubernetes tests in Go.

### Initialization

```go
link := sawchain.NewLink(
  t,                             // testing.TB for internal assertions
  testClient,                    // K8s client for internal API calls
  map[string]any{"foo", "bar"},  // Global bindings to apply to all template operations
  "10s",                         // Default Eventually timeout
  "1s",                          // Default Eventually polling interval
)
```

### Setup/Cleanup Utilities

#### Create Resources

```go
// Create resources and wait for client cache to sync
link.CreateResourceAndWait(ctx, obj)            // Create resource with obj
link.CreateResourceAndWait(ctx, template)       // Create resource with single-document template, don't save state
link.CreateResourceAndWait(ctx, obj, template)  // Create resource with single-document template, save state to obj
link.CreateResourcesAndWait(ctx, objList)            // Create resources with objList
link.CreateResourcesAndWait(ctx, template)           // Create resources with multi-document template, don't save state
link.CreateResourcesAndWait(ctx, objList, template)  // Create resources with multi-document template, save state to objList
```

#### Update Resources

```go
// Update resources and wait for client cache to sync
link.UpdateResourceAndWait(ctx, obj)            // Update resource with obj
link.UpdateResourceAndWait(ctx, template)       // Update resource with single-document template, don't save state
link.UpdateResourceAndWait(ctx, obj, template)  // Update resource with single-document template, save state to obj
link.UpdateResourcesAndWait(ctx, objList)            // Update resources with objList
link.UpdateResourcesAndWait(ctx, template)           // Update resources with multi-document template, don't save state
link.UpdateResourcesAndWait(ctx, objList, template)  // Update resources with multi-document template, save state to objList
```

#### Delete Resources

```go
// Delete resources and wait for client cache to sync
link.DeleteResourceAndWait(ctx, obj)            // Delete resource with obj
link.DeleteResourceAndWait(ctx, template)       // Delete resource with single-document template, don't save metadata
link.DeleteResourceAndWait(ctx, obj, template)  // Delete resource with single-document template, save metadata to obj
link.DeleteResourcesAndWait(ctx, objList)            // Delete resources with objList
link.DeleteResourcesAndWait(ctx, template)           // Delete resources with multi-document template, don't save metadata
link.DeleteResourcesAndWait(ctx, objList, template)  // Delete resources with multi-document template, save metadata to objList
```

### Assertion Utilities

#### Assert Existence

```go
// Get resources from the cluster
var err error
err = link.GetResource(ctx, obj)            // Get resource using obj, save state to obj
err = link.GetResource(ctx, template)       // Get resource using single-document template, don't save state
err = link.GetResource(ctx, obj, template)  // Get resource using single-document template, save state to obj
err = link.GetResources(ctx, objList)            // Get resources using objList, save state to objList
err = link.GetResources(ctx, template)           // Get resources using multi-document template, don't save state
err = link.GetResources(ctx, objList, template)  // Get resources using multi-document template, save state to objList

// Assert existence immediately
Expect(link.GetResource(ctx, template)).To(Succeed())
Expect(link.GetResources(ctx, template)).To(Succeed())

// Assert existence eventually
Eventually(link.GetResourceFunc(ctx, template)).Should(Succeed())
Eventually(link.GetResourcesFunc(ctx, template)).Should(Succeed())
```

#### Assert State

```go
// Get resources from the cluster and return state for matching
var fetched client.Object
fetched = link.FetchResource(ctx, obj)            // Fetch resource using obj, save state to obj
fetched = link.FetchResource(ctx, template)       // Fetch resource using single-document template, don't save state
fetched = link.FetchResource(ctx, obj, template)  // Fetch resource using single-document template, save state to obj
var fetchedList []client.Object
fetchedList = link.FetchResources(ctx, objList)            // Fetch resources using objList, save state to objList
fetchedList = link.FetchResources(ctx, template)           // Fetch resources using multi-document template, don't save state
fetchedList = link.FetchResources(ctx, objList, template)  // Fetch resources using multi-document template, save state to objList

// Assert state immediately
Expect(link.FetchResource(ctx, template)).To(HaveField("Foo", "Bar"))
Expect(link.FetchResources(ctx, template)).To(ConsistOf(HaveField("Foo", "Bar")))

// Assert state eventually
Eventually(link.FetchResourceFunc(ctx, template)).Should(HaveField("Foo", "Bar"))
Eventually(link.FetchResourcesFunc(ctx, template)).Should(ConsistOf(HaveField("Foo", "Bar")))

// Custom matchers (single resource only)
Expect(obj).To(link.Match(template))                        // Assert client.Object (structured or not) matches Chainsaw template
Expect(obj).To(link.HaveStatusCondition("Type", "Status"))  // Assert client.Object (structured or not) has specific status condition
```

#### Assert Existence/State

```go
// Check cluster for a matching resource
var err error
err = link.CheckResource(ctx, obj)            // Check for exact match with obj
err = link.CheckResource(ctx, template)       // Execute Chainsaw check with single-document template
err = link.CheckResource(ctx, obj, template)  // Execute Chainsaw check with single-document template, save first match to obj
err = link.CheckResources(ctx, objList)            // Check for exact match with each object in objList
err = link.CheckResources(ctx, template)           // Execute Chainsaw check with each document in template
err = link.CheckResources(ctx, objList, template)  // Execute Chainsaw check with each document in template, save first matches to objList

// Assert match found immediately
Expect(link.CheckResource(ctx, template)).To(Succeed())
Expect(link.CheckResources(ctx, template)).To(Succeed())

// Assert match found eventually
Eventually(link.CheckResourcesFunc(ctx, template)).Should(Succeed())
Eventually(link.CheckResourceFunc(ctx, template)).Should(Succeed())
```

### Notes

* Template documents used in create and update operations must contain complete resource definitions.
* Template documents used in get and delete operations must contain valid resource keys.
