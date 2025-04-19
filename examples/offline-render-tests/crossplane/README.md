# Crossplane Offline Render Test

This example demonstrates how Sawchain can be used to do offline render testing for Crossplane function-based [compositions](https://docs.crossplane.io/latest/concepts/compositions/)

## Details

Uses the ExtraResources example composition from the [function-go-templating repository](https://github.com/crossplane-contrib/function-go-templating/tree/main/example/extra-resources)

Includes positive and negative test cases and uses Chainsaw templating to reuse input and expectation YAMLs

For each test case:

1. Renders input files and expected output using Sawchain's `RenderToFile` and `RenderToString` utilities
1. Runs `crossplane render` and parses the output into unstructured K8s objects
1. Verifies the rendered resources match the expected output using Sawchain's `MatchYAML` matcher

## Run

Ensure you have Docker running and the [Crossplane CLI](https://docs.crossplane.io/latest/cli/) installed, then:

```bash
# Navigate to the project root
cd /path/to/sawchain

# Run the test
go test -v ./examples/offline-render-tests/crossplane
```
