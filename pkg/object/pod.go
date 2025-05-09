package object

import (
	"encoding/json"
	"time"
)

var kindStruct struct {
	Kind string `yaml:"kind"`
}

// DNS 数据结构定义
type DNS struct {
	Kind     string   `yaml:"kind"`     // 固定为 "DNS"
	Metadata Metadata `yaml:"metadata"` // 名称、命名空间、标签
	Spec     DNSSpec  `yaml:"spec"`     // DNS 配置
}

// DNSSpec 表示 DNS 的配置
type DNSSpec struct {
	Host  string    `yaml:"host"`  // 域名主路径
	Paths []DNSPath `yaml:"paths"` // 子路径列表
}

// DNSPath 表示每个子路径的配置
type DNSPath struct {
	Path        string `yaml:"path"`        // 子路径地址
	ServiceName string `yaml:"serviceName"` // 对应的 Service 名称
	ServicePort int    `yaml:"servicePort"` // 对应的 Service 端口
}

// ReplicaSet 数据结构定义
type ReplicaSet struct {
	Kind     string           `yaml:"kind"`     // 固定为 "ReplicaSet"
	Metadata Metadata         `yaml:"metadata"` // 名称、命名空间、标签
	Spec     ReplicaSetSpec   `yaml:"spec"`     // 选择器、期望副本数、Pod 模板
	Status   ReplicaSetStatus `yaml:"status"`   // 当前状态（副本数）
}

// ReplicaSetSpec 表示 ReplicaSet 的期望状态
type ReplicaSetSpec struct {
	Selector map[string]string `yaml:"selector"` // 通过 Labels 筛选 Pod
	Replicas int               `yaml:"replicas"` // 期望的 Pod 副本数
	Template PodTemplateSpec   `yaml:"template"` // Pod 模板
}

// PodTemplateSpec 表示 ReplicaSet 中 Pod 的模板
type PodTemplateSpec struct {
	Metadata Metadata `yaml:"metadata"` // Pod 元数据（如名称、标签）
	Spec     PodSpec  `yaml:"spec"`     // Pod 的详细配置
}

// ReplicaSetStatus 表示 ReplicaSet 的当前状态
type ReplicaSetStatus struct {
	Replicas          int `yaml:"replicas"`          // 期望的 Pod 副本数
	AvailableReplicas int `yaml:"availableReplicas"` // 当前可用的 Pod 数量
}

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
	ID        string          `yaml:"id"`        // 容器 ID
	Name      string          `yaml:"name"`      // 容器名称
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

// HorizontalPodAutoscaler 数据结构定义
type HorizontalPodAutoscaler struct {
	Kind     string                        `yaml:"kind"`     // 固定为 "HorizontalPodAutoscaler"
	Metadata Metadata                      `yaml:"metadata"` // 名称、命名空间、标签
	Spec     HorizontalPodAutoscalerSpec   `yaml:"spec"`     // HPA 配置
	Status   HorizontalPodAutoscalerStatus `yaml:"status"`   // 当前状态
}

// HorizontalPodAutoscalerSpec 表示 HPA 的期望状态
type HorizontalPodAutoscalerSpec struct {
	ScaleTargetRef CrossVersionObjectReference      `yaml:"scaleTargetRef"`     // 目标工作负载
	MinReplicas    int32                            `yaml:"minReplicas"`        // 最小副本数
	MaxReplicas    int32                            `yaml:"maxReplicas"`        // 最大副本数
	Metrics        []MetricSpec                     `yaml:"metrics"`            // 监控指标
	Behavior       *HorizontalPodAutoscalerBehavior `yaml:"behavior,omitempty"` // 扩缩容策略（可选）
}

// CrossVersionObjectReference 表示可伸缩的目标工作负载
type CrossVersionObjectReference struct {
	Kind       string `yaml:"kind"`       // 类型：ReplicaSet/Deployment/Pod
	Name       string `yaml:"name"`       // 工作负载名称
	APIVersion string `yaml:"apiVersion"` // API 版本
}

