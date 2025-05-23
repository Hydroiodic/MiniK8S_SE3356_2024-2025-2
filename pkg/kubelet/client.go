package kubelet

import (
	"fmt"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type APIServerClient interface {
	// 发送心跳
	SendHeartbeat(kubelet *object.Kubelet) error
	// 获取最新的 Pod 配置
	FetchPods(nodeName string) ([]object.Pod, error)
	RegisterKubelet(kubelet *object.Kubelet) error
}

type APIServerClientImpl struct {
	c *apiserver.APIClient
}

func NewAPIServerClient(c *apiserver.APIClient) APIServerClient {
	return &APIServerClientImpl{
		c: c,
	}
}

func (a *APIServerClientImpl) SendHeartbeat(kubelet *object.Kubelet) error {
	// 发送心跳到 API Server
	return a.c.HeartbeatKubelet(kubelet)
}

func (a *APIServerClientImpl) FetchPods(nodeName string) ([]object.Pod, error) {
	// 获取最新的 Pod 配置
	nodes, err := a.c.GetNodes()
	if err != nil {
		return nil, err
	}

	for i := range nodes {
		node := &nodes[i]
		if node.Config.Name == nodeName {
			return node.Pods, nil
		}
	}

	return nil, fmt.Errorf("node %s not found", nodeName)
}

func (a *APIServerClientImpl) RegisterKubelet(kubelet *object.Kubelet) error {
	return a.c.RegisterKubelet(kubelet)
}
