package object

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
	for key, value := range podLabels {
		// If the key is in the selector.
		if svcValue, ok := s.Spec.Selector[key]; ok {
			// If the value in the selector is not equal to the pod label.
			if value != svcValue {
				return false
			}
		} else {
			// If the key is not in the selector, return false.
			return false
		}
	}

	return true
}
