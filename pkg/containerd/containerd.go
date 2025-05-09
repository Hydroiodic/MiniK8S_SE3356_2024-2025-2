package containerd

import "github.com/containerd/containerd"

func NewContainerdClient() (*containerd.Client, error) {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}

	return client, nil
}

// 拉取镜像（如果已经存在则不拉取）

// 创建容器（创建SnapShot和创建容器一体）
// 能够处理SnapShot已经存在的情况
