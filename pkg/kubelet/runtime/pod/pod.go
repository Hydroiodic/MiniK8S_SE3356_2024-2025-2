package pod

import (
	"log"

	ctr_runtime "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/docker/docker/api/types/container"
)

type PodService struct {
	CtrService *ctr_runtime.ContainerService
}

func NewPodService(ctrService *ctr_runtime.ContainerService) *PodService {
	return &PodService{
		CtrService: ctrService,
	}
}

/**
 * NOTE: Pod内的数据结构会被修改
 * Container的ID会在创建后被赋值
 */
func (p *PodService) CreatePod(pod *object.Pod) error {
	// Create Pause Container
	pauseId, err := CreatePauseContainer(p.CtrService, pod)
	if err != nil {
		log.Printf("Failed to create pause container: %v", err)
		// 如果创建 Pause Container 失败，后面创建也没有意义了，返回错误！
		return err
	}

	// 将 Pause Container 的 ID 储存到 Pod
	(*pod).Spec.PauseContainerID = pauseId

	// Create Pod Containers
	pauseNsArg := "container:" + pauseId

	// 所有容器共享相同的 IP 地址（Pod IP）。
	// 容器间可以通过 localhost 通信。
	// 容器共享进程视图和 IPC 资源。
	for i, ctrConfig := range pod.Spec.Containers {
		// 无需端口映射
		ctr := object.Container{
			Name: utils.FormatContainerName(
				pod.Metadata.Namespace,
				pod.Metadata.Name,
				ctrConfig.Name,
			), // TODO: Edit the Name Here?
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
		ctrId, err := p.CtrService.CreateContainer(ctr, hostConfig)
		if err != nil {
			log.Printf("Failed to create container %s: %v", ctr.Name, err)
			return err
		}

		(*pod).Spec.Containers[i].ID = ctrId

		log.Printf("Created container %s with ID %s", ctr.Name, ctrId)
	}

	// TODO: 写入 Pod 的状态？

	return nil
}

func (p *PodService) StartPod(pod *object.Pod) error {
	// Start Pause Container
	err := p.CtrService.StartContainer(pod.Spec.PauseContainerID)
	if err != nil {
		log.Printf("Failed to start pause container: %v", err)
		return err
	}

	// Start Pod Containers
	for _, ctrConfig := range pod.Spec.Containers {
		err = p.CtrService.StartContainer(ctrConfig.ID)
		if err != nil {
			log.Printf("Failed to start container %s: %v", ctrConfig.Name, err)
			return err
		}
	}

	// TODO: 写入 Pod 的状态？

	return nil
}

func (p *PodService) StopPod(pod *object.Pod) error {
	// Stop Pod Containers
	for _, ctrConfig := range pod.Spec.Containers {
		err := p.CtrService.StopContainer(ctrConfig.ID)
		if err != nil {
			log.Printf("Failed to stop container %s: %v", ctrConfig.Name, err)
			return err
		}
	}

	// Stop Pause Container
	err := p.CtrService.StopContainer(pod.Spec.PauseContainerID)
	if err != nil {
		log.Printf("Failed to stop pause container: %v", err)
		return err
	}

	// TODO: 写入 Pod 的状态？

	return nil
}

func (p *PodService) DeletePod(pod *object.Pod) error {
	// Stop Pod Containers
	err := p.StopPod(pod)
	if err != nil {
		log.Printf("Failed to stop pod: %v", err)
	}

	// Delete Pod Containers
	for _, ctrConfig := range pod.Spec.Containers {
		err := p.CtrService.DeleteContainer(ctrConfig.ID)
		if err != nil {
			log.Printf("Failed to delete container %s: %v", ctrConfig.Name, err)
			return err
		}
	}

	// Delete Pause Container
	err = p.CtrService.DeleteContainer(pod.Spec.PauseContainerID)
	if err != nil {
		log.Printf("Failed to delete pause container: %v", err)
		return err
	}

	// TODO: 修改其他的状态？
	(*pod).Spec.PauseContainerID = ""

	return nil
}
