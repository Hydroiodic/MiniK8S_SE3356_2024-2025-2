package containerd

import (
	containerd "github.com/containerd/containerd/v2/client"
)

func NewContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}
	return client, nil
}
