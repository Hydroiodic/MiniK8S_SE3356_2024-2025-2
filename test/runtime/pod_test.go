package runtime_test

import (
	"testing"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestCreatePod(t *testing.T) {
	containerService, err := container.NewContainerService()
	if err != nil {
		t.Fatalf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	pod := &object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "test-pod",
			Namespace: "example-pod-namespace",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:  "test-container",
					Image: "docker.io/library/nginx:latest",
					// Image:   "nginx:latest",
					Command: []string{"nginx", "-g", "daemon off;"},
					Ports: []object.ContainerPort{
						{ContainerPort: 80},
					},
				},
			},
		},
		// Status: object.PodStatus{
		// 	Phase:      "Pending",
		// 	StartTime:  time.Now(),
		// 	Conditions: []string{"Initialized", "Ready"},
		// },
	}

	err = podService.CreatePod(pod)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	t.Logf("Pod created: %v", pod)

	containerId := pod.Spec.Containers[0].ID
	assert.NotEmpty(t, containerId, "Container ID should not be empty")

	// 检查 Pause 容器是否存在
	res, err := containerService.GetContainerInfo(containerId)

	if err != nil {
		t.Fatalf("failed to get pause container info: %v", err)
	}

	t.Logf("Pause container info: %v", res)
}

// TestPodContainerCommunication 测试 Pod 内部容器间的网络通信
func TestPodContainerCommunication(t *testing.T) {
	// 初始化容器服务
	containerService, err := container.NewContainerService()
	if err != nil {
		t.Fatalf("failed to create container service: %v", err)
	}

	// 初始化 Pod 服务
	podService := pod.NewPodService(containerService)

	// 定义一个包含两个容器的 Pod
	pod := &object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "test-pod-communication",
			Namespace: "example-pod-namespace",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:    "server-container",
					Image:   "docker.io/library/nginx:latest",
					Command: []string{"nginx", "-g", "daemon off;"},
					Ports: []object.ContainerPort{
						{ContainerPort: 80}, // 服务端监听 80 端口
					},
				},
				{
					Name:    "client-container",
					Image:   "docker.io/library/busybox:latest", // 替换为 busybox
					Command: []string{"sh", "-c", "sleep 3600"}, // 保持容器运行
				},
			},
		},
	}

	// 创建 Pod
	err = podService.CreatePod(pod)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	t.Logf("Pod created for communication test: %v", pod)

	err = podService.StartPod(pod)
	if err != nil {
		t.Fatalf("failed to start pod: %v", err)
	}

	t.Logf("Pod started for communication test: %v", pod)

	// 等待 Pod 和容器启动
	time.Sleep(5 * time.Second) // 等待容器启动，实际环境中可能需要更精确的健康检查

	// 验证两个容器的 ID 不为空
	serverContainerID := pod.Spec.Containers[0].ID

	clientContainerID := pod.Spec.Containers[1].ID

	assert.NotEmpty(
		t,
		serverContainerID,
		"Server container ID should not be empty",
	)
	assert.NotEmpty(
		t,
		clientContainerID,
		"Client container ID should not be empty",
	)

	// 获取容器信息
	serverInfo, err := containerService.GetContainerInfo(serverContainerID)
	if err != nil {
		t.Fatalf("failed to get server container info: %v", err)
	}

	clientInfo, err := containerService.GetContainerInfo(clientContainerID)
	if err != nil {
		t.Fatalf("failed to get client container info: %v", err)
	}

	t.Logf("Server container info: %v", serverInfo)
	t.Logf("Client container info: %v", clientInfo)

	// 测试容器间通信：从客户端容器向服务端容器发起 HTTP 请求
	testCommand := []string{
		"wget",
		"-q",
		"-O",
		"-",
		"http://localhost:80", // 使用 wget 访问服务端
	}

	output, err := containerService.ExecCommand(clientContainerID, testCommand)
	if err != nil {
		t.Fatalf("failed to execute wget command in client container: %v", err)
	}

	// 验证通信结果
	t.Logf("Wget command output: %s", output)
	assert.Contains(
		t,
		output,
		"Welcome to nginx",
		"Expected nginx welcome page in response",
	)

	// 清理 Pod
	err = podService.DeletePod(pod)
	if err != nil {
		t.Logf("failed to delete pod: %v", err)
	}
}

func TestListPods(t *testing.T) {
	containerService, err := container.NewContainerService()
	if err != nil {
		t.Fatalf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	// 创建一个测试 Pod
	testPod := &object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "test-pod-list",
			Namespace: "example-pod-namespace",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:    "client-container",
					Image:   "docker.io/library/busybox:latest", // 替换为 busybox
					Command: []string{"sh", "-c", "sleep 3600"}, // 保持容器运行
				},
			},
		},
	}
	err = podService.CreatePod(testPod)

	if err != nil {
		t.Fatalf("failed to create test pod: %v", err)
	}

	pods, err := podService.ListPods()
	if err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}

	t.Logf("List of pods: %v", pods)

	for _, p := range pods {
		t.Logf("Pod: %s/%s", p.Metadata.Namespace, p.Metadata.Name)
	}

	// 清理测试 Pod
	err = podService.DeletePod(testPod)
	if err != nil {
		t.Logf("failed to delete test pod: %v", err)
	}

	t.Logf(
		"Test pod deleted: %s/%s",
		testPod.Metadata.Namespace,
		testPod.Metadata.Name,
	)
}
