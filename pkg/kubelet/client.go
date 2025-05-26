package kubelet

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type APIServerClient interface {
	// Send Heartbeat to API Server (Main sync routine).
	SendHeartbeat(kubelet *object.Kubelet) error
	// Register Kubelet to API Server.
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
	return a.c.HeartbeatKubelet(kubelet)
}

func (a *APIServerClientImpl) RegisterKubelet(kubelet *object.Kubelet) error {
	return a.c.RegisterKubelet(kubelet)
}
