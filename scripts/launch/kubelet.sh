#!/bin/bash

# This script is used to start the `kubelet` component.
# NOTE: This script should be run with root privileges.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the root directory of the project
ROOT_DIR=$SCRIPT_DIR/../..

# Get the `kubelet.go` file path
KUBELET_FILE="$ROOT_DIR/pkg/kubelet/main/main.go"

# Execute the `kubelet.go` file using `go run`
go run "$KUBELET_FILE" $@
