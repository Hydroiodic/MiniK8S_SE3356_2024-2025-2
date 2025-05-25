package kubelet

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type APIServerClient interface {
	// 发送心跳
	SendHeartbeat(kubelet *object.Kubelet) error
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

func (a *APIServerClientImpl) RegisterKubelet(kubelet *object.Kubelet) error {
	return a.c.RegisterKubelet(kubelet)
}
