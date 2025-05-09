package pod

import (
	"log"

	ctr_runtime "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
)

type PodService struct {
	ctr_service *ctr_runtime.ContainerService
}

func (p *PodService) CreatePod(pod *object.Pod) error {
	// Create Pause Container
	pauseId, err := CreatePauseContainer(p.ctr_service, pod)
	if err != nil {
		log.Printf("Failed to create pause container: %v", err)
	}

	// Create Pod Containers
	pauseNsArg := "container:" + pauseId

	// 所有容器共享相同的 IP 地址（Pod IP）。
	// 容器间可以通过 localhost 通信。
	// 容器共享进程视图和 IPC 资源。
	for _, ctrConfig := range pod.Spec.Containers {
		// 无需端口映射
		ctr := object.Container{
			Name:    ctrConfig.Name,
			Image:   ctrConfig.Image,
			Command: ctrConfig.Command,
			Args:    ctrConfig.Args,
			// Ports:   ctrConfig.Ports,
			Resources: ctrConfig.Resources,
		}

		// 普通容器在创建时会通过 Docker 的
		// --net=container:<pauseId>、 --ipc=container:<pauseId> 和 --pid=container:<pauseId> 选项，
		// 加入 Pause Container 的命名空间
		hostConfig := &container.HostConfig{
			NetworkMode: container.NetworkMode(pauseNsArg),
			IpcMode:     container.IpcMode(pauseNsArg),
			PidMode:     container.PidMode(pauseNsArg),
		}

		// 创建容器
		ctrId, err := p.ctr_service.CreateContainer(ctr, hostConfig)
		if err != nil {
			log.Printf("Failed to create container %s: %v", ctr.Name, err)
		}

		log.Printf("Created container %s with ID %s", ctr.Name, ctrId)
	}

	return nil
}
