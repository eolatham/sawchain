# K8s Test Helper

A Go library that makes testing K8s resources and operators easy

TODO: consider renaming to chainsaw-helper, chainsaw-wrapper, or sawchain (the part of the chainsaw that actually cuts things)

## Packages

### [pkg/chainsaw](./pkg/chainsaw/)

TODO: document and give examples

### [pkg/helper](./pkg/helper/)

TODO:

- document and give examples
- consider additional options for unique name and namespace generation
- consider exposing global bindings via a helper field or method

## Testing

To run all tests, use `make test`

```sh
make test
```

To run tests for a specific package in debug mode, use `make debug` with `PACKAGE` set to the package path

```sh
PACKAGE=./pkg/chainsaw make debug
```
