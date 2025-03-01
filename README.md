# Sawchain

Sawchain (the part of the chainsaw that cuts things) is a Kubernetes testing library that exposes the power of [Chainsaw](https://github.com/kyverno/chainsaw) YAML assertions through a [Gomega](https://github.com/onsi/gomega)-friendly interface.

## TODO

* Implement brainstorm
* Add docstrings to Link fields, methods, and constructor
* Add tests for Link methods
  * Structured and unstructured objects
  * With and without templates
  * With and without bindings
  * With and without custom timeout/interval settings
* Add example test suites using Link
* Finish README (generate docs from Link dosctrings)

## Brainstorm

```go
// Configure default timeout and interval, global bindings
link := sawchain.NewLink(t, testClient)

// Create/update/delete and wait for client cache to sync
// Operate from client.Object or YAML template
// Save resulting state to client.Object
// Create and update require the given client.Object or template to contain a valid resource definition
// Delete requires the given client.Object or template to contain a valid resource key
link.CreateResourceAndWait(ctx, obj, template)
link.UpdateResourceAndWait(ctx, obj, template)
link.DeleteResourceAndWait(ctx, obj, template)

// Assert immediate existence
Expect(link.GetResource(ctx, obj, template)).To(Succeed())

// Assert immediate state
Expect(link.FetchResource(ctx, obj, template)).To(HaveField("Foo", "Bar"))

// Assert eventual existence
Eventually(link.GetResourceFunc(ctx, obj, template)).Should(Succeed())

// Assert eventual state
Eventually(link.FetchResourceFunc(ctx, obj, template)).Should(HaveField("Foo", "Bar"))

// Get semantics
// Operate from client.Object or YAML template
// Save resulting state to client.Object
// Return error from K8s get operation
// Requires the given client.Object or template to contain a valid resource key

// Fetch semantics
// Operate from client.Object or YAML template
// Save resulting state to client.Object
// Return updated client.Object
// Requires the given client.Object or template to contain a valid resource key

// Custom matchers
link.MatchYAML(template)
link.HaveNewGeneration(obj)
link.HaveNewResourceVersion(obj)
link.HaveStatusCondition("Type", "Status")

// MatchYAML semantics
// Use Chainsaw to compare the actual client.Object with expectations defined in a YAML template
// Requires the given template to contain assertions for a single resource

/*
MatchYAML (and each other utility brainstormed so far) is focused on a single given object

We also need broader utilities that support full Chainsaw assert and error operations for K8s resources
- single candidate
- multiple candidates
    - assert all match, save all
    - assert one matches, save all
    - assert none match, save all
- multiple documents
*/
```

## Usage

```go
// Create Link
link := sawchain.NewLink(t, testClient,
  sawchain.WithBinding("binding1", "value1"),
  sawchain.WithBinding("binding2", "value2"),
  sawchain.WithTimeout("10s"),
  sawchain.WithInterval("1s"))
```

```go
// Create resource and wait for cache to sync
link.SafeCreate(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))

// Update resource and wait for cache to sync
link.SafeUpdate(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))

// Delete resource and wait for cache to sync
link.SafeDelete(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))
```

```go
// Assert resource existence
Eventually(link.Get(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))).
  Should(Succeed())

// Assert resource state - pure Gomega
Eventually(link.GetObject(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))).
  Should(HaveField("Status.ReadyReplicas", 3))

// Assert resource state - using Chainsaw
Eventually(link.Check(ctx, obj,
  sawchain.WithTemplateFile("template.yaml"),
  sawchain.WithBinding("binding3", "value3"))).
  Should(Succeed())
```

## Options

```go
// Options - Link
sawchain.WithBinding("name", "value")
sawchain.WithTimeout("10s")
sawchain.WithInterval("1s")
```

```go
// Options - SafeCreate, SafeUpdate, SafeDelete
sawchain.WithTemplateContent(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
data:
  key: value
`)
sawchain.WithTemplateFile("template.yaml")
sawchain.WithBinding("name", "value")
sawchain.WithTimeout("10s")
sawchain.WithInterval("1s")
```

```go
// Options - Get, GetObject, Check
sawchain.WithTemplateContent(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: default
data:
  key: value
`)
sawchain.WithTemplateFile("template.yaml")
sawchain.WithBinding("name", "value")
```
