package object

type PersistentVolumeClaimName struct {
	ClaimName string `yaml:"claimName" json:"claimName"` // PersistentVolumeClaim 名称
}

type HostPath struct {
	Path string `yaml:"path" json:"path"` // 主机路径
	Type string `yaml:"type" json:"type"` // 主机路径类型，如 Directory, File 等
}

type Volume struct {
	Name                      string                     `yaml:"name"`
	HostPath                  *HostPath                  `yaml:"hostPath,omitempty"`
	PersistentVolumeClaimName *PersistentVolumeClaimName `yaml:"persistentVolumeClaim,omitempty"`
}
