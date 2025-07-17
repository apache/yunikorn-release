#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Script to install clusterloader2 binary
# Based on: https://github.com/kubernetes/perf-tests/blob/master/clusterloader2/docs/GETTING_STARTED.md#clusterloader2

set -e

echo "Installing clusterloader2..."

# Check if clusterloader2 is already installed
if command -v clusterloader2 &> /dev/null; then
    echo "✅ clusterloader2 is already installed!"
    echo "Current version information:"
    clusterloader2 --version || echo "clusterloader2 binary is ready to use"
    echo ""
    echo "To reinstall, remove the existing binary first:"
    EXISTING_PATH=$(which clusterloader2)
    echo "  rm $EXISTING_PATH"
    echo ""
    echo "🎉 clusterloader2 installation check completed successfully!"
    exit 0
fi

# Check if Go and Git are installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is required to build clusterloader2. Please install Go first."
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

if ! command -v git &> /dev/null; then
    echo "ERROR: Git is required to clone the repository. Please install Git first."
    exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
echo "Using temporary directory: $TEMP_DIR"

# Cleanup function
cleanup() {
    echo "Cleaning up temporary directory..."
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "Cloning kubernetes/perf-tests repository..."
git clone --depth 1 https://github.com/kubernetes/perf-tests.git "$TEMP_DIR/perf-tests"

cd "$TEMP_DIR/perf-tests/clusterloader2"
echo "Building clusterloader2..."
go build -o clusterloader2 cmd/clusterloader.go

# Determine installation directory using Go environment
GOBIN_DIR=$(go env GOBIN)
if [[ -n "$GOBIN_DIR" ]]; then
    INSTALL_DIR="$GOBIN_DIR"
else
    GOPATH_DIR=$(go env GOPATH)
    if [[ -z "$GOPATH_DIR" ]]; then
        echo "ERROR: Neither GOBIN nor GOPATH is set. Please configure your Go environment."
        exit 1
    fi
    INSTALL_DIR="$GOPATH_DIR/bin"
fi

mkdir -p "$INSTALL_DIR"

echo "Installing clusterloader2 to $INSTALL_DIR..."
cp clusterloader2 "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/clusterloader2"

echo "Verifying clusterloader2 installation..."
if command -v clusterloader2 &> /dev/null; then
    echo "✅ clusterloader2 successfully installed!"
    echo "Version information:"
    clusterloader2 --version || echo "clusterloader2 binary is ready to use"
    echo ""
    echo "Usage example:"
    echo "  clusterloader2 --testconfig=config.yaml --provider=kind --kubeconfig=~/.kube/config"
else
    echo "❌ ERROR: clusterloader2 installation failed"
    echo "The binary may not be in your PATH. Make sure Go binaries are in your PATH:"
    echo "  export PATH=\$PATH:$INSTALL_DIR"
    echo ""
    echo "Alternatively, you can add the Go binary directory to your PATH permanently:"
    echo "  echo 'export PATH=\$PATH:$INSTALL_DIR' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    exit 1
fi

echo ""
echo "🎉 clusterloader2 installation completed successfully!"
echo ""
