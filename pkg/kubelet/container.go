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
	"github.com/containerd/typeurl/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func formatSnapshotName(containerName string) string {
	return fmt.Sprintf("snapshot-%s", containerName)
}

func CreateContainer(
	ctx context.Context,
	client *containerd.Client,
	containerSpec object.Container, // 容器规格
	netNSPath string, // 网络命名空间路径
) error {
	// 1. 拉取镜像
	image, err := client.Pull(
		ctx,
		containerSpec.Image,
		containerd.WithPullUnpack,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to pull image %s: %v",
			containerSpec.Image,
			err,
		)
	}

	// 2. 创建容器
	container, err := client.NewContainer(
		ctx,
		containerSpec.Name, // 容器名称
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			formatSnapshotName(containerSpec.Name),
			image,
		),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithLinuxNamespace(
				specs.LinuxNamespace{
					Type: specs.NetworkNamespace,
					Path: netNSPath,
				},
			), // 加入Pause容器的网络命名空间
			// oci.WithEnv(
			// 	[]string{
			// 		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			// 	},
			// ),
			oci.WithProcessArgs(containerSpec.Command...), // 设置命令和参数
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

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
			return fmt.Errorf("container %s is not running", containerName)
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
	var containerName string = containerSpec.Name
	// 1. 获取容器
	container, err := client.LoadContainer(ctx, containerName)

	if err != nil {
		return fmt.Errorf("failed to load container %s: %v", containerName, err)
	}

	// 2. 获取任务
	task, err := container.Task(ctx, nil)

	if err != nil {
		return fmt.Errorf(
			"failed to load task for container %s: %v",
			containerName,
			err,
		)
	}

	// 3. 停止并删除任务
	// First, attempt to stop the task gracefully
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf(
			"failed to send SIGTERM to task %s: %v",
			containerName,
			err,
		)
	}

	// Wait for task to exit
	_, err = task.Wait(ctx)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to wait for task %s: %v", containerName, err)
	}

	// Delete task with force kill if necessary
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil &&
		!strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to delete task %s: %v", containerName, err)
	}

	// TODO: 删除快照？
	// snapshotName := formatSnapshotName(containerSpec.Name)
	// if err := client.SnapshotService("overlayfs").Remove(ctx, snapshotName); err != nil {
	// 	return fmt.Errorf("failed to remove snapshot %s: %v", snapshotName, err)
	// }

	// 4. 删除容器
	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf(
			"failed to delete container %s: %v",
			containerName,
			err,
		)
	}

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
func GetContainerInfo(
	ctx context.Context,
	client *containerd.Client,
	podNameSpace string,
	containerName string,
) (*ContainerInfo, error) {
	// 设置命名空间（如果需要）
	// TODO: 修改命名空间
	ctx = namespaces.WithNamespace(ctx, podNameSpace) // 例如 k8s.io/podName 命名空间

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
