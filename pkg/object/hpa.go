package object

import "time"

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
