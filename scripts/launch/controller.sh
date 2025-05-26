#!/bin/bash

# This script is used to start the `controller` component.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the root directory of the project
ROOT_DIR=$SCRIPT_DIR/../..

# Get the `controller.go` file path
CONTROLLER_FILE="$ROOT_DIR/pkg/controller/main/main.go"

# Execute the `controller.go` file using `go run`
go run "$CONTROLLER_FILE" $@
