package kubelet

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func NewKubelet(config object.KubeletConfig) *object.Kubelet {
	return &object.Kubelet{
		Config:         config,
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
		// CachedPods:     []object.Pod{},
	}
}

type KubeletService struct {
	kubelet          *object.Kubelet
	podController    *PodController
	statusController *PodStatusController
	apiClient        APIServerClient
}

func NewKubeletService(
	config object.KubeletConfig,
	podService pod.PodServiceInterface,
	apiClient APIServerClient,
) *KubeletService {
	kubelet := NewKubelet(config)
	podController := NewPodController(
		kubelet,
		podService,
		apiClient,
		30*time.Second,
	)
	statusController := NewPodStatusController(
		kubelet,
		podService,
		apiClient,
		10*time.Second,
	)

	return &KubeletService{
		kubelet:          kubelet,
		podController:    podController,
		statusController: statusController,
		apiClient:        apiClient,
	}
}

func (s *KubeletService) Run(stopCh <-chan struct{}) {
	err := s.apiClient.RegisterKubelet(s.kubelet)
	if err != nil {
		log.Printf("Failed to register kubelet: %v", err)
	}

	// 先恢复本地状态
	localPods, err := s.podController.podService.ListPods()
	if err != nil {
		log.Printf("Failed to fetch pods: %v", err)
	}

	log.Printf("Restoring local pods: %v", utils.ExtractPodNames(localPods))
	s.kubelet.Pods = localPods

	go s.podController.Run(stopCh)
	go s.statusController.Run(stopCh)
	<-stopCh
}
