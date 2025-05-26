package object

import (
	"time"

	"github.com/docker/go-connections/nat"
)

type Pod struct {
	Kind     string    `yaml:"kind"`     // 固定为 "Pod"
	Metadata Metadata  `yaml:"metadata"` // 包含 name, namespace, labels
	Spec     PodSpec   `yaml:"spec"`     // 容器配置
	Status   PodStatus `yaml:"status"`   // 禁止 YAML 解析
}

type Metadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

type PodSpec struct {
	PauseContainerID string      `yaml:"pauseContainerId"` // Pause 容器 ID（哈希值）
	RestartPolicy    string      `yaml:"restartPolicy"`    // 重启策略：Always, OnFailure, Never
	Containers       []Container `yaml:"containers"`       // 容器列表
	Volumes          []Volume    `yaml:"volumes"`          // 共享卷
}

type Container struct {
	ID        string         `yaml:"id"`        // 容器 ID，由 Docker 生成
	Name      string         `yaml:"name"`      // 容器名称
	Image     string         `yaml:"image"`     // 镜像名和 Tag，如 nginx:latest
	Command   []string       `yaml:"command"`   // 容器执行命令
	Args      []string       `yaml:"args"`      // 命令参数
	Ports     []int          `yaml:"ports"`     // 暴露端口
	Resources ResourceLimits `yaml:"resources"` // 资源限制

	Labels       map[string]string `yaml:"labels"` // 容器标签
	ExposedPorts nat.PortSet       `yaml:"-"`      // FIXME：这是干什么的？
	IP           string            `yaml:"ip"`     // TODO: 容器 IP 地址，能删除吗？
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
	IP         string    `yaml:"ip"`         // Pod IP 地址
}

const (
	PodUnknown  = "Unknown"
	PodCreating = "Creating"
	PodRunning  = "Running"
	PodDeleting = "Deleting"
	PodDeleted  = "Deleted"
	PodFailed   = "Failed"
)

type PodMetrics struct {
	Resources map[string]float64 `yaml:"-"`
}
