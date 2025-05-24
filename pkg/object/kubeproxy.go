package object

import (
	"sync"
	"time"
)

type KubeProxyConfig struct {
	ApiServerAddress string `yaml:"apiServerAddress" json:"apiServerAddress"`
	Name             string `yaml:"name"             json:"name"`
	Roles            string `yaml:"roles"            json:"roles"`
	Version          string `yaml:"version"          json:"version"`
	NodeIP           string `yaml:"nodeIP"           json:"nodeIP"`
}

type KubeProxy struct {
	Config   KubeProxyConfig `yaml:"config"   json:"config"`
	Services []Service       `yaml:"services" json:"services"`

	// We need to record the last update time of the kubeproxy for heartbeat.
	LastUpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`

	// A mutex to protect the KubeProxy instance.
	Mu sync.Mutex
}

type KubeProxyCopy struct {
	Config   KubeProxyConfig `yaml:"config"   json:"config"`
	Services []Service       `yaml:"services" json:"services"`

	// We need to record the last update time of the kubeproxy for heartbeat.
	LastUpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`
}

func (k *KubeProxy) GetKubeProxyCopy() KubeProxyCopy {
	k.Mu.Lock()
	defer k.Mu.Unlock()

	return k.GetKubeProxyCopyWithoutLock()
}

func (k *KubeProxy) GetKubeProxyCopyWithoutLock() KubeProxyCopy {
	return KubeProxyCopy{
		Config:         k.Config,
		Services:       k.Services,
		LastUpdateTime: k.LastUpdateTime,
	}
}

func (k *KubeProxy) Heartbeat() {
	k.Mu.Lock()
	defer k.Mu.Unlock()

	k.LastUpdateTime = time.Now()
}
