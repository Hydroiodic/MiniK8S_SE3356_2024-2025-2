package object

type PersistentVolumeClaimName struct {
	ClaimName string `yaml:"claimName" json:"claimName"` // PersistentVolumeClaim 名称
}

type HostPath struct {
	Path string `yaml:"path" json:"path"` // 主机路径
	Type string `yaml:"type" json:"type"` // 主机路径类型，如 Directory, File 等
}

const (
	HostPathTypeDirectory = "Directory" // 主机路径类型：目录
	HostPathTypeFile      = "File"      // 主机路径类型：文件
)

type Volume struct {
	Name                  string                     `yaml:"name"`
	HostPath              *HostPath                  `yaml:"hostPath,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimName `yaml:"persistentVolumeClaim,omitempty"`
}
