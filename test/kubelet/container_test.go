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

func setupNetNS(nsName string) {
	cmd := exec.Command("ip", "netns", "add", nsName)
	_ = cmd.Run()
}

func cleanupNetNS(nsName string) {
	cmd := exec.Command("ip", "netns", "delete", nsName)
	_ = cmd.Run()
}

func TestCreateContainer(t *testing.T) {
	// 定义网络命名空间路径
	setupNetNS("test-ns")

	netNSPath := "/var/run/netns/test-ns"

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

	info, err := kubelet.GetContainerInfo(
		ctx,
		client,
		containerSpec.Name,
	)

	if err != nil {
		t.Fatalf("failed to get container info: %v", err)
	}

	// 检查容器信息
	if info == nil {
		t.Fatalf("container info is nil")
	}

	if info.Name != containerSpec.Name {
		t.Fatalf(
			"expected container name %s, got %s",
			containerSpec.Name,
			info.Name,
		)
	}

	t.Logf("Container Info: %v", info)

	// 停止容器
	err = kubelet.StopContainer(ctx, client, containerSpec.Name)
	if err != nil {
		t.Fatalf("failed to stop container: %v", err)
	}

	// 删除容器
	err = kubelet.DeleteContainer(ctx, client, containerSpec)
	if err != nil {
		t.Fatalf("failed to delete container: %v", err)
	}

	cleanupNetNS("test-ns")
}
