#!/bin/bash

set -o errexit -o nounset

source scripts/init.sh

# Run integration tests with envtest

echo "Running tests..."
KUBEBUILDER_ASSETS="$($SETUP_ENVTEST use -p path "$ENVTEST_K8S_VERSION")" go test ./...
