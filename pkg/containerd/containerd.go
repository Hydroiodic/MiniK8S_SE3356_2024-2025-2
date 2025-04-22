package containerd

import "github.com/containerd/containerd"

func NewContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}

	return client, nil
}
