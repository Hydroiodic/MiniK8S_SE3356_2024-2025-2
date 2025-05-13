package kubelet

import (
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
	Config KubeletConfig `yaml:"config"    json:"config"`
	// Status    string        `yaml:"status"    json:"status"`
	StartTime time.Time     `yaml:"startTime" json:"startTime"`
	Runtime   time.Duration `yaml:"runtime"   json:"runtime"` // FIXME: 这玩意有用吗？
	Pods      []object.Pod  `yaml:"pods"      json:"pods"`

	// 用于让api-server动态感知到kubelet的状态，记录每次发出心跳的时间
	LastUpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`
}

func NewKubelet(config KubeletConfig) *Kubelet {
	return &Kubelet{
		Config:         config,
		StartTime:      time.Now(),
		Runtime:        0,
		LastUpdateTime: time.Now(),
	}
}

type KubeletService struct {
	Data       *Kubelet
	podService pod.PodServiceInterface
}

// 从运行时获取当前节点上所有 Pod 的运行状态
func updateLocalPodStatus(kubelet *KubeletService) {
}

// TODO: 从API Server 获取完整的 Pod 信息

// TODO: 添加缺少，删除多余

func (k *KubeletService) SyncPods() {
	// 1. 从运行时获取当前节点上所有 Pod 的运行状态

	// 2. 从API Server 获取完整的 Pod 信息

	// 3. 添加缺少，删除多余
}
