package kubelet_test

import (
	"context"
	"testing"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd/namespaces"
)

func TestCreatePod(t *testing.T) {

	inst := kubelet.NewKubeletInstance()
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
					Name:    "test-container",
					Image:   "docker.io/library/nginx:latest",
					Command: []string{"nginx", "-g", "daemon off;"},
					Ports: []object.ContainerPort{
						{ContainerPort: 80},
					},
				},
			},
		},
		Status: object.PodStatus{
			Phase:      "Pending",
			StartTime:  time.Now(),
			Conditions: []string{"Initialized", "Ready"},
		},
	}

	err := inst.CreatePod(context.Background(), pod)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	// 检查 Pause 容器是否存在
	cli := inst.GetContainerdClient()
	info, err := kubelet.GetContainerInfo(
		namespaces.WithNamespace(context.Background(), pod.Metadata.Namespace),
		cli,
		"test-pod-pause", // Pause 容器名称为 "Pod名-pause"
	)

	if err != nil {
		t.Fatalf("failed to get pause container info: %v", err)
	}

	if info == nil {
		t.Fatalf("pause container not found")
	}

	// 测试Nginx容器运行情况
	info, err = kubelet.GetContainerInfo(
		namespaces.WithNamespace(context.Background(), pod.Metadata.Namespace),
		cli,
		"test-container",
	)

	if err != nil {
		t.Fatalf("failed to get container info: %v", err)
	}

	if info == nil {
		t.Fatalf("container not found")
	}

}
