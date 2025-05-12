package runtime_test

import (
	"testing"
	"time"

	cadvisor "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/cadvisor"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestCreateCadvisorContainer(t *testing.T) {
	// 定义容器规格
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "docker.io/library/nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}
	containerID, err := containerService.RunCadvisorContainer()

	// 创建一个新的容器
	containerID2, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)
	containerService.StartContainer(containerID2)
	cadvisor.GetContainerCPUandMem("localhost", "8090", "test-container")

	// 检查容器是否存在
	info, err := containerService.GetContainerInfo(containerID2)
	t.Logf("Container info: %+v", info)
	assert.NoError(t, err)
	// 检查容器是否存在
	info2, err := containerService.GetContainerInfo(containerID)
	t.Logf("Container info: %+v", info2)
	assert.NoError(t, err)
	t.Logf("Container created with ID: %s", containerID2)
	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
	// 删除容器
	err = containerService.DeleteContainer(containerID2)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID2)
}

func TestCreateContainer(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "docker.io/library/nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)

	// 检查容器是否存在
	info, err := containerService.GetContainerInfo(containerID)
	t.Logf("Container info: %+v", info)
	assert.NoError(t, err)

	t.Logf("Container created with ID: %s", containerID)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
}

func TestForceCreateContainer(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.ForceCreateContainer(
		containerSpec,
		nil,
	)
	assert.NoError(t, err)

	t.Logf("Container created with ID: %s", containerID)

	// 创建一个新的容器
	containerID, err = containerService.ForceCreateContainer(containerSpec, nil)
	assert.NoError(t, err)

	t.Logf("Container created with ID: %s", containerID)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
}

func TestStartContainer(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)

	t.Logf("Container created with ID: %s", containerID)

	// 启动容器
	err = containerService.StartContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container started with ID: %s", containerID)

	// 等待3秒钟
	time.Sleep(3 * time.Second)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
}

func TestDeleteContainer(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)

	t.Logf("Container created with ID: %s", containerID)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)

	// 检查容器是否已删除
	_, err = containerService.GetContainerInfo(containerID)
	assert.Error(t, err)
	t.Logf("Container %s not found, as expected", containerID)
}

func TestGetContainerStatus(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "nginx:latest",
		Command: []string{
			"nginx",
			"-g",
			"daemon off;",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)
	t.Logf("Container created with ID: %s", containerID)

	// 获取容器状态
	status, err := containerService.GetContainerStatus(containerID)
	assert.NoError(t, err)
	assert.Equal(t, "created", status)
	t.Logf("Container status: %s", status)

	// 启动容器
	err = containerService.StartContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container started with ID: %s", containerID)

	// 获取容器状态
	status, err = containerService.GetContainerStatus(containerID)
	assert.NoError(t, err)
	assert.Equal(t, "running", status)
	t.Logf("Container status: %s", status)

	// 停止容器
	err = containerService.StopContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container stopped with ID: %s", containerID)

	// 获取容器状态
	status, err = containerService.GetContainerStatus(containerID)
	assert.NoError(t, err)
	assert.Equal(t, "exited", status)
	t.Logf("Container status: %s", status)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)

	// 检查容器是否已删除
	_, err = containerService.GetContainerInfo(containerID)
	assert.Error(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
}

func TestContainerExecCommand(t *testing.T) {
	containerService, err := container.NewContainerService()
	assert.NoError(t, err)

	// 定义容器规格
	containerSpec := object.Container{
		Name:  "test-container",
		Image: "alpine:latest",
		Command: []string{
			"sh",
			"-c",
			"while true; do sleep 1; done",
		},
	}

	// 创建一个新的容器
	containerID, err := containerService.CreateContainer(containerSpec, nil)
	assert.NoError(t, err)
	t.Logf("Container created with ID: %s", containerID)

	// 启动容器
	err = containerService.StartContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container started with ID: %s", containerID)

	// 执行命令
	output, err := containerService.ExecCommand(
		containerID,
		[]string{"echo", "Hello World"},
	)
	assert.NoError(t, err)
	t.Logf("Command output: %s", output)

	// 停止容器
	err = containerService.StopContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container stopped with ID: %s", containerID)

	// 删除容器
	err = containerService.DeleteContainer(containerID)
	assert.NoError(t, err)
	t.Logf("Container deleted with ID: %s", containerID)
}
