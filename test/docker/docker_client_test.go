package docker_test

import (
	"context"
	"log"
	"testing"

	cli "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/docker"
	"github.com/docker/docker/api/types/container"
)

func TestMain(t *testing.T) {
	dockerClient := cli.GetDockerClient()

	// 列出当前运行的容器
	containers, err := dockerClient.ContainerList(
		context.Background(),
		container.ListOptions{},
	)
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	// 打印容器信息
	t.Logf("Currently running containers:")
	for _, container := range containers {
		t.Logf(
			"ID: %s, Image: %s, Names: %v, State: %s, Status: %s",
			container.ID,
			container.Image,
			container.Names,
			container.State,
			container.Status,
		)
	}

	// 清理 Docker 客户端
	if err := dockerClient.Close(); err != nil {
		log.Fatalf("Failed to close Docker client: %v", err)
	}
}
