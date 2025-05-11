package object

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
