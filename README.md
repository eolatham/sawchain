# Sawchain

Sawchain (the part of the chainsaw that cuts things) is a Kubernetes testing library that exposes the power of [Chainsaw](https://github.com/kyverno/chainsaw) YAML assertions through a [Gomega](https://github.com/onsi/gomega)-friendly interface.

## TODO

* Finish Link implementation
* Add docstrings to Link fields, methods, and constructor
* Add tests for Link methods
* Add example test suites using Link
* Finish README (generate docs from Link dosctrings)

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
