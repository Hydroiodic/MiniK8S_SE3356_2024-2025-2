package pod

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

type PodServiceInterface interface {
	CreatePod(pod *object.Pod) error
	StartPod(pod *object.Pod) error
	StopPod(pod *object.Pod) error
	DeletePod(pod *object.Pod) error
	GetPodInfo(podName string) (*object.Pod, error)
	GetPodStatus(podName string) (string, error)
	ListPods() ([]object.Pod, error)
	GetPodContainerStatus(podName string) (map[string]string, error)
	GetPodContainerInfo(podName string) (map[string]*object.Container, error)
}
