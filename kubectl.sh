#/bin/bash

# This script is used to run `kubectl` commands
# without needing to specify `go run` each time.

# Get the path of the current script
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Get the `kubectl.go` file path
GO_FILE="$SCRIPT_DIR/pkg/kubectl/kubectl.go"

# Execute the `kubectl.go` file using `go run`
go run "$GO_FILE" $@
