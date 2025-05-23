package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	client := apiserver.NewAPIClient("")

	// 创建 Pod
	pod := &object.Pod{
		Metadata: object.Metadata{
			Name:      "nginx-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "nginx",
			},
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
		Status: object.PodStatus{
			IP:    "172.17.0.2", // 假设 Pod 的 IP
			Phase: "Running",
		},
	}

	if err := client.CreatePod(pod); err != nil {
		panic(err)
	}

	// Create a service
	service := &object.Service{
		Kind: "Service",
		Type: "NodePort",
		Metadata: object.Metadata{
			Name:      "nginx-service2",
			Namespace: "default",
		},
		Spec: object.ServiceSpec{
			Selector: map[string]string{
				"app": "nginx",
			},
			Ports: []object.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: 80,
					NodePort:   30080,
				},
			},
		},
		Status: object.ServiceStatus{
			ClusterIP: "10.96.123.45", // 假设 ClusterIP
		},
	}

	if err := client.CreateService(service); err != nil {
		panic(err)
	}
}
