package kubelet

import (
	"context"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/errdefs"
	"github.com/containerd/typeurl/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func formatSnapshotName(containerName string) string {
	return fmt.Sprintf("snapshot-%s", containerName)
}

// 拉取镜像，如果已经存在则不拉取
// 只许成功，不许失败！
func pullImage(
	ctx context.Context,
	client *containerd.Client,
	imageName string,
) (containerd.Image, error) {
	image, err := client.GetImage(ctx, imageName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// 镜像不存在，拉取镜像
			image, err = client.Pull(
				ctx,
				imageName,
				containerd.WithPullUnpack,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to pull image %s: %v",
					imageName,
					err,
				)
			}
		} else {
			return nil, fmt.Errorf("failed to check image %s: %v", imageName, err)
		}
	}
	return image, nil
}

// 如果快照已经存在，则删除快照
func deleteSnapShotIfExists(
	ctx context.Context,
	client *containerd.Client,
	snapshotName string,
) error {
	// 获取快照服务
	snapshotter := client.SnapshotService("overlayfs")
	log.Printf("检查快照 %s 是否存在", snapshotName)

	// 检查快照是否已存在
	_, err := snapshotter.Stat(ctx, snapshotName)
	if err == nil {
		// 快照已存在，尝试删除
		log.Printf("快照 %s 已存在，正在删除", snapshotName)
		if err := snapshotter.Remove(ctx, snapshotName); err != nil {
			return fmt.Errorf("删除已有快照 %s 失败: %v", snapshotName, err)
		}
	} else if !errdefs.IsNotFound(err) {
		// 如果不是 "not found" 错误，返回错误
		return fmt.Errorf("检查快照 %s 失败: %v", snapshotName, err)
	}
	return nil
}

/**
 * 创建容器
 * 1. 保证在SnapShot已经存在的情况下可以处理
 */
func CreateContainer(
	ctx context.Context,
	client *containerd.Client,
	containerSpec object.Container, // 容器规格
	netNSPath string, // 网络命名空间路径
) error {
	// 1. 拉取镜像
	image, err := pullImage(
		ctx,
		client,
		containerSpec.Image,
	)

	if err != nil {
		// 不应该发生！除非网络错误。
		return fmt.Errorf(
			"failed to pull image %s: %v",
			containerSpec.Image,
			err,
		)
	}

	// 2. 删除已有的快照
	snapshotName := formatSnapshotName(containerSpec.Name)
	err = deleteSnapShotIfExists(
		ctx,
		client,
		snapshotName,
	)

	if err != nil {
		// 不应该发生！除非网络错误。
		return fmt.Errorf(
			"[FATAL]: failed to delete snapshot %s: %v",
			snapshotName,
			err,
		)
	}

	// 检查容器是否已经存在
	_, err = client.LoadContainer(ctx, containerSpec.Name)
	if err == nil {
		// 容器已存在，删除容器
		if err := DeleteContainer(ctx, client, containerSpec); err != nil {
			return fmt.Errorf(
				"failed to delete existing container %s: %v",
				containerSpec.Name,
				err,
			)
		}
	}

	opts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}

	if netNSPath == "" {
		opts = append(
			opts,
			oci.WithHostNamespace(specs.PIDNamespace),
		) // 可以共享PID命名空间
	} else {
		opts = append(opts, oci.WithLinuxNamespace(
			specs.LinuxNamespace{
				Type: specs.NetworkNamespace,
				Path: netNSPath,
			},
		)) // 加入Pause容器的网络命名空间

		opts = append(opts,
			oci.WithProcessArgs(containerSpec.Command...), // 设置命令和参数
		)
	}

	// 2. 创建容器
	container, err := client.NewContainer(
		ctx,
		containerSpec.Name, // 容器名称
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			formatSnapshotName(containerSpec.Name), image,
		),
		containerd.WithNewSpec(
			opts...,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	log.Printf(
		"容器 %s 创建成功，快照名称: %s, 命名空间: %s",
		containerSpec.Name,
		formatSnapshotName(containerSpec.Name),
		func() string {
			ns, _ := namespaces.Namespace(ctx)
			return ns
		}(),
	)

	// 3. 启动容器
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	err = task.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start task: %v", err)
	}

	return nil
}

/**
 * 停止容器
 * ctx: 上下文。定义了Containerd的命名空间
 */
