package kubelet

import (
	"encoding/json"
	"log"
	"path"
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

func (c *PodController) DeletePodHandler(msg map[string]interface{}) error {
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

	// 删除 Pod 对象
	if err := c.podService.DeletePod(&pod); err != nil {
		log.Printf("Failed to delete pod: %v", err)
		return err
	}

	// Notice API Server
	if err := c.apiClient.DeletePodFromEtcd(&pod); err != nil {
		log.Printf(
			"Failed to notify API Server about pod deletion %v",
			err,
		)
	}

	return nil
}

func (c *PodController) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(c.syncPeriod)
	defer ticker.Stop()

	// 处理消息队列中的 Pod 创建请求
	// TODO: 这玩意停不住啊？
	go func() {
		// The name of the queue listening to is `KubeletCreatePodQueue/nodeName`.
		queueName := path.Join(
			mqtemplate.KubeletCreatePodQueue,
			c.kubelet.Config.Name,
		)
		if err := mqtemplate.ConsumeMessageOnQueue(queueName, c.CreatePodHandler); err != nil {
			log.Printf("Failed to consume message: %v", err)
		}
	}()

	go func() {
		err := mqtemplate.ConsumeMessageOnQueue(
			mqtemplate.KubeletDeletePodQueue,
			c.DeletePodHandler,
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

	// TODO: DesiredPods 会缺少字段，不要使用这个直接对 Kubelet 进行更新
	desiredPods, err := c.apiClient.FetchPods(c.kubelet.Config.Name)

	useCache := false
	if err != nil {
		useCache = true
		desiredPods = currentPods

		log.Printf("API Server unavailable, using cached pods: %v", err)
	}

	log.Printf("Desired pods: %v", utils.ExtractPodNames(desiredPods))

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
			if _, exists := desiredMap[key]; exists {
				continue
			}
			// Pod 不在 desiredPods 中，删除它
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

			c.kubelet.Mu.Lock()
			// 从 Kubelet 的 Pod 列表中删除
			for i, p := range c.kubelet.Pods {
				if p.Metadata.Name == pod.Metadata.Name &&
					p.Metadata.Namespace == pod.Metadata.Namespace {
					c.kubelet.Pods = append(
						c.kubelet.Pods[:i],
						c.kubelet.Pods[i+1:]...,
					)

					break
				}
			}
			c.kubelet.Mu.Unlock()
			log.Printf("Pod deleted: %s", pod.Metadata.Name)
		}
	}

	// 创建缺少的 Pod
	for key, pod := range desiredMap {
		if _, exists := currentMap[key]; !exists {
			log.Printf("Creating pod %s (cache=%v)", key, useCache)

			// 创建 Pod
			if err := c.podService.CreatePod(&pod); err != nil {
				log.Printf("Failed to create pod %s: %v", key, err)
				continue
			}
			// 启动 Pod
			if err := c.podService.StartPod(&pod); err != nil {
				log.Printf("Failed to start pod %s: %v", key, err)
				continue
			}

			// 记得上锁
			c.kubelet.Mu.Lock()
			c.kubelet.Pods = append(c.kubelet.Pods, pod)
			c.kubelet.Mu.Unlock()
		}
	}

	currentMap = make(map[string]object.Pod)

	// 验证已有的 Pod和 期望的 Pod 是否一致
	currentPods, err := c.podService.ListPods()
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return
	}

	// 构建 currentPods 的 map
	for _, pod := range currentPods {
		key := utils.GeneratePodNsNameLabel(
			pod.Metadata.Namespace,
			pod.Metadata.Namespace,
		)
		currentMap[key] = pod
	}

	// TODO: 验证状态，仅作调试用
	for key := range desiredMap {
		// If not in currentPods, panic
		currPod, exists := currentMap[key]

		if !exists {
			log.Printf("Error: Pod %s not found in current pods after reconcile", key)
			continue
		}

		if currPod.Spec.PauseContainerID == "" {
			log.Printf("Error: Pod %s is not running after reconcile", key)
			continue
		}

		// 检查所有容器的ContainerID 是否存在
		for _, ctr := range currPod.Spec.Containers {
			if ctr.ID == "" {
				log.Printf(
					"Error: Container %s in pod %s is not running after reconcile",
					ctr.Name,
					key,
				)
				continue
			}
		}
	}
}
