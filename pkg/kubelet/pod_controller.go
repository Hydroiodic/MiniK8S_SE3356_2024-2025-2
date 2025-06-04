package kubelet

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"time"

	"slices"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/mqtemplate"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type PodController struct {
	kubelet    *object.Kubelet
	podService pod.PodServiceInterface
	apiClient  *apiserver.APIClient
}

// PodController 是一个控制器，用于管理 Pod 的生命周期
// 它会定期检查 Pod 的状态，并通知API Server
// API Server 会在需要时通知 PodController 创建、删除或更新 Pod
// 以确保本地 Pod 的状态与 API Server 上的 Pod 状态一致
func NewPodController(
	kubelet *object.Kubelet,
	podService pod.PodServiceInterface,
	apiClient *apiserver.APIClient,
) *PodController {
	return &PodController{
		kubelet:    kubelet,
		podService: podService,
		apiClient:  apiClient,
	}
}

func (c *PodController) CreatePodHandler(msg map[string]any) error {
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

	c.kubelet.Mu.RLock()
	// 遍历 Kubelet 的 Pod 列表，检查 Pod 是否已经存在
	for _, p := range c.kubelet.Pods {
		if p.Metadata.Name == pod.Metadata.Name &&
			p.Metadata.Namespace == pod.Metadata.Namespace {
			// 找到 Pod，不予处理
			c.kubelet.Mu.RUnlock()
			return nil
		}
	}
	c.kubelet.Mu.RUnlock()

	// 将 Pod 对象添加到 Kubelet 的 Pod 列表中
	// 设置 Pod 的状态为 PodCreating
	pod.Status.Phase = object.PodCreating // NOTE: Pod Phase

	// 添加到 Kubelet 的 Pod 列表中
	c.kubelet.Mu.Lock()
	c.kubelet.Pods = append(c.kubelet.Pods, pod)
	c.kubelet.Mu.Unlock()

	// 在这里可以对 Pod 进行进一步处理，比如创建或更新
	// TODO: 错误处理，创建失败时
	if err := c.podService.CreatePod(&pod); err != nil {
		log.Printf("Failed to create pod: %v", err)
		return err
	}

	// 启动 Pod
	if err := c.podService.StartPod(&pod); err != nil {
		log.Printf("Failed to start pod: %v", err)
		return err
	}

	log.Printf("Pod started: %s", pod.Metadata.Name)

	c.kubelet.Mu.Lock()
	for i, p := range c.kubelet.Pods {
		if p.Metadata.Name == pod.Metadata.Name &&
			p.Metadata.Namespace == pod.Metadata.Namespace {
			// Update the status of the Pod.
			c.kubelet.Pods[i].Status.Phase = object.PodRunning
			c.kubelet.Pods[i].Status.StartTime = time.Now()

			break
		}
	}
	c.kubelet.Mu.Unlock()

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

	c.kubelet.Mu.RLock()
	// 检查 Pod 是否存在
	podExists := false
	// 遍历 Kubelet 的 Pod 列表，检查 Pod 是否存在
	for _, p := range c.kubelet.Pods {
		if p.Metadata.Name == pod.Metadata.Name &&
			p.Metadata.Namespace == pod.Metadata.Namespace {
			// 找到 Pod，设置状态为 PodDeleting
			p.Status.Phase = object.PodDeleting
			// 更新布尔值
			podExists = true

			break
		}
	}
	c.kubelet.Mu.RUnlock()

	// 如果 Pod 不存在，直接返回
	if !podExists {
		fmt.Printf(
			"Pod %s does not exist, skipping deletion",
			pod.Metadata.Name,
		)

		return fmt.Errorf(
			"pod %s does not exist, skipping deletion",
			pod.Metadata.Name,
		)
	}

	// 删除 Pod 对象kubelet
	if err := c.podService.DeletePod(&pod); err != nil {
		log.Printf("Failed to delete pod: %v", err)
		return err
	}

	c.kubelet.Mu.Lock()
	for i, p := range c.kubelet.Pods {
		if p.Metadata.Name == pod.Metadata.Name &&
			p.Metadata.Namespace == pod.Metadata.Namespace {
			// Delete the Pod.
			c.kubelet.Pods = slices.Delete(c.kubelet.Pods, i, i+1)
			break
		}
	}
	c.kubelet.Mu.Unlock()

	return nil
}

func (c *PodController) Run(stopCh <-chan struct{}) {
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
		queueName := path.Join(
			mqtemplate.KubeletDeletePodQueue,
			c.kubelet.Config.Name,
		)
		if err := mqtemplate.ConsumeMessageOnQueue(queueName, c.DeletePodHandler); err != nil {
			log.Printf("Failed to consume message: %v", err)
		}
	}()

	for range stopCh {
		// When stopCh is closed, exit the loop
		return
	}
}
