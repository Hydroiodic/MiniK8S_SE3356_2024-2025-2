package object

type Volume struct {
	Name string     `yaml:"name"` // 卷的名称
	Type VolumeType `yaml:"type"` // 卷的类型（hostPath, persistentVolumeClaim 等）
	Spec VolumeSpec `yaml:"spec"` // 卷的具体配置
}

// VolumeType 定义支持的卷类型
type VolumeType string

const (
	VolumeTypeHostPath              VolumeType = "hostPath"
	VolumeTypePersistentVolumeClaim VolumeType = "persistentVolumeClaim"
)

// VolumeSpec 定义不同类型卷的具体配置
type VolumeSpec struct {
	HostPath              *HostPathVolume              `yaml:"hostPath,omitempty"`              // hostPath 卷的配置
	PersistentVolumeClaim *PersistentVolumeClaimVolume `yaml:"persistentVolumeClaim,omitempty"` // PVC 卷的配置
}

// HostPathVolume 定义 hostPath 卷的配置
type HostPathVolume struct {
	Path string `yaml:"path"` // 主机上的路径
}

// PersistentVolumeClaimVolume 定义 PVC 卷的配置
type PersistentVolumeClaimVolume struct {
	ClaimName string `yaml:"claimName"` // PVC 的名称
	ReadOnly  bool   `yaml:"readOnly"`  // 是否只读
}
