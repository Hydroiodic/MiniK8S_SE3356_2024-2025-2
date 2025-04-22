// Kubelet持有一个全局唯一的Containerd Client实例
// 一个Pod包含多个Container
// Pod内的所有容器共享以下资源：
// 1. 网络命名空间
// 2. 储存卷
// 资源的共享通过Pause容器实现

package kubelet

import (
	"context"
	"fmt"
	"log"
	"strings"
	"syscall"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl/v2"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func formatContainerName(podName, containerName string) string {
	return fmt.Sprintf("%s-%s", podName, containerName)
}

func formatSnapshotName(containerName string) string {
	return fmt.Sprintf("snapshot-%s", containerName)
}

func formatPauseContainerName(podName string) string {
	return fmt.Sprintf("%s-pause", podName)
}

/*
 * 按照Pod的规格创建Pause容器和业务容器
 * Pause容器用于提供网络命名空间
 */
func CreatePod(
	ctx context.Context,
	client *containerd.Client,
	pod *object.Pod,
) error {
	// 设置 Containerd namespace（对应 Pod 的 namespace）
	ctx = namespaces.WithNamespace(ctx, pod.Metadata.Namespace)

	// 创建Pause容器
	err := createPauseContainer(ctx, client, pod.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to create pause container: %v", err)
	}

	// 获取Pause容器的网络命名空间路径
	pauseContainer, err := client.LoadContainer(
		ctx,
		formatPauseContainerName(pod.Metadata.Name),
	)

	if err != nil {
		return fmt.Errorf(
			"failed to load pause container: %v",
			err,
		)
	}

	pauseTask, err := pauseContainer.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to load pause task: %v", err)
	}

	// Get Proceess ID of the pause container
	pausePid := pauseTask.Pid()
	netNSPath := fmt.Sprintf("/proc/%d/ns/net", pausePid)

	// 让业务容器加入Pause容器的网络命名空间
	// 创建所有容器
	for _, container := range pod.Spec.Containers {
		err := CreateContainer(
			ctx,
			client,
			pod.Metadata.Name,
			container,
			netNSPath,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to create container %s: %v",
				container.Name,
				err,
			)
		}
	}

	return nil
}

func createPauseContainer(
	ctx context.Context,
	client *containerd.Client,
	podName string) error {
	// 1. 拉取Pause镜像
	image, err := client.Pull(ctx, PauseImage, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", PauseImage, err)
	}

	// 2. 创建Pause容器
	container, err := client.NewContainer(
		ctx,
		formatContainerName(podName, "pause"),
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			// 反正一个Pod只有一个Pause容器
			formatSnapshotName("pause"), image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			// 共享网络NETWORKNAMESPACE？隔离网络！？
			oci.WithHostNamespace(specs.PIDNamespace)),
	)

	if err != nil {
		return fmt.Errorf("failed to create pause container: %v", err)
	}

	// 3. 启动Pause容器
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task for pause container: %v", err)
	}

	err = task.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start pause container: %v", err)
	}

	// TODO: Configure CNI on the pause container?

	return nil
}

func CreateContainer(
	ctx context.Context,
	client *containerd.Client,
	podName string, // Pod名称
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
		formatContainerName(podName, containerSpec.Name), // 容器名称
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			fmt.Sprintf("snapshot-%s", containerSpec.Name),
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
			oci.WithEnv(
				[]string{
					"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				},
			),
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

// StopContainer 停止指定容器
func StopContainer(
	ctx context.Context,
	client *containerd.Client,
	podName string,
	containerName string,
) error {
	// 1. 获取容器
	containerID := formatContainerName(podName, containerName)
	container, err := client.LoadContainer(ctx, containerID)

	if err != nil {
		return fmt.Errorf("failed to load container %s: %v", containerID, err)
	}

	// 2. 获取任务
	task, err := container.Task(ctx, nil)
	if err != nil {
		if strings.Contains(err.Error(), "no running task") {
			return fmt.Errorf("container %s is not running", containerID)
		}

		return fmt.Errorf(
			"failed to load task for container %s: %v",
			containerID,
			err,
		)
	}

	// 3. 发送 SIGTERM 信号，优雅停止
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf(
			"failed to send SIGTERM to task %s: %v",
			containerID,
			err,
		)
	}

	// 4. 等待任务退出
	_, err = task.Wait(ctx)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to wait for task %s: %v", containerID, err)
	}

	// 5. 删除任务（清理运行时资源）
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil &&
		!strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to delete task %s: %v", containerID, err)
	}

	return nil
}

func DeleteContainer(
	ctx context.Context,
	client *containerd.Client,
	podName string,
	containerSpec object.Container,
) error {
	// 1. 获取容器
	containerID := formatContainerName(podName, containerSpec.Name)
	container, err := client.LoadContainer(ctx, containerID)

	if err != nil {
		return fmt.Errorf("failed to load container %s: %v", containerID, err)
	}

	// 2. 获取任务
	task, err := container.Task(ctx, nil)

	if err != nil {
		return fmt.Errorf(
			"failed to load task for container %s: %v",
			containerID,
			err,
		)
	}

	// 3. 停止并删除任务
	// First, attempt to stop the task gracefully
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf(
			"failed to send SIGTERM to task %s: %v",
			containerID,
			err,
		)
	}

	// Wait for task to exit
	_, err = task.Wait(ctx)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to wait for task %s: %v", containerID, err)
	}

	// Delete task with force kill if necessary
	if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil &&
		!strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("failed to delete task %s: %v", containerID, err)
	}

	// TODO: 删除快照？
	// snapshotName := formatSnapshotName(containerSpec.Name)
	// if err := client.SnapshotService("overlayfs").Remove(ctx, snapshotName); err != nil {
	// 	return fmt.Errorf("failed to remove snapshot %s: %v", snapshotName, err)
	// }

	// 4. 删除容器
	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete container %s: %v", containerID, err)
	}

	return nil
}

/*
	监控容器事件

* 事件类型：
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
