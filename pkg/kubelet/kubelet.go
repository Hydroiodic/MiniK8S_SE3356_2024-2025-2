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

	containerd_manager "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/containerd"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

func formatPauseContainerName(podName string) string {
	return fmt.Sprintf("%s-pause", podName)
}

func formatContainerName(podName string, containerName string) string {
	return fmt.Sprintf("%s-%s", podName, containerName)
}

func createPauseContainer(
	ctx context.Context,
	client *containerd.Client,
	containerName string) error {
	err := CreateContainer(ctx, client, object.Container{
		Name:  containerName,
		Image: PauseImage,
	}, "")

	if err != nil {
		log.Printf("Failed to create pause container: %v", err)
		return fmt.Errorf("failed to create pause container: %v", err)
	}

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

func (inst *KubeletInstance) GetContainerdClient() *containerd.Client {
	return inst.cli
}

/*
 * 按照Pod的规格创建Pause容器和业务容器
 * Pause容器用于提供网络命名空间
 * 注意：Pause容器的名称是固定的，格式为 <pod-name>-pause
 * 上下文会被修改为Pod指定的NameSpace
 */
func (inst *KubeletInstance) CreatePod(
	ctx context.Context,
	pod *object.Pod,
) error {
	// 设置 Containerd namespace（对应 Pod 的 namespace）
	ctx = namespaces.WithNamespace(ctx, pod.Metadata.Namespace)
	client := inst.cli

	pauseContainerName := formatPauseContainerName(pod.Metadata.Name)

	log.Printf(
		"Creating pod %s with pause container %s",
		pod.Metadata.Name,
		pauseContainerName,
	)

	// 创建Pause容器
	err := createPauseContainer(ctx, client, pauseContainerName)
	if err != nil {
		return fmt.Errorf("failed to create pause container: %v", err)
	}

	// 获取Pause容器的网络命名空间路径
	pauseContainer, err := client.LoadContainer(
		ctx,
		pauseContainerName,
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

	log.Printf(
		"Pause container %s created with PID %d and netns %s",
		pauseContainerName,
		pausePid,
		netNSPath,
	)

	// 让业务容器加入Pause容器的网络命名空间
	// 创建所有容器
	for _, container := range pod.Spec.Containers {
		// Edit Container Name
		container.Name = formatContainerName(
			pod.Metadata.Name,
			container.Name,
		)

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
