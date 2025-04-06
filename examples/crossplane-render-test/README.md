# Crossplane Render Test

This example demonstrates how Sawchain can be used to test YAML templating and transformation tools like Crossplane's [function-based composition engine](https://docs.crossplane.io/latest/concepts/compositions/)

## Details

Uses the ExtraResources example composition from the [function-go-templating repository](https://github.com/crossplane-contrib/function-go-templating/tree/main/example/extra-resources)

Includes multiple test cases and uses Chainsaw templating to reuse expectation YAMLs

For each test case:

1. Uses the `crossplane render` command to render a sample XR with extra resources using [function-go-templating](https://github.com/crossplane-contrib/function-go-templating)
1. Parses the output into unstructured Kubernetes objects
1. Uses Sawchain's `MatchYAML` matcher to verify the rendered resources match the expected output

## Run

Ensure you have Docker running and the [Crossplane CLI](https://docs.crossplane.io/latest/cli/) installed, then:

```bash
# Navigate to the project root
cd /path/to/sawchain

# Run the test
go test -v ./examples/crossplane-render-test
```
