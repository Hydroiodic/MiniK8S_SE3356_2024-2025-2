package containerd_test

import (
	"testing"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/containerd"
	"github.com/stretchr/testify/assert"
)

func TestNewContainerdClient(t *testing.T) {
	// 调用 NewContainerdClient 函数
	client, err := containerd.NewContainerdClient()

	// 检查是否返回错误
	assert.NoError(t, err, "expected no error when creating containerd client")

	// 检查返回的 client 是否不为 nil
	assert.NotNil(t, client, "expected a valid containerd client")

	// 关闭 client 以释放资源
	if client != nil {
		err := client.Close()
		assert.NoError(
			t,
			err,
			"expected no error when closing containerd client",
		)
	}
}
