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

	containerd_manager "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/containerd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func formatPauseContainerName(podName string) string {
	return fmt.Sprintf("%s-pause", podName)
}

func createPauseContainer(
	ctx context.Context,
	client *containerd.Client) error {
	// 1. 拉取Pause镜像
	image, err := client.Pull(ctx, PauseImage, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", PauseImage, err)
	}

	// 2. 创建Pause容器
	container, err := client.NewContainer(
		ctx,
		"pause",
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			// 反正一个Pod只有一个Pause容器
			formatSnapshotName("pause"), image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithHostNamespace(specs.PIDNamespace)), // 可以共享PID命名空间？
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

// Kubelet Monitors all the pods on the node
type KubeletInstance struct {
	cli *containerd.Client
	// TODO: Do I need to hold some pod or just ask API Server
}

func NewKubeletInstance() *KubeletInstance {
	cli, err := containerd_manager.NewContainerdClient()

	if err != nil {
		return nil
	}

	return &KubeletInstance{
		cli: cli,
	}
}

/*
 * 按照Pod的规格创建Pause容器和业务容器
 * Pause容器用于提供网络命名空间
 */
func (inst *KubeletInstance) CreatePod(
	ctx context.Context,
	pod *object.Pod,
) error {
	// 设置 Containerd namespace（对应 Pod 的 namespace）
	ctx = namespaces.WithNamespace(ctx, pod.Metadata.Namespace)
	client := inst.cli

	// 创建Pause容器
	err := createPauseContainer(ctx, client)
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

// TODO: 获取本Node下的Pod状态

// TODO：获取特定的Pod状态