func StopContainer(
	ctx context.Context,
	client *containerd.Client,
	containerName string,
) error {
	// 1. 获取容器
	container, err := client.LoadContainer(ctx, containerName)

	if err != nil {
		return fmt.Errorf("failed to load container %s: %v", containerName, err)
	}

	// 2. 获取任务
	task, err := container.Task(ctx, nil)
	if err != nil {
		if strings.Contains(err.Error(), "no running task") {
			// 任务不存在，容器已经停止
			return nil
		}

		return fmt.Errorf(
			"failed to load task for container %s: %v",
			containerName,
			err,
		)
	}

	// 3. 发送 SIGTERM 信号，优雅停止
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf(
			"failed to send SIGTERM to task %s: %v",
			containerName,
			err,
		)
	}

	// 4. 等待任务退出
	_, err = task.Wait(ctx)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to wait for task %s: %v", containerName, err)
	}

	// 5. 删除任务（清理运行时资源）
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil &&
		!strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to delete task %s: %v", containerName, err)
	}

	return nil
}

func DeleteContainer(
	ctx context.Context,
	client *containerd.Client,
	containerSpec object.Container,
) error {
	var containerName = containerSpec.Name
	// 1. 获取容器
	container, err := client.LoadContainer(ctx, containerName)

	if err != nil {
		// 容器已经不存在
		return nil
	}

	// 2. 停止容器
	if err := StopContainer(ctx, client, containerName); err != nil {
		return fmt.Errorf(
			"failed to stop container %s: %v",
			containerName,
			err,
		)
	}

	// 2. 删除容器
	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf(
			"failed to delete container %s: %v",
			containerName,
			err,
		)
	}

	// 3. 删除快照
	snapshotName := formatSnapshotName(containerSpec.Name)
	_ = client.SnapshotService("overlayfs").Remove(ctx, snapshotName)

	return nil
}

/*
		监控容器事件
	  - 容器创建：/containers/create
	  - 快照准备：/snapshots/prepare
	  - 任务创建和启动：/tasks/create, /tasks/start
	  - 任务退出：/tasks/exit（进程退出）
	  - 任务删除：/tasks/delete（清理任务）
	  - 快照清理：/snapshots/remove（清理文件系统）
	  - 容器删除：/containers/delete（清理元数据）
*/
func MonitorContainerEvents(
	ctx context.Context,
	client *containerd.Client,
	namespace string,
) error {
	// 设置命名空间
	ctx = namespaces.WithNamespace(ctx, namespace)

	// 获取事件服务
	eventService := client.EventService()

	// 创建事件订阅
	eventsCh, errCh := eventService.Subscribe(ctx)

	// 监听事件
	go func() {
		for {
			select {
			case event := <-eventsCh:
				// 解码事件
				// namespace, topic, event
				any, err := typeurl.UnmarshalAny(event.Event)
				if err != nil {
					log.Printf("Failed to unmarshal event: %v", err)
					continue
				}

				// 打印事件类型和详情
				fmt.Printf("Event received: %T - %+v\n", any, any)

			case err := <-errCh:
				if err != nil {
					log.Printf("Error receiving events: %v", err)
					return
				}
			}
		}
	}()

	return nil
}

// TODO: 返回特定容器信息
// ContainerInfo 结构体用于存储容器信息
type ContainerInfo struct {
	ID    string // 容器ID
	Name  string // 容器名称
	Image string // 镜像名称
	// Command   []string          // 启动命令
	Status    string            // 运行状态 (e.g., "running", "stopped")
	CreatedAt time.Time         // 创建时间
	Labels    map[string]string // 容器标签
}

// GetContainerInfo 获取特定容器的信息
// 命名空间应预先设置
// ctx = namespaces.WithNamespace(ctx, podNameSpace) // 例如 k8s.io/podName 命名空间
func GetContainerInfo(
	ctx context.Context,
	client *containerd.Client,
	containerName string,
) (*ContainerInfo, error) {
	// 加载容器
	container, err := client.LoadContainer(ctx, containerName)
	if err != nil {
		return nil, fmt.Errorf("无法加载容器 %s: %v", containerName, err)
	}

	// 获取容器信息
	info, err := container.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("无法获取容器信息 %s: %v", containerName, err)
	}

	// 获取镜像名称
	image, err := container.Image(ctx)
	if err != nil {
		return nil, fmt.Errorf("无法获取镜像信息 %s: %v", containerName, err)
	}

	// 获取任务状态
	task, err := container.Task(ctx, nil)

	var status string

	if err != nil {
		if strings.Contains(err.Error(), "no running task") {
			status = string(containerd.Stopped)
		} else {
			return nil, fmt.Errorf("无法加载任务 %s: %v", containerName, err)
		}
	} else {
		taskStatus, err := task.Status(ctx)
		if err != nil {
			return nil, fmt.Errorf("无法获取任务状态 %s: %v", containerName, err)
		}

		status = string(taskStatus.Status) // e.g., "running", "paused", "stopped"
	}

	// 构造返回信息
	containerInfo := &ContainerInfo{
		// ID:    containerID,
		Name:  containerName,
		Image: image.Name(),
		// Command:   info.Spec.Value.Process.Args, // 获取命令
		Status:    status,
		CreatedAt: info.CreatedAt,
		Labels:    info.Labels,
	}

	return containerInfo, nil
}
