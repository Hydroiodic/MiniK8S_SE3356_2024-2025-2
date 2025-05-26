package kubelet

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func NewKubelet(config object.KubeletConfig) *object.Kubelet {
	return &object.Kubelet{
		Config:         config,
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}
}

type KubeletService struct {
	kubelet           *object.Kubelet
	podController     *PodController
	statusController  *PodStatusController
	serviceController *kubeproxy.ServiceController // TODO: 添加Service Controller
	apiClient         *apiserver.APIClient
}

func NewKubeletService(
	config object.KubeletConfig,
	podService pod.PodServiceInterface,
	apiClient *apiserver.APIClient,
) *KubeletService {
	kubelet := NewKubelet(config)

	podController := NewPodController(
		kubelet,
		podService,
		apiClient,
	)

	// TODO: 改变Client
	serviceController := kubeproxy.NewServiceController(
		kubelet,
		ipvs_ops.NewIpvsOps(ipvs_ops.CLUSTER_CIDR_DEFAULT),
		apiClient,
	)

	// Routine: heartbeat every 10 seconds to API Server
	statusController := NewPodStatusController(
		kubelet,
		podService,
		apiClient,
		5*time.Second,
	)

	return &KubeletService{
		kubelet:           kubelet,
		podController:     podController,
		statusController:  statusController,
		serviceController: serviceController,
		apiClient:         apiClient,
	}
}

func (s *KubeletService) Run(stopCh <-chan struct{}) {
	err := s.apiClient.RegisterKubelet(s.kubelet)
	if err != nil {
		log.Printf("Failed to register kubelet: %v", err)
	}

	// Only once: restore local pods status.
	localPods, err := s.podController.podService.ListPods()
	if err != nil {
		log.Printf("Failed to fetch pods: %v", err)
	}

	// 清理 KubeProxy 本地状态
	s.serviceController.IpvsOps.Init()
	s.serviceController.IpvsOps.Clear()

	log.Printf("Restoring local pods: %v", utils.ExtractPodNames(localPods))
	s.kubelet.Mu.Lock()
	s.kubelet.Pods = localPods
	s.kubelet.Mu.Unlock()

	go s.podController.Run(stopCh)
	go s.statusController.Run(stopCh)
	go s.serviceController.Run(stopCh)
	<-stopCh
}
