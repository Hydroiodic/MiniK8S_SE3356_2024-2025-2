package object

import (
	"sync"
	"time"
)

type KubeletConfig struct {
	ApiServerAddress string `yaml:"apiServerAddress" json:"apiServerAddress"`
	Name             string `yaml:"name"             json:"name"`
	Roles            string `yaml:"roles"            json:"roles"`
	Version          string `yaml:"version"          json:"version"`
	NodeIP           string `yaml:"nodeIP"           json:"nodeIP"`
}

type Kubelet struct {
	Config    KubeletConfig `yaml:"config"    json:"config"`
	Status    string        `yaml:"status"    json:"status"`
	StartTime time.Time     `yaml:"startTime" json:"startTime"`
	Runtime   time.Duration `yaml:"runtime"   json:"runtime"`
	Pods      []Pod         `yaml:"pods"      json:"pods"`

	// We need to record the last update time of the kubelet for heartbeat.
	LastUpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`

	// A mutex to protect the kubelet instance.
	Mu sync.Mutex
}

type KubeletCopy struct {
	Config    KubeletConfig `yaml:"config"    json:"config"`
	Status    string        `yaml:"status"    json:"status"`
	StartTime time.Time     `yaml:"startTime" json:"startTime"`
	Runtime   time.Duration `yaml:"runtime"   json:"runtime"`
	Pods      []Pod         `yaml:"pods"      json:"pods"`

	// We need to record the last update time of the kubelet for heartbeat.
	LastUpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`
}

func (k *Kubelet) GetKubeletCopy() KubeletCopy {
	k.Mu.Lock()
	defer k.Mu.Unlock()

	return k.GetKubeletCopyWithoutLock()
}

func (k *Kubelet) GetKubeletCopyWithoutLock() KubeletCopy {
	return KubeletCopy{
		Config:         k.Config,
		Status:         k.Status,
		StartTime:      k.StartTime,
		Runtime:        k.Runtime,
		Pods:           k.Pods,
		LastUpdateTime: k.LastUpdateTime,
	}
}

func (k *Kubelet) Heartbeat() {
	k.Mu.Lock()
	defer k.Mu.Unlock()

	k.LastUpdateTime = time.Now()
}
