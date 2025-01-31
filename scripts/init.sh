#!/bin/bash

set -o errexit -o nounset

# Install go dependencies

echo "Installing go dependencies..."
go mod tidy

# Install binaries

echo "Setting up local bin directory..."
LOCALBIN=$(pwd)/bin
if [ ! -d "$LOCALBIN" ]; then
    mkdir -p $LOCALBIN
fi

echo "Installing setup-envtest..."
ENVTEST_K8S_VERSION="1.31.x"
SETUP_ENVTEST=$LOCALBIN/setup-envtest
GOBIN=$LOCALBIN go install sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20250130183723-1a91ccca639b