// MetricSpec 定义监控指标
type MetricSpec struct {
	Type     string                `yaml:"type"`               // 指标类型："Resource"/"Pods"/"Object"
	Resource *ResourceMetricSource `yaml:"resource,omitempty"` // 资源指标（CPU/Memory等）
}

// ResourceMetricSource 定义资源指标
type ResourceMetricSource struct {
	Name   string       `yaml:"name"`   // 资源名称："cpu"/"memory"
	Target MetricTarget `yaml:"target"` // 目标值
}

// MetricTarget 定义指标目标
type MetricTarget struct {
	Type               string  `yaml:"type"`                         // 类型："Utilization"/"AverageValue"
	AverageUtilization *int32  `yaml:"averageUtilization,omitempty"` // 目标利用率（百分比）
	AverageValue       *string `yaml:"averageValue,omitempty"`       // 目标平均值（如 "100m"）
}

// HorizontalPodAutoscalerBehavior 定义扩缩容行为策略
type HorizontalPodAutoscalerBehavior struct {
	ScaleUp   *HPAScalingRules `yaml:"scaleUp,omitempty"`   // 扩容策略
	ScaleDown *HPAScalingRules `yaml:"scaleDown,omitempty"` // 缩容策略
}

// HPAScalingRules 定义扩缩容规则
type HPAScalingRules struct {
	StabilizationWindowSeconds *int32             `yaml:"stabilizationWindowSeconds,omitempty"` // 稳定窗口时间
	SelectPolicy               *string            `yaml:"selectPolicy,omitempty"`               // 策略选择方式
	Policies                   []HPAScalingPolicy `yaml:"policies"`                             // 策略列表
}

// HPAScalingPolicy 定义单个扩缩容策略
type HPAScalingPolicy struct {
	Type          string `yaml:"type"`          // 类型："Pods"/"Percent"
	Value         int32  `yaml:"value"`         // 调整值（如每次增减的Pod数量）
	PeriodSeconds int32  `yaml:"periodSeconds"` // 调整间隔（秒）
}

// HorizontalPodAutoscalerStatus 表示 HPA 的当前状态
type HorizontalPodAutoscalerStatus struct {
	ObservedGeneration *int64                             `yaml:"observedGeneration,omitempty"` // 观察到的生成版本
	LastScaleTime      *time.Time                         `yaml:"lastScaleTime,omitempty"`      // 最后扩缩容时间
	CurrentReplicas    int32                              `yaml:"currentReplicas"`              // 当前副本数
	DesiredReplicas    int32                              `yaml:"desiredReplicas"`              // 期望副本数
	CurrentMetrics     []MetricStatus                     `yaml:"currentMetrics"`               // 当前指标值
	Conditions         []HorizontalPodAutoscalerCondition `yaml:"conditions"`                   // 状态条件
}

// MetricStatus 表示当前指标状态
type MetricStatus struct {
	Type     string                `yaml:"type"`
	Resource *ResourceMetricStatus `yaml:"resource,omitempty"`
}

// ResourceMetricStatus 表示资源指标状态
type ResourceMetricStatus struct {
	Name    string            `yaml:"name"`
	Current MetricValueStatus `yaml:"current"`
}

// MetricValueStatus 表示指标值状态
type MetricValueStatus struct {
	AverageUtilization *int32 `yaml:"averageUtilization,omitempty"`
	AverageValue       string `yaml:"averageValue"`
}

// HorizontalPodAutoscalerCondition 表示 HPA 状态条件
type HorizontalPodAutoscalerCondition struct {
	Type    string `yaml:"type"`
	Status  string `yaml:"status"`
	Reason  string `yaml:"reason"`
	Message string `yaml:"message"`
}

func (c *Container) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func ContainerFromJSON(data string) (*Container, error) {
	var c Container

	err := json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
