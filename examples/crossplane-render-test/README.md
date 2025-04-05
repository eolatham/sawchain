# Sawchain Example: Crossplane Render Test

This example demonstrates how Sawchain can be used to test YAML templating and transformation tools like Crossplane's [function-based composition engine](https://docs.crossplane.io/latest/concepts/compositions/)

## Composition Source

This test uses the ExtraResources composition example from the [function-go-templating repository](https://github.com/crossplane-contrib/function-go-templating/tree/main/example/extra-resources)

## Run Test

Ensure you have Docker running and the [Crossplane CLI](https://docs.crossplane.io/latest/cli/) installed, then do the following:

```bash
# Navigate to the project root
cd /path/to/sawchain

# Run the test
go test ./examples/crossplane-render-test/... -v
```

## What It Does

This test:

1. Uses the `crossplane render` command to render a sample XR with extra resources using function-go-templating (v0.9.0)
2. Parses the output into unstructured Kubernetes objects
3. Uses Sawchain's `MatchYAML` matcher to verify the rendered resources match the expected output
