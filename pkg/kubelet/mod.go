package kubelet

import (
	"sync"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type KubeletConfig struct {
	ApiServerAddress string `yaml:"apiServerAddress" json:"apiServerAddress"`
	Name             string `yaml:"name"             json:"name"`
	Roles            string `yaml:"roles"            json:"roles"`
	Version          string `yaml:"version"          json:"version"`
	NodeIP           string `yaml:"nodeIP"           json:"nodeIP"`
}

type Kubelet struct {
	Config         KubeletConfig `yaml:"config"         json:"config"`
	StartTime      time.Time     `yaml:"startTime"      json:"startTime"`
	LastUpdateTime time.Time     `yaml:"lastUpdateTime" json:"lastUpdateTime"`
	Pods           []object.Pod  `yaml:"pods"           json:"pods"`
	CachedPods     []object.Pod  `yaml:"cachedPods"     json:"cachedPods"`
	mutex          sync.RWMutex
}

func NewKubelet(config KubeletConfig) *Kubelet {
	return &Kubelet{
		Config:         config,
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
		CachedPods:     []object.Pod{},
	}
}

type KubeletService struct {
	kubelet          *Kubelet
	podController    *PodController
	statusController *PodStatusController
}

func NewKubeletService(
	config KubeletConfig,
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
		10*time.Second,
	)
	return &KubeletService{
		kubelet:          kubelet,
		podController:    podController,
		statusController: statusController,
	}
}

func (s *KubeletService) Run(stopCh <-chan struct{}) {
	go s.podController.Run(stopCh)
	go s.statusController.Run(stopCh)
	<-stopCh
}
