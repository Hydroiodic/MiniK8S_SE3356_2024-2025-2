package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	client := apiserver.NewAPIClient("")

	// Delete a pod
	pod := object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "nginx",
			Namespace: "default",
			Labels:    map[string]string{"app": "nginx"},
		},
		Spec: object.PodSpec{
			PauseContainerID: "",
			RestartPolicy:    "",
			Containers:       []object.Container{},
			Volumes:          []object.Volume{},
		},
		Status: object.PodStatus{
			Phase: object.PodRunning,
		},
	}

	// Delete the pod using the client
	if err := client.DeletePod(&pod); err != nil {
		panic(err)
	}
}
