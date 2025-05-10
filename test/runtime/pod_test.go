package runtime_test

import (
	"testing"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestCreatePod(t *testing.T) {
	containerService, err := container.NewContainerService()
	if err != nil {
		t.Fatalf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	pod := &object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "test-pod",
			Namespace: "example-pod-namespace",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:  "test-container",
					Image: "docker.io/library/nginx:latest",
					// Image:   "nginx:latest",
					Command: []string{"nginx", "-g", "daemon off;"},
					Ports: []object.ContainerPort{
						{ContainerPort: 80},
					},
				},
			},
		},
		// Status: object.PodStatus{
		// 	Phase:      "Pending",
		// 	StartTime:  time.Now(),
		// 	Conditions: []string{"Initialized", "Ready"},
		// },
	}

	err = podService.CreatePod(pod)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	t.Logf("Pod created: %v", pod)

	containerId := pod.Spec.Containers[0].ID
	assert.NotEmpty(t, containerId, "Container ID should not be empty")

	// 检查 Pause 容器是否存在
	res, err := containerService.GetContainerInfo(containerId)

	if err != nil {
		t.Fatalf("failed to get pause container info: %v", err)
	}

	t.Logf("Pause container info: %v", res)
}
