package kubelet

import (
	"fmt"
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodController struct {
	kubelet    *Kubelet
	podService pod.PodServiceInterface
	apiClient  APIServerClient
	syncPeriod time.Duration
}

// PodController 是一个控制器，用于管理 Pod 的生命周期
// 它会定期检查 Pod 的状态，并在需要时创建、删除或更新 Pod
// 它还会从 API Server 获取最新的 Pod 配置，并与本地缓存进行比较
// 以确保本地 Pod 的状态与 API Server 上的 Pod 状态一致
func NewPodController(
	kubelet *Kubelet,
	podService pod.PodServiceInterface,
	apiClient APIServerClient,
	syncPeriod time.Duration,
) *PodController {
	return &PodController{
		kubelet:    kubelet,
		podService: podService,
		apiClient:  apiClient,
		syncPeriod: syncPeriod,
	}
}

func (c *PodController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(c.syncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.kubelet.SyncPods()
		case <-stopCh:
			return
		}
	}
}

func (c *PodController) SyncPods() {
	// 1. 从运行时获取当前节点上所有 Pod 的运行状态
	currentPods, err := c.podService.ListPods()
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	useCache := false
	// 2. 从API Server 获取完整的 Pod 信息
	desiredPods, err := c.apiClient.ListPods(c.kubelet.Config.Name)
	if err != nil {
		log.Printf("API Server unavailable, using cached pods: %v", err)
		c.kubelet.mutex.RLock()
		desiredPods = c.kubelet.CachedPods
		c.kubelet.mutex.RUnlock()
		useCache = true
	} else {
		c.kubelet.mutex.Lock()
		c.kubelet.CachedPods = desiredPods
		c.kubelet.mutex.Unlock()
	}

	// 3. 添加缺少，删除多余
	return c.Reconcile(desiredPods, currentPods, useCache)
}

func (c *PodController) Reconcile(
	desiredPods, currentPods []object.Pod,
	useCache bool,
) error {
	desiredMap := make(map[string]object.Pod)
	for _, pod := range desiredPods {
		key := pod.Metadata.Name + "/" + pod.Metadata.Namespace
		desiredMap[key] = pod
	}

	if !useCache {
		for _, pod := range currentPods {
			key := pod.Metadata.Name + "/" + pod.Metadata.Namespace
			if _, exists := desiredMap[key]; !exists {
				if err := c.podService.DeletePod(pod); err != nil {
					log.Printf("Failed to delete pod %s: %v", key, err)
					continue
				}
				log.Printf("Deleted pod %s", key)
			}
		}
	}

	for key, pod := range desiredMap {
		if _, exists := findPod(currentPods, pod.Metadata.Name, pod.Metadata.Namespace); !exists {
			if err := c.podService.CreatePod(pod); err != nil {
				log.Printf("Failed to create pod %s: %v", key, err)
				continue
			}
			log.Printf("Created pod %s", key)
		}
	}

	c.kubelet.mutex.Lock()
	c.kubelet.Pods = desiredPods
	c.kubelet.mutex.Unlock()
	return nil
}
