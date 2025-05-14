package kubelet

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

type APIServerClient interface {
	// 发送心跳
	SendHeartbeat(kubelet *object.Kubelet) error
	// 获取最新的 Pod 配置
	FetchPods(nodeName string) ([]object.Pod, error)
	// 发送 Kubelet 的状态
	UpdateNodeStatus(kubelet *object.Kubelet) error
}
