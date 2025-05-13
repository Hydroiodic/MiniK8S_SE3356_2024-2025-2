package pod

import (
	runtime "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime"
	ctr_runtime "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
)

// 创建Pause容器
func CreatePauseContainer(
	c ctr_runtime.ContainerServiceInterface,
	pod *object.Pod,
) (string, error) {
	ctr := object.Container{
		Name: utils.FormatContainerName(
			pod.Metadata.Namespace,
			pod.Metadata.Name,
			"pause",
		),
		Image: runtime.PauseImage,
	}

	// Append Port
	for _, ctrConfig := range pod.Spec.Containers {
		ctr.Ports = append(ctr.Ports, ctrConfig.Ports...)
	}

	// Append Labels
	ctr.Labels = utils.NewLabelForPauseContainer(
		pod.Metadata.Namespace,
		pod.Metadata.Name,
		pod.Metadata.Labels,
	)

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
