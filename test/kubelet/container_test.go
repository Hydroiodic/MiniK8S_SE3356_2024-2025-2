package kubelet_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/containerd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd/namespaces"
)

func setupNetNS(t *testing.T) string {
	nsPath := "/var/run/netns/test-ns"
	cmd := exec.Command("ip", "netns", "add", "test-ns")
	err := cmd.Run()

	if err != nil {
		t.Fatalf("Failed to create network namespace: %v", err)
	}

	return nsPath
}

func TestCreateContainer(t *testing.T) {
	// 定义网络命名空间路径
	netNSPath := setupNetNS(t)

	// 创建一个新的 containerd 客户端
	client, err := containerd.NewContainerdClient()
	if err != nil {
		t.Fatalf("failed to create containerd client: %v", err)
	}
	defer client.Close()

	// 创建一个新的上下文
	ctx := namespaces.WithNamespace(context.Background(), "example_pod_name")

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "docker.io/library/nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 调用 CreateContainer 函数
	err = kubelet.CreateContainer(ctx, client, containerSpec, netNSPath)
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	err = kubelet.DeleteContainer(ctx, client, containerSpec)
	if err != nil {
		t.Fatalf("failed to delete container: %v", err)
	}
}
