package interfaces

import (
	"path"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func retrievePodsName(pods []object.Pod) []string {
	// Create a slice to store the pod names.
	podNames := make([]string, 0)

	// Iterate over the pods and append their names to the slice.
	for _, pod := range pods {
		podNames = append(
			podNames,
			path.Join(pod.Metadata.Namespace, pod.Metadata.Name),
		)
	}

	return podNames
}
