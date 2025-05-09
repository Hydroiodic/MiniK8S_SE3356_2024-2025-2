package pod

import (
	ctr_runtime "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
)

// 创建Pause容器
func CreatePauseContainer(
	c ctr_runtime.ContainerServiceInterface,
	pod *object.Pod,
) (string, error) {
	// TODO: How to use UID?
	ctr := object.Container{
		Name:  pod.Metadata.Name + "-pause",
		Image: "k8s.gcr.io/pause:3.6",
	}

	// Append Port
	for _, ctrConfig := range pod.Spec.Containers {
		ctr.Ports = append(ctr.Ports, ctrConfig.Ports...)
	}

	// TODO: Append Label

	// 让容器共享 IPC 资源，允许使用共享内存、信号量等机制
	hostConfig := &container.HostConfig{
		IpcMode: container.IPCModeShareable,
	}

	// 创建Pause容器
	pauseId, err := c.CreateContainer(ctr, hostConfig)
	if err != nil {
		return "", err
	}

	// TODO: 分配IP

	return pauseId, nil
}
