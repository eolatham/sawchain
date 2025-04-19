# KubeVela Offline Render Test

This example demonstrates how Sawchain can be used to do offline render testing for [KubeVela](https://kubevela.io/docs/) components and traits

## Details

Uses the [webservice](https://kubevela.io/docs/end-user/components/references/#webservice) component with [annotations](https://kubevela.io/docs/end-user/traits/references/#annotations) and [gateway](https://kubevela.io/docs/end-user/traits/references/#gateway) traits

Includes positive and negative test cases and uses Chainsaw templating to reuse input and expectation YAMLs

For each test case:

1. Renders input files and expected output using Sawchain's `RenderToFile` and `RenderToString` utilities
1. Runs `vela dry-run --offline` and parses the output into unstructured K8s objects
1. Verifies the rendered resources match the expected output using Sawchain's `MatchYAML` matcher

## Run

Ensure you have the [KubeVela CLI](https://kubevela.io/docs/installation/kubernetes/) installed, then:

```bash
# Navigate to the project root
cd /path/to/sawchain

# Run the test
go test -v ./examples/offline-render-tests/kubevela
```
