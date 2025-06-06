package object

import "fmt"

type Service struct {
	Kind     string        `yaml:"kind"`     // 固定为 "Service"
	Type     string        `yaml:"type"`     // ClusterIP / NodePort
	Metadata Metadata      `yaml:"metadata"` // 名称、命名空间、标签
	Spec     ServiceSpec   `yaml:"spec"`     // 选择器、端口
	Status   ServiceStatus `yaml:"status"`   // 状态（IP、Endpoints）
}

const (
	SERVICE_TYPE_CLUSTERIP_STR = "ClusterIP"
	SERVICE_TYPE_NODEPORT_STR  = "NodePort"
	SERVICE_INTERNEL_LABEL     = "internal"
)

type ServiceSpec struct {
	Selector map[string]string `yaml:"selector"` // 匹配 Pod 的标签
	Ports    []ServicePort     `yaml:"ports"`    // 端口配置
}

type ServicePort struct {
	Name       string `yaml:"name"`       // 端口名称（可选）
	Port       int    `yaml:"port"`       // Service 监听的端口（集群内访问）
	TargetPort int    `yaml:"targetPort"` // Pod 实际服务端口
	NodePort   int    `yaml:"nodePort"`   // NodePort 方式暴露的端口（可选）
}

type ServiceStatus struct {
	ClusterIP string     `yaml:"clusterIP"` // Service 的 ClusterIP
	Endpoints []Endpoint `yaml:"endpoints"` // 后端 Pod 的 IP+Port
}

type Endpoint struct {
	IP   string `yaml:"ip"`   // Pod IP
	Port int    `yaml:"port"` // Pod Port
}

func (s *Service) CheckReady() bool {
	// If the ClusterIP is empty, the service is not ready.
	if s.Status.ClusterIP == "" {
		return false
	}

	// If the Endpoints are empty, the service is not ready.
	if len(s.Status.Endpoints) == 0 {
		return false
	}

	return true
}

func (s *Service) MatchLabels(
	podLabels map[string]string,
) bool {
	for key, svcValue := range s.Spec.Selector {
		// If the key is in the pod labels.
		if value, ok := podLabels[key]; ok {
			// If the value in the selector is not equal to the pod label.
			if value != svcValue {
				return false
			}
		} else {
			// If the key is not in the pod labels, return false.
			return false
		}
	}

	return true
}

func (s *Service) GetEndpoints() []string {
	endpoints := []string{}
	for _, endpoint := range s.Status.Endpoints {
		endpoints = append(
			endpoints,
			endpoint.IP+":"+fmt.Sprint(endpoint.Port),
		)
	}

	return endpoints
}

func (s *Service) GetPorts() []string {
	ports := []string{}

	for _, port := range s.Spec.Ports {
		if port.Port != 0 {
			ports = append(
				ports,
				fmt.Sprint(port.Port),
			)
		}
	}

	return ports
}

func (s *Service) GetTargetPorts() []string {
	targetPorts := []string{}

	for _, port := range s.Spec.Ports {
		if port.TargetPort != 0 {
			targetPorts = append(
				targetPorts,
				fmt.Sprint(port.TargetPort),
			)
		}
	}

	return targetPorts
}

func (s *Service) GetNodePorts() []string {
	nodePorts := []string{}

	for _, port := range s.Spec.Ports {
		if port.NodePort != 0 {
			nodePorts = append(
				nodePorts,
				fmt.Sprint(port.NodePort),
			)
		}
	}

	return nodePorts
}
