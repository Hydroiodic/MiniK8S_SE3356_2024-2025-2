package kubelet

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodStatusController struct {
	kubelet    *object.Kubelet
	podService pod.PodServiceInterface
	apiClient  APIServerClient
	period     time.Duration
}

// PodStatusController 是一个控制器，用于更新 Pod 的状态
// 它会定期检查 Pod 的状态，并在需要时重启容器
func NewPodStatusController(
	kubelet *object.Kubelet,
	podService pod.PodServiceInterface,
	apiClient APIServerClient,
	period time.Duration,
) *PodStatusController {
	return &PodStatusController{
		kubelet:    kubelet,
		podService: podService,
		apiClient:  apiClient,
		period:     period,
	}
}

func (c *PodStatusController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(c.period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.updatePodStatus()
		case <-stopCh:
			log.Println("Stopping pod status controller")
			return
		}
	}
}

func (c *PodStatusController) updatePodStatus() {
	// Get a copy of the pods to avoid holding the lock for too long.
	c.kubelet.Mu.RLock()
	pods := c.kubelet.Pods
	c.kubelet.Mu.RUnlock()

	log.Printf("Updating pod status for pods: %v", utils.ExtractPodNames(pods))

	for i, pod := range pods {
		status, err := c.podService.GetPodStatus(&pod)
		if err != nil {
			log.Printf(
				"Failed to get status for pod %s/%s: %v",
				pod.Metadata.Name,
				pod.Metadata.Namespace,
				err,
			)
		}

		log.Printf(
			"Pod %s/%s status: %s",
			pod.Metadata.Name,
			pod.Metadata.Namespace,
			status,
		)

		// 尝试重启
		err = c.podService.AutoRestartPod(&pod)
		if err != nil {
			log.Printf(
				"Failed to restart pod %s/%s: %v",
				pod.Metadata.Name,
				pod.Metadata.Namespace,
				err,
			)

			continue
		}

		c.kubelet.Mu.Lock()
		// 填写Pod的状态
		c.kubelet.Pods[i].Status.Phase = status
		c.kubelet.Mu.Unlock()
	}

	// TODO：上报状态到 API Server
	// 上报 kubelet 状态
	c.kubelet.Mu.RLock()
	kubeletCopy := &object.Kubelet{
		Config:         c.kubelet.Config,
		Pods:           c.kubelet.Pods,
		StartTime:      c.kubelet.StartTime,
		LastUpdateTime: time.Now(),
		// Copy other fields as needed, excluding the Mu
	}
	c.kubelet.Mu.RUnlock()

	// TODO: 发送心跳可行吗
	if err := c.apiClient.SendHeartbeat(kubeletCopy); err != nil {
		log.Printf("Failed to update node status to API Server: %v", err)
	}
}
