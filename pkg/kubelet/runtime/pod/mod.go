package pod

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// 这传进去的指针都可能被修改。
type PodServiceInterface interface {
	CreatePod(pod *object.Pod) error
	StartPod(pod *object.Pod) error
	StopPod(pod *object.Pod) error
	DeletePod(pod *object.Pod) error
	GetPodInfo(podName string) (*object.Pod, error)
	GetPodStatus(pod *object.Pod) (string, error)
	ListPods() ([]object.Pod, error) // 根据节点上的 Container 倒推出 Pod

	// 更新状态并尝试重启容器
	AutoRestartPod(pod *object.Pod) error
}
