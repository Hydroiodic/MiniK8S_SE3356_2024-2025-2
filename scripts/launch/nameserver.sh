#!/bin/bash

# This script is used to start the `nameserver` component.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the root directory of the project
ROOT_DIR=$SCRIPT_DIR/../..

# Get the `nameserver.go` file path
NAMESERVER_FILE="$ROOT_DIR/pkg/kubeproxy/nameserver/main.go"

# Execute the `nameserver.go` file using `go run`
go run "$NAMESERVER_FILE" $@
