package object

import "time"

type Pod struct {
	Kind     string    `yaml:"kind"`     // 固定为 "Pod"
	Metadata Metadata  `yaml:"metadata"` // 包含 name, namespace, labels
	Spec     PodSpec   `yaml:"spec"`     // 容器配置
	Status   PodStatus `yaml:"status"`   // 运行状态
}

type Metadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

type PodSpec struct {
	Containers []Container `yaml:"containers"` // 容器列表
	Volumes    []Volume    `yaml:"volumes"`    // 共享卷
}

type Container struct {
	Name      string          `yaml:"name"`
	Image     string          `yaml:"image"`     // 镜像名和 Tag，如 nginx:latest
	Command   []string        `yaml:"command"`   // 容器执行命令
	Args      []string        `yaml:"args"`      // 命令参数
	Ports     []ContainerPort `yaml:"ports"`     // 暴露端口
	Resources ResourceLimits  `yaml:"resources"` // 资源限制
}

type ContainerPort struct {
	ContainerPort int `yaml:"containerPort"`
}

type ResourceLimits struct {
	CPU    string `yaml:"cpu"`    // 如 "1"
	Memory string `yaml:"memory"` // 如 "128Mi"
}

type Volume struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"` // 挂载路径
}

type PodStatus struct {
	Phase      string    `yaml:"phase"`      // 运行状态：Pending, Running, Failed 等
	StartTime  time.Time `yaml:"startTime"`  // 启动时间
	Conditions []string  `yaml:"conditions"` // 状态条件
}
