package kubelet

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodController struct {
	kubelet    *object.Kubelet
	podService pod.PodServiceInterface
	apiClient  APIServerClient
	syncPeriod time.Duration
}

// PodController 是一个控制器，用于管理 Pod 的生命周期
// 它会定期检查 Pod 的状态，并在需要时创建、删除或更新 Pod
// 它还会从 API Server 获取最新的 Pod 配置，并与本地缓存进行比较
// 以确保本地 Pod 的状态与 API Server 上的 Pod 状态一致
func NewPodController(
	kubelet *object.Kubelet,
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

func (c *PodController) CreatePodHandler(msg map[string]interface{}) error {
	// 解析消息体
	msgBody, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	// 解析 JSON 为 Pod 对象
	var pod object.Pod
	if err := json.Unmarshal(msgBody, &pod); err != nil {
		log.Printf("Failed to unmarshal message to Pod: %v", err)
		return err
	}

	// 将 Pod 对象添加到 Kubelet 的 Pod 列表中
	c.kubelet.Mu.Lock()
	c.kubelet.Pods = append(c.kubelet.Pods, pod)
	c.kubelet.Mu.Unlock()

	// 在这里可以对 Pod 进行进一步处理，比如创建或更新
	if err := c.podService.CreatePod(&pod); err != nil {
		log.Printf("Failed to create pod: %v", err)
	}

	err = c.podService.StartPod(&pod)
	if err != nil {
		log.Printf("Failed to start pod: %v", err)
	}

	log.Printf("Pod started: %s", pod.Metadata.Name)

	return nil
}

func (c *PodController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(c.syncPeriod)
	defer ticker.Stop()

	// 处理消息队列中的 Pod 创建请求
	// TODO: 这玩意停不住啊？
	go func() {
		err := mqtemplate.ConsumeMessageOnQueue(
			mqtemplate.KubeletCreatePodQueue,
			c.CreatePodHandler,
		)
		if err != nil {
			log.Printf("Failed to consume message: %v", err)
		}
	}()

	for {
		select {
		case <-ticker.C:
			c.SyncPods()
		case <-stopCh:
			return
		}
	}
}

func (c *PodController) SyncPods() {
	currentPods, err := c.podService.ListPods()
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
	}

	log.Printf("Current pods: %v", len(currentPods))

	desiredPods, err := c.apiClient.FetchPods(c.kubelet.Config.Name)

	useCache := false
	if err != nil {
		useCache = true
		desiredPods = currentPods

		log.Printf("API Server unavailable, using cached pods: %v", err)
	}

	log.Printf("Desired pods: %v", desiredPods)

	c.Reconcile(desiredPods, currentPods, useCache)
}

func (c *PodController) Reconcile(
	desiredPods, currentPods []object.Pod,
	useCache bool,
) {
	// 使用 map 优化查找
	desiredMap := make(map[string]object.Pod)
	currentMap := make(map[string]object.Pod)

	// 构建 desiredPods 的 map
	for _, pod := range desiredPods {
		key := utils.GeneratePodNsNameLabel(
			pod.Metadata.Namespace,
			pod.Metadata.Namespace,
		)
		desiredMap[key] = pod
	}

	// 构建 currentPods 的 map
	for _, pod := range currentPods {
		key := utils.GeneratePodNsNameLabel(
			pod.Metadata.Namespace,
			pod.Metadata.Namespace,
		)
		currentMap[key] = pod
	}

	// 删除多余的 Pod（仅在非缓存模式下执行）
	if !useCache {
		for key, pod := range currentMap {
			if _, exists := desiredMap[key]; !exists {
				log.Printf("Deleting pod %s (cache=%v)", key, useCache)

				if err := c.podService.DeletePod(&pod); err != nil {
					log.Printf("Failed to delete pod %s: %v", key, err)
					continue
				}

				// Notice API Server
				if err := c.apiClient.DeletePodFromEtcd(&pod); err != nil {
					log.Printf(
						"Failed to notify API Server about pod deletion %s: %v",
						key,
						err,
					)

					continue
				}
			}
		}
	}

	// 创建缺少的 Pod
	for key, pod := range desiredMap {
		if _, exists := currentMap[key]; !exists {
			log.Printf("Creating pod %s (cache=%v)", key, useCache)

			if err := c.podService.CreatePod(&pod); err != nil {
				log.Printf("Failed to create pod %s: %v", key, err)
				continue
			}
		}
	}

	// 更新 Kubelet.Pods
	c.kubelet.Mu.Lock()
	c.kubelet.Pods = desiredPods
	c.kubelet.Mu.Unlock()
}
