package kubelet

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodStatusController struct {
	kubelet    *object.Kubelet
	podService pod.PodServiceInterface
	apiClient  *apiserver.APIClient
	period     time.Duration
}

// PodStatusController 是一个控制器，用于更新 Pod 的状态
// 它会定期检查 Pod 的状态，并在需要时重启容器
func NewPodStatusController(
	kubelet *object.Kubelet,
	podService pod.PodServiceInterface,
	apiClient *apiserver.APIClient,
	period time.Duration,
) *PodStatusController {
	return &PodStatusController{
		kubelet:    kubelet,
		podService: podService,
		apiClient:  apiClient,
		period:     period,
	}
}

func (c *PodStatusController) UpdatePodRoutine(stopCh <-chan struct{}) {
	// Create a ticker that ticks every `c.period` duration.
	ticker := time.NewTicker(c.period)
	defer ticker.Stop()

	log.Printf("Starting pod status controller")

	for {
		select {
		case <-ticker.C:
			c.updatePodStatus()
		case <-stopCh:
			log.Printf("Stopping pod status controller")
			return
		}
	}
}

func (c *PodStatusController) HeartBeatRoutine(stopCh <-chan struct{}) {
	// Create a ticker that ticks every 5 seconds for the kubelet heartbeat.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Printf("Starting kubelet heartbeat controller")

	for {
		select {
		case <-ticker.C:
			err := c.apiClient.HeartbeatKubelet(c.kubelet)
			if err != nil {
				log.Printf("Failed to heartbeat kubelet: %v", err)
			} else {
				log.Printf("Kubelet heartbeat sent successfully")
			}
		case <-stopCh:
			log.Printf("Stopping kubelet heartbeat controller")
			return
		}
	}
}

func (c *PodStatusController) updatePodStatus() {
	// Get a copy of the pods from the kubelet.
	pods := c.kubelet.Pods
	// Log the start of the pod status update process.
	log.Printf("Updating pod status for pods: %v", utils.ExtractPodNames(pods))

	// Iterate over the pods and update their status.
	for i, pod := range pods {
		// Get the current status of the pod.
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

		// Try to restart the pod if its status is not running.
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

		// Now let's update the pod status in `pods` slice.
		// NOTE: local variable, no need to lock the kubelet's Mu here.
		pods[i].Status.Phase = status
		pods[i].Status.IP = pod.Status.IP
		pods[i].Spec.PauseContainerID = pod.Spec.PauseContainerID
	}

	c.kubelet.Mu.Lock()
	defer c.kubelet.Mu.Unlock()

	// Now with the lock held, we can safely update the kubelet's pods.
	for _, pod := range pods {
		for j := range c.kubelet.Pods {
			// Try to find the pod in the kubelet's pods slice.
			if c.kubelet.Pods[j].Metadata.Name == pod.Metadata.Name &&
				c.kubelet.Pods[j].Metadata.Namespace == pod.Metadata.Namespace {
				// Update the pod status in the kubelet's pods slice.
				c.kubelet.Pods[j].Status.Phase = pod.Status.Phase
				c.kubelet.Pods[j].Status.IP = pod.Status.IP
				c.kubelet.Pods[j].Spec.PauseContainerID = pod.Spec.PauseContainerID
			}
		}
	}
}
