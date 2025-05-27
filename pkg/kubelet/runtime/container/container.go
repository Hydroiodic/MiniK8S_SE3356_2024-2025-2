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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

const cadvisorContainerName = "cadvisor"

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

	GetContainerInfo(containerID string) (container.InspectResponse, error)
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
		Image:        ctr.Image,
		Cmd:          ctr.Command,
		Labels:       ctr.Labels,
		ExposedPorts: ctr.ExposedPorts,
		User: combineUserAndGroup(
			ctr.SecurityContexts.RunAsUser,
			ctr.SecurityContexts.RunAsGroup,
		),
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
) (container.InspectResponse, error) {
	ctx := context.Background()
	ctrInfo, err := cs.client.ContainerInspect(ctx, containerID)

	if err != nil {
		return ctrInfo, fmt.Errorf("无法获取容器 %s 信息: %v", containerID, err)
	}

	return ctrInfo, nil
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

func (cs *ContainerService) GetContainerByName(
	containerName string,
) (*container.Summary, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", containerName)

	allContainers, err := cs.client.ContainerList(
		context.Background(),
		container.ListOptions{
			All:     true,
			Filters: filterArgs,
		},
	)
	if err != nil {
		return nil, err
	}

	for _, container := range allContainers {
		for _, containerName := range container.Names {
			if containerName == "/"+containerName {
				return &container, nil
			}
		}
	}

	return nil, fmt.Errorf("container %s not found", containerName)
}

func (cs *ContainerService) RunCadvisorContainer() (string, error) {
	cadvisorContainer, _ := cs.GetContainerByName(cadvisorContainerName)
	if cadvisorContainer == nil {
		//创建容器
		// 配置 cAdvisor 容器
		exposedPorts := nat.PortSet{
			"8080/tcp": struct{}{},
		}

		containerSpec := object.Container{
			Name:         "cadvisor",
			Image:        "gcr.nju.edu.cn/cadvisor/cadvisor:v0.49.1",
			ExposedPorts: exposedPorts,
		}

		portBindings := nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostPort: "8090", // 将容器 8080 映射到主机 8090
				},
			},
		}

		hostConfig := &container.HostConfig{
			Binds: []string{
				"/:/rootfs:ro",
				"/var/run:/var/run:ro",
				"/sys:/sys:ro",
				"/var/lib/docker/:/var/lib/docker:ro",
				"/dev/disk/:/dev/disk:ro",
				"/var/run/docker.sock:/var/run/docker.sock", // Docker Socket 必须挂载
			},
			PortBindings: portBindings,
			Privileged:   true, // 必须开启特权模式
			RestartPolicy: container.RestartPolicy{
				Name: "always", // 自动重启
			},
			Mounts: []mount.Mount{
				{
					Source: "/dev/kmsg",
					Target: "/dev/kmsg",
					Type:   mount.TypeBind,
				},
			},
		}
		containerID, _ := cs.CreateContainer(containerSpec, hostConfig)
		err := cs.StartContainer(containerID)

		return containerID, err
	} else {
		fmt.Println("the cadevisor container has already been created ")

		return "0", nil
	}
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
		container.ListOptions{
			All: true,
		},
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
		container.ListOptions{
			All: true,
		},
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
		container.ListOptions{
			All: true,
		},
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

		// 检查标签是否匹配
		matches := true

		for key, value := range labels {
			if ctrInfo.Config.Labels[key] != value {
				matches = false
				break
			}
		}

		// TODO: 为了兼容性，这里特事特办，解析一下名称
		containerName := strings.TrimPrefix(container.Names[0], "/")
		_, _, containerName = utils.ParseContainerName(containerName)

		// ExposedPorts 怎么转换到 Ports 列表
		ports := make([]int, 0)

		for port := range ctrInfo.Config.ExposedPorts {
			ports = append(ports, port.Int())
		}

		log.Printf("Container %s Exposed Ports: %v", containerName, ports)

		ipAddr := ""

		if len(ctrInfo.NetworkSettings.Networks) > 0 {
			for _, network := range ctrInfo.NetworkSettings.Networks {
				ipAddr = network.IPAddress
				break
			}
		}

		if matches {
			result = append(result, object.Container{
				ID:      container.ID,
				Name:    containerName,
				Image:   ctrInfo.Config.Image,
				Command: ctrInfo.Config.Cmd,
				// TODO: 重建Limits？
				Ports:  ports,
				Labels: ctrInfo.Config.Labels,
				// TODO: Flannel
				IP: ipAddr,
			})
		}
	}

	return result, nil
}

func (cs *ContainerService) GetContainerInspectsByLabels(
	labels map[string]string,
) ([]*container.InspectResponse, error) {
	ctx := context.Background()

	// 构建 label 过滤器
	filterArgs := filters.NewArgs()
	for key, value := range labels {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", key, value))
	}

	// 获取所有匹配标签的容器
	containers, err := cs.client.ContainerList(
		ctx,
		container.ListOptions{
			All:     true,
			Filters: filterArgs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器列表: %v", err)
	}

	var result []*container.InspectResponse

	for _, container := range containers {
		ctrInfo, err := cs.client.ContainerInspect(ctx, container.ID)
		if err != nil {
			return nil, fmt.Errorf("无法获取容器 %s 信息: %v", container.ID, err)
		}

		result = append(result, &ctrInfo)
	}

	return result, nil
}
