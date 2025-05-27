#!/bin/bash

# This script is used to start the `scheduler` component.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the root directory of the project
ROOT_DIR=$SCRIPT_DIR/../..

# Get the `scheduler.go` file path
SCHEDULER_FILE="$ROOT_DIR/pkg/scheduler/main/main.go"

# Execute the `scheduler.go` file using `go run`
go run "$SCHEDULER_FILE" $@
