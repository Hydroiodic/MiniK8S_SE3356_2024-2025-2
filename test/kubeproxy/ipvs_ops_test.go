package kubeproxy_test

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// 测试ClusterIP

func TestClusterIP(t *testing.T) {
	containerService, err := container.NewContainerService()
	if err != nil {
		t.Fatalf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	// 初始化 IPVS 操作对象
	ops := ipvs_ops.NewIpvsOps(
		ipvs_ops.CLUSTER_CIDR_DEFAULT,
	) // 假设 ClusterIPCIDR 为 "222.111.0.0/16"
	defer ops.Close()
	ops.Init()

	// 1. 创建一个 Pod
	pod := &object.Pod{
		Kind: "Pod",
		Metadata: object.Metadata{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
					Ports: []object.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			},
		},
	}

	err = podService.CreatePod(pod)
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	t.Logf("Pod created: %v", pod)

	err = podService.StartPod(pod)
	if err != nil {
		t.Fatalf("failed to start pod: %v", err)
	}

	t.Logf("Pod started: %s", pod.Metadata.Name)

	// 2. 获取 Pod 的 IP 地址
	podIP := pod.Status.IP
	if podIP == "" {
		t.Fatalf("failed to get pod IP")
	}

	t.Logf("Pod IP: %s", podIP)

	// 2. 创建一个ClusterIP类型的Service
	svc := &object.Service{
		Kind: "Service",
		Type: object.SERVICE_TYPE_CLUSTERIP_STR,
		Metadata: object.Metadata{
			Name:      "test-clusterip",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{"app": "test"},
			Ports: []object.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: 80,
				},
			},
		},
		Status: object.ServiceStatus{
			ClusterIP: "222.111.0.1", // 假设 ClusterIP
			Endpoints: []object.Endpoint{
				{
					IP:   podIP,
					Port: 80,
				},
			},
		},
	}

	ops.AddService(svc)
	t.Logf("Service added: %v", svc)

	// 4. 测试访问 ClusterIP
	clusterIPAddr := fmt.Sprintf("%s:%d", svc.Status.ClusterIP, 8080)
	output, err := exec.Command("curl", "--max-time", "3", clusterIPAddr).
		Output()
	if err != nil || !strings.Contains(string(output), "Welcome") {
		t.Fatalf(
			"Failed to curl ClusterIP Service: %v, output: %s",
			err,
			string(output),
		)
	}

	// 5. 删除 Service
	ops.DelService(svc)

	// 6. 测试删除后无法访问
	output, err = exec.Command("curl", "--max-time", "3", clusterIPAddr).
		Output()
	if err == nil && string(output) != "" {
		t.Fatalf(
			"ClusterIP Service still accessible after deletion, output: %s",
			string(output),
		)
	}
}
