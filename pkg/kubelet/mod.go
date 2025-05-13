package kubelet

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"

type KubeletServiceInterface interface {
	Run()
	SyncPods() error
	NoticeServer() error

	Reconcile(target []object.Pod, source []object.Pod) error
}
