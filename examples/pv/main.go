package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// 创建PV

// 创建PVC

// 检查PVC能否自动绑定到PV

// 直接创建PVC,观察PV能否自动创建

// Pod通过名称绑定到PVC

// TODO: 删除PVC可以导致PV被删除，至少可以保证PV解绑

func main() {
	client := apiserver.NewAPIClient("")
	// 创建 PersistentVolume
	pv := object.PersistentVolume{
		Metadata: object.Metadata{
			Name:      "example-pv",
			Namespace: "default",
		},
		Spec: object.PersistentVolumeSpec{
			Capacity: object.ResourceList{Storage: "10Gi"},
			HostPath: &object.HostPathVolumeSource{
				Path: "/tmp/example-pv",
			},
		},
	}
	err := client.CreatePV(&pv)

	if err != nil {
		panic("创建 PersistentVolume 失败: " + err.Error())
	}

	// 创建 PersistentVolumeClaim
	pvc := object.PersistentVolumeClaim{
		Metadata: object.Metadata{
			Name:      "example-pvc-2",
			Namespace: "default",
		},
		Spec: object.PersistentVolumeClaimSpec{
			Capacity: object.ResourceList{Storage: "5Gi"},
			// VolumeName: "example-pv", // 如果需要绑定到特定 PV，可以设置此字段
		},
	}

	err = client.CreatePVC(&pvc)
	if err != nil {
		panic("创建 PersistentVolumeClaim 失败: " + err.Error())
	}

	// 创建 Pod，绑定PVC
	pod := object.Pod{
		Metadata: object.Metadata{
			Name:      "example-pod",
			Namespace: "default",
		},
		Spec: object.PodSpec{
			Containers: []object.Container{
				{
					Name:  "example-container",
					Image: "nginx:latest",
					VolumeMounts: []object.VolumeMount{
						{
							Name:      "example-pvc",
							MountPath: "/usr/share/nginx/html", // 挂载路径
						},
					},
				},
			},
			Volumes: []object.Volume{
				{
					Name: "example-pvc",
					PersistentVolumeClaim: &object.PersistentVolumeClaimName{
						ClaimName: "example-pvc-2", // 绑定到 PVC
					},
				},
			},
		},
	}

	err = client.CreatePod(&pod)
	if err != nil {
		panic("创建 Pod 失败: " + err.Error())
	}
}

// TODO: 测试自动创建
