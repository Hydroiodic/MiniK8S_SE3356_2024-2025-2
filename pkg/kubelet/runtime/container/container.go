package container

import (
	"context"
	"fmt"
	"log"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/image"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerServiceInterface interface {
	// Returns Container ID
	CreateContainer(
		container object.Container,
		hostConfig *container.HostConfig,
	) (string, error)
	ForceCreateContainer(
		container object.Container,
		hostConfig *container.HostConfig,
	) (string, error)
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	GetContainerInfo(containerID string) (*container.InspectResponse, error)
	GetContainerStatus(containerID string) (string, error)
}

type ContainerService struct {
	client      *client.Client
	img_service *image.ImageService
}

func NewContainerService() (*ContainerService, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	// 创建 ImageService 实例
	imgService, err := image.NewImageService()
	if err != nil {
		return nil, fmt.Errorf("无法创建镜像服务: %v", err)
	}

	return &ContainerService{client: cli, img_service: imgService}, nil
}

func (cs *ContainerService) CreateContainer(
	ctr object.Container, hostConfig *container.HostConfig,
) (string, error) {
	ctx := context.Background()

	// 拉取镜像
	err := cs.img_service.PullImage(ctr.Image)
	if err != nil {
		return "", fmt.Errorf("无法拉取镜像 %s: %v", ctr.Image, err)
	}

	// TODO: Generate Docker Configs using Container Object

	/**
	 * config *container.Config,
	 * hostConfig *container.HostConfig,
	 * networkingConfig *network.NetworkingConfig
	 */
	// 配置容器
	config := &container.Config{
		Image: ctr.Image,
		Cmd:   ctr.Command,
	}

	if hostConfig == nil {
		hostConfig = &container.HostConfig{
			NetworkMode: "default", // 可根据需要设置 netNSPath
		}
	}

	// TODO: 设置网络配置

	// 创建容器
	resp, err := cs.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil,
		nil,
		ctr.Name,
	)
	if err != nil {
		return "", fmt.Errorf("无法创建容器 %s: %v", ctr.Name, err)
	}

	log.Printf("容器 %s 创建成功，ID: %s", ctr.Name, resp.ID)

	return resp.ID, nil
}

func (cs *ContainerService) ForceCreateContainer(
	ctr object.Container,
	hostConfig *container.HostConfig,
) (string, error) {
	ctx := context.Background()

	// 检查容器是否已存在
	_, err := cs.client.ContainerInspect(ctx, ctr.Name)
	if err == nil {
		// 容器存在，删除
		_ = cs.DeleteContainer(ctr.Name)
		log.Printf("容器 %s 已存在，已删除", ctr.Name)
	}

	id, err := cs.CreateContainer(ctr, hostConfig)

	return id, err
}

func (cs *ContainerService) StartContainer(containerID string) error {
	ctx := context.Background()
	err := cs.client.ContainerStart(
		ctx,
		containerID,
		container.StartOptions{},
	)

	if err != nil {
		return fmt.Errorf("无法启动容器 %s: %v", containerID, err)
	}

	log.Printf("容器 %s 启动成功", containerID)

	return nil
}

func (cs *ContainerService) StopContainer(containerID string) error {
	ctx := context.Background()
	err := cs.client.ContainerStop(
		ctx,
		containerID,
		container.StopOptions{},
	)

	if err != nil {
		return fmt.Errorf("无法停止容器 %s: %v", containerID, err)
	}

	log.Printf("容器 %s 停止成功", containerID)

	return nil
}

func (cs *ContainerService) DeleteContainer(containerID string) error {
	ctx := context.Background()

	// 首先停止容器
	err := cs.StopContainer(containerID)
	if err != nil {
		return fmt.Errorf("无法停止容器 %s: %v", containerID, err)
	}

	// 删除容器
	err = cs.client.ContainerRemove(
		ctx,
		containerID,
		container.RemoveOptions{
			Force: true,
		},
	)

	if err != nil {
		return fmt.Errorf("无法删除容器 %s: %v", containerID, err)
	}

	log.Printf("容器 %s 删除成功", containerID)

	return nil
}

func (cs *ContainerService) GetContainerInfo(
	containerID string,
) (*container.InspectResponse, error) {
	ctx := context.Background()
	ctrInfo, err := cs.client.ContainerInspect(ctx, containerID)

	if err != nil {
		return nil, fmt.Errorf("无法获取容器 %s 信息: %v", containerID, err)
	}

	return &ctrInfo, nil
}

func (cs *ContainerService) GetContainerStatus(
	containerID string,
) (string, error) {
	ctx := context.Background()
	ctrInfo, err := cs.client.ContainerInspect(ctx, containerID)

	if err != nil {
		return "", fmt.Errorf("无法获取容器 %s 状态: %v", containerID, err)
	}

	// TODO: Use our own type
	return ctrInfo.State.Status, nil
}
