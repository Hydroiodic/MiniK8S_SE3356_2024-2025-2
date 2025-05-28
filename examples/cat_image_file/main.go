package main

import (
	"fmt"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/image"
)

func main() {
	// Read the contents of /etc/group from the "ubuntu:latest" image.
	content, err := image.ReadFileFromImage("ubuntu:latest", "/etc/group")
	if err != nil {
		panic(err)
	}

	// Print the contents of /etc/group.
	fmt.Println("Contents of /etc/group:\n" + content)
}
