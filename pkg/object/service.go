package object

type Service struct {
	Kind     string        `yaml:"kind"`     // 固定为 "Service"
	Type     string        `yaml:"type"`     // ClusterIP / NodePort
	Metadata Metadata      `yaml:"metadata"` // 名称、命名空间、标签
	Spec     ServiceSpec   `yaml:"spec"`     // 选择器、端口
	Status   ServiceStatus `yaml:"status"`   // 状态（IP、Endpoints）
}

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
