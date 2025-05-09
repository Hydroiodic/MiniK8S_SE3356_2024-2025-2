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

type ContainerService interface {
	CreateContainer(container object.Container) error
	StartContainer(containerID string) error
	StopContainer(containerID string) error
	DeleteContainer(containerID string) error
	GetContainerInfo(containerID string) (*container.InspectResponse, error)
	GetContainerStatus(containerID string) (string, error)
}

type containerService struct {
	client      *client.Client
	img_service *image.ImageService
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

func (cs *containerService) CreateContainer(ctr object.Container) error {
	ctx := context.Background()

	// 拉取镜像
	err = cs.img_service.PullImage(ctr.Image)
	if err != nil {
		return fmt.Errorf("无法拉取镜像 %s: %v", ctr.Image, err)
	}

	// TODO: 删除已有镜像？
	// // 检查容器是否已存在
	// if _, err := cli.ContainerInspect(ctx, container.Name); err == nil {
	// 	// 容器存在，删除
	// 	if err := cs.DeleteContainer(container.Name); err != nil {
	// 		return fmt.Errorf("无法删除现有容器 %s: %v", container.Name, err)
	// 	}
	// }

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

	hostConfig := &container.HostConfig{
		NetworkMode: "default", // 可根据需要设置 netNSPath
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
		return fmt.Errorf("无法创建容器 %s: %v", container.Name, err)
	}

	log.Printf("容器 %s 创建成功，ID: %s", container.Name, resp.ID)
	return nil
}

func (cs *containerService) StartContainer(containerID string) error {
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

func (cs *containerService) StopContainer(containerID string) error {
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

func (cs *containerService) DeleteContainer(containerID string) error {
	ctx := context.Background()
	err := cs.client.ContainerRemove(
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

func (cs *containerService) GetContainerInfo(
	containerID string,
) (*container.InspectResponse, error) {
	ctx := context.Background()
	ctrInfo, err := cs.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器 %s 信息: %v", containerID, err)
	}
	return &ctrInfo, nil
}

func (cs *containerService) GetContainerStatus(
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
