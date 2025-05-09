package container

import (
	"context"
	"fmt"
	"log"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type ContainerService interface {
	CreateContainer(container object.Container) error
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	GetContainerInfo(containerID string) (*ContainerInfo, error)
	GetContainerStatus(containerID string) (string, error)
}

type containerService struct {
	client *client.Client
}

func NewContainerService() (*containerService, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &containerService{client: cli}, nil
}

func (cs *containerService) CreateContainer(container object.Container) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("无法创建 Docker 客户端: %v", err)
	}
	defer cli.Close()

	// 拉取镜像
	_, err = cli.ImagePull(ctx, container.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("无法拉取镜像 %s: %v", container.Image, err)
	}

	// 检查容器是否已存在
	if _, err := cli.ContainerInspect(ctx, container.Name); err == nil {
		// 容器存在，删除
		if err := cs.DeleteContainer(container.Name); err != nil {
			return fmt.Errorf("无法删除现有容器 %s: %v", container.Name, err)
		}
	}

	// 配置容器
	config := &container.Config{
		Image: container.Image,
		Cmd:   container.Command,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: "default", // 可根据需要设置 netNSPath
	}

	// 创建容器
	resp, err := cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil,
		nil,
		container.Name,
	)
	if err != nil {
		return fmt.Errorf("无法创建容器 %s: %v", container.Name, err)
	}

	log.Printf("容器 %s 创建成功，ID: %s", container.Name, resp.ID)
	return nil
}
