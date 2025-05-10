package docker

import (
	"sync"

	"github.com/docker/docker/client"
)

// dockerClientSingleton 管理 Docker 客户端单例
type dockerClientSingleton struct {
	client *client.Client
	once   sync.Once
}

// 全局单例实例
var singleton *dockerClientSingleton = &dockerClientSingleton{}

// GetDockerClient 获取单例 Docker 客户端
func GetDockerClient() *client.Client {
	singleton.once.Do(func() {
		// 初始化 Docker 客户端
		cli, _ := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		singleton.client = cli
	})

	return singleton.client
}
