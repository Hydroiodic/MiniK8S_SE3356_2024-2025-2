package object

import "fmt"

// PersistentVolume 定义精简的 PV 结构体
type PersistentVolume struct {
	Kind     string               `json:"kind"`
	Metadata Metadata             `json:"metadata"`
	Spec     PersistentVolumeSpec `json:"spec"`
	Status   string               `json:"status,omitempty"` // 状态字段，表示 PV 的当前状态
}

const (
	// PersistentVolumeAvailable 表示 PV 可用
	PersistentVolumeAvailable = "Available"
	// PersistentVolumeBound 表示 PV 已绑定
	PersistentVolumeBound = "Bound"
	// PersistentVolumeReleased 表示 PV 已释放
	PersistentVolumeReleased = "Released"
	// PersistentVolumeFailed 表示 PV 创建失败
	PersistentVolumeFailed = "Failed"
)

// PersistentVolumeSpec 定义 PV 规格
type PersistentVolumeSpec struct {
	Capacity                      ResourceList          `json:"capacity"`
	PersistentVolumeReclaimPolicy string                `json:"persistentVolumeReclaimPolicy"`
	NFS                           *NFSVolumeSource      `json:"nfs,omitempty"`
	HostPath                      *HostPathVolumeSource `json:"hostPath,omitempty"`
}

// TODO: 要不要呢？
const (
	// PersistentVolumeReclaimDelete 表示 PVC 删除时 PV 也删除
	PersistentVolumeReclaimDelete = "Delete"
	// PersistentVolumeReclaimRetain 表示 PV 保留
	PersistentVolumeReclaimRetain = "Retain"
)

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
	Capacity   ResourceList `json:"resources"`            // 期望的存储容量
	VolumeName string       `json:"volumeName,omitempty"` // 绑定的 PV 名称
}

// ResourceList 定义存储容量
type ResourceList struct {
	Storage string `json:"storage"` // 如 "5Gi"
}

func StorageToMegabytes(storage string) (int64, error) {
	// 解析存储容量字符串，转换为 MB
	var size int64
	_, err := fmt.Sscanf(storage, "%dGi", &size)

	if err != nil {
		return 0, fmt.Errorf("invalid storage format: %s", storage)
	}

	return size * 1024, nil // 转换为 MB
}
