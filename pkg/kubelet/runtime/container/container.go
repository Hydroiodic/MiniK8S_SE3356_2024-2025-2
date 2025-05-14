package container

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/image"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	ExecCommand(containerID string, cmd []string) (string, error)
	DeleteContainer(containerID string) error

	GetContainerInfo(containerID string) (*container.InspectResponse, error)
	// String representation of the container state.
	// Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
	GetContainerStatus(containerID string) (string, error)

	GetContainerIdByName(name string) (string, error)
	GetContainerNameById(id string) (string, error)

	ListContainerIds() ([]string, error)

	GetContainersByLabels(
		labels map[string]string,
	) ([]object.Container, error)

	GetContainerInspectsByLabels(
		labels map[string]string,
	) ([]*container.InspectResponse, error)
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

	repo, img, version := utils.ChunkImageName(ctr.Image)

	if repo == "docker.io/library" {
		ctr.Image = img + ":" + version
	} else {
		ctr.Image = fmt.Sprintf("%s/%s:%s", repo, img, version)
	}

	// 配置容器
	config := &container.Config{
		Image:  ctr.Image,
		Cmd:    ctr.Command,
		Labels: ctr.Labels,
	}

	log.Printf("Labels: %v", ctr.Labels)

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

// ExecCommand 在指定容器中执行命令并返回输出
func (cs *ContainerService) ExecCommand(
	containerID string,
	cmd []string,
) (string, error) {
	ctx := context.Background()

	// 创建 Exec 配置
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,  // 捕获标准输出
		AttachStderr: true,  // 捕获标准错误
		Tty:          false, // 不分配伪终端
	}

	// 创建 Exec 实例
	execResp, err := cs.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("无法创建 Exec 实例: %v", err)
	}

	// 启动 Exec 实例并捕获输出
	resp, err := cs.client.ContainerExecAttach(
		ctx,
		execResp.ID,
		container.ExecStartOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("无法启动 Exec 实例: %v", err)
	}
	defer resp.Close()

	// 捕获标准输出和标准错误
	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, resp.Reader)

	if err != nil {
		return "", fmt.Errorf("无法读取 Exec 输出: %v", err)
	}

	// 检查 Exec 命令的退出状态
	inspectResp, err := cs.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return "", fmt.Errorf("无法检查 Exec 状态: %v", err)
	}

	// 如果退出码非 0，返回错误并包含 stderr
	if inspectResp.ExitCode != 0 {
		return "", fmt.Errorf(
			"命令执行失败，退出码 %d: %s",
			inspectResp.ExitCode,
			stderr.String(),
		)
	}

	// 返回标准输出，移除多余的换行符
	return strings.TrimSpace(stdout.String()), nil
}

func (cs *ContainerService) GetContainerIdByName(name string) (string, error) {
	ctx := context.Background()

	// 获取所有容器
	containers, err := cs.client.ContainerList(
		ctx,
		container.ListOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("无法获取容器列表: %v", err)
	}

	// 遍历容器，查找匹配的名称
	for _, container := range containers {
		// > $ docker inspect goofy_lalande | grep goofy
		// "Name": "/goofy_lalande",
		if container.Names[0] == "/"+name {
			return container.ID, nil
		}
	}

	return "", fmt.Errorf("未找到名为 %s 的容器", name)
}

func (cs *ContainerService) GetContainerNameById(id string) (string, error) {
	ctx := context.Background()
	// 获取容器信息
	ctrInfo, err := cs.client.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("无法获取容器 %s 信息: %v", id, err)
	}
	// 提取容器名称
	name := strings.TrimPrefix(ctrInfo.Name, "/")
	// 返回容器名称
	return name, nil
}

func (cs *ContainerService) ListContainerIds() ([]string, error) {
	ctx := context.Background()

	// 获取所有容器
	containers, err := cs.client.ContainerList(
		ctx,
		container.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器列表: %v", err)
	}

	// 提取容器 ID
	var containerIDs []string
	for _, container := range containers {
		containerIDs = append(containerIDs, container.ID)
	}

	return containerIDs, nil
}

func (cs *ContainerService) GetContainersByLabels(
	labels map[string]string,
) ([]object.Container, error) {
	ctx := context.Background()

	// 获取所有容器
	containers, err := cs.client.ContainerList(
		ctx,
		container.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器列表: %v", err)
	}

	var result []object.Container

	for _, container := range containers {
		// 获取容器的标签
		ctrInfo, err := cs.client.ContainerInspect(ctx, container.ID)
		if err != nil {
			return nil, fmt.Errorf("无法获取容器 %s 信息: %v", container.ID, err)
		}

		log.Printf(
			"容器 %s 的标签: %v",
			container.ID,
			ctrInfo.Config.Labels,
		)

		// 检查标签是否匹配
		matches := true

		for key, value := range labels {
			if ctrInfo.Config.Labels[key] != value {
				log.Printf(
					"标签不匹配: %s=%s, 实际为: %s",
					key,
					value,
					ctrInfo.Config.Labels[key],
				)

				matches = false

				break
			}
		}

		if matches {
			result = append(result, object.Container{
				ID:     container.ID,
				Name:   strings.TrimPrefix(container.Names[0], "/"),
				Image:  ctrInfo.Config.Image,
				Labels: ctrInfo.Config.Labels,
			})
		}
	}

	return result, nil
}

func (cs *ContainerService) GetContainerInspectsByLabels(
	labels map[string]string,
) ([]*container.InspectResponse, error) {
	ctx := context.Background()

	// 获取所有容器
	containers, err := cs.client.ContainerList(
		ctx,
		container.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器列表: %v", err)
	}

	var result []*container.InspectResponse

	for _, container := range containers {
		// 获取容器的标签
		ctrInfo, err := cs.client.ContainerInspect(ctx, container.ID)
		if err != nil {
			return nil, fmt.Errorf("无法获取容器 %s 信息: %v", container.ID, err)
		}

		// 检查标签是否匹配
		matches := true

		for key, value := range labels {
			if ctrInfo.Config.Labels[key] != value {
				matches = false
				break
			}
		}

		if matches {
			result = append(result, &ctrInfo)
		}
	}

	return result, nil
}
