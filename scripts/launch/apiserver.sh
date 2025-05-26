#!/bin/bash

# This script is used to start the `apiserver` component.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the root directory of the project
ROOT_DIR=$SCRIPT_DIR/../..

# Get the `apiserver.go` file path
APISERVER_FILE="$ROOT_DIR/pkg/apiserver/main/main.go"

# Execute the `apiserver.go` file using `go run`
go run "$APISERVER_FILE" $@
