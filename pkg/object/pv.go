package object

// PersistentVolume 定义精简的 PV 结构体
type PersistentVolume struct {
	Kind       string               `json:"kind"`
	APIVersion string               `json:"apiVersion"`
	Metadata   Metadata             `json:"metadata"`
	Spec       PersistentVolumeSpec `json:"spec"`
}

// PersistentVolumeSpec 定义 PV 规格
type PersistentVolumeSpec struct {
	Capacity                      ResourceList          `json:"capacity"`
	AccessModes                   []string              `json:"accessModes"`
	PersistentVolumeReclaimPolicy string                `json:"persistentVolumeReclaimPolicy"`
	StorageClassName              string                `json:"storageClassName,omitempty"`
	NFS                           *NFSVolumeSource      `json:"nfs,omitempty"`
	HostPath                      *HostPathVolumeSource `json:"hostPath,omitempty"`
}

// NFSVolumeSource 定义 NFS 存储
type NFSVolumeSource struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

// HostPathVolumeSource 定义 hostPath 存储
type HostPathVolumeSource struct {
	Path string `json:"path"`
}

// PersistentVolumeClaim 定义精简的 PVC 结构体
type PersistentVolumeClaim struct {
	Kind       string                    `json:"kind"`
	APIVersion string                    `json:"apiVersion"`
	Metadata   Metadata                  `json:"metadata"`
	Spec       PersistentVolumeClaimSpec `json:"spec"`
}

// ObjectMeta 定义元数据
type ObjectMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// PersistentVolumeClaimSpec 定义 PVC 规格
type PersistentVolumeClaimSpec struct {
	AccessModes      []string             `json:"accessModes"`
	Resources        ResourceRequirements `json:"resources"`
	StorageClassName string               `json:"storageClassName,omitempty"`
}

// ResourceRequirements 定义资源请求
type ResourceRequirements struct {
	Requests ResourceList `json:"requests"`
}

// ResourceList 定义存储容量
type ResourceList struct {
	Storage string `json:"storage"` // 如 "5Gi"
}
