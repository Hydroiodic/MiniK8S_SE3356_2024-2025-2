package apiserver

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

func (c *APIClient) GetNodes() ([]object.Kubelet, error) {
	var nodes []object.Kubelet
	err := c.getAndUnmarshalList(KubeletGetNodesURL, &nodes)

	return nodes, err
}
