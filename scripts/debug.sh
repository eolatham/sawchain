#!/bin/bash

set -o errexit -o nounset

source scripts/init.sh

# Run integration tests in debug mode

echo "Debugging tests..."
KUBEBUILDER_ASSETS="$($SETUP_ENVTEST use -p path "$ENVTEST_K8S_VERSION")" \
    dlv test "${PACKAGE}" --listen=:40000 --headless=true --api-version=2
