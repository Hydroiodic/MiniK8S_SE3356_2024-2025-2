package volume

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type VolumeManager struct {
	apiClient *apiserver.APIClient
}

// NewVolumeManager 创建 VolumeManager 实例
func NewVolumeManager(apiClient *apiserver.APIClient) *VolumeManager {
	return &VolumeManager{
		apiClient: apiClient,
	}
}

// FormatPodDir 生成 Pod 的卷目录路径
func FormatPodDir(namespace, name string) string {
	return fmt.Sprintf("/tmp/kubelet/pods/%s-%s/volumes", namespace, name)
}

// MountVolumes 处理 Pod 的卷挂载，返回卷名称到节点挂载路径（Pod专属路径）的映射
func (vm *VolumeManager) MountVolumes(
	pod *object.Pod,
) (map[string]string, error) {
	volumePaths := make(map[string]string)
	podDir := FormatPodDir(pod.Metadata.Namespace, pod.Metadata.Name)

	// 创建 Pod 专属的卷目录
	err := os.MkdirAll(podDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create pod volume directory %s: %v",
			podDir,
			err,
		)
	}

	for _, volume := range pod.Spec.Volumes {
		// Pod专属的卷目录
		volumePath := filepath.Join(podDir, volume.Name)

		switch {
		case volume.HostPath != nil:
			// 处理直接的 HostPath 卷
			volumePaths[volume.Name] = volume.HostPath.Path

		case volume.PersistentVolumeClaim != nil:
			// 处理 PVC 卷，支持两种 PV
			err = vm.handlePVCVolume(
				pod.Metadata.Namespace,
				volume.PersistentVolumeClaim.ClaimName,
				volumePath,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to handle PVC volume %s: %v",
					volume.Name,
					err,
				)
			}

			volumePaths[volume.Name] = volumePath

		default:
			return nil, fmt.Errorf(
				"unsupported volume type for volume %s",
				volume.Name,
			)
		}
	}

	return volumePaths, nil
}

// handleHostPathVolume 处理直接的 HostPath 卷
func (vm *VolumeManager) handleHostPathVolume(
	hostPath, volumePath string,
) error {
	// 确保 HostPath 存在
	err := os.MkdirAll(hostPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create HostPath %s: %v", hostPath, err)
	}
	// 创建软链接
	err = os.Symlink(hostPath, volumePath)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf(
			"failed to create symlink %s -> %s: %v",
			volumePath,
			hostPath,
			err,
		)
	}

	return nil
}

// handlePVCVolume 处理 PVC 卷（NFS 或 HostPath PV）
func (vm *VolumeManager) handlePVCVolume(
	namespace, pvcName, volumePath string,
) error {
	// 获取 PVC
	pvc, err := vm.getPVC(namespace, pvcName)
	if err != nil {
		return fmt.Errorf(
			"failed to get PVC %s/%s: %v",
			namespace,
			pvcName,
			err,
		)
	}

	// 检查 PVC 是否绑定到 PV
	if pvc.Spec.VolumeName == "" {
		return fmt.Errorf(
			"PVC %s/%s is not bound to any PV",
			namespace,
			pvcName,
		)
	}

	// 获取 PV
	pv, err := vm.getPV(pvc.Spec.VolumeName)
	if err != nil {
		return fmt.Errorf(
			"failed to get PV %s for PVC %s/%s: %v",
			pvc.Spec.VolumeName,
			namespace,
			pvcName,
			err,
		)
	}

	if pv.Spec.NFS != nil && pv.Spec.NFS.Server != "" &&
		pv.Spec.NFS.Path != "" {
		// 如果目录已经存在，不再处理
		if _, err := os.Stat(volumePath); err == nil {
			log.Printf(
				"Volume path %s already exists, skipping NFS mount",
				volumePath,
			)
		} else if os.IsNotExist(err) {
			// 处理 NFS 类型的 PV
			err = os.MkdirAll(volumePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf(
					"failed to create volume path %s: %v",
					volumePath,
					err,
				)
			}
		}

		// TODO: 如果不行，就换成类似 symlink 的方式
		err = vm.mountNFSVolume(pv, volumePath)
		if err != nil {
			return fmt.Errorf(
				"failed to mount NFS volume %s: %v",
				pv.Metadata.Name,
				err,
			)
		}
	} else if pv.Spec.HostPath != nil && pv.Spec.HostPath.Path != "" {
		// 处理 HostPath 类型的 PV
		err = vm.handleHostPathVolume(pv.Spec.HostPath.Path, volumePath)
		if err != nil {
			return fmt.Errorf("failed to handle HostPath PV %s: %v", pv.Metadata.Name, err)
		}
	} else {
		return fmt.Errorf("unsupported PV type for PV %s", pv.Metadata.Name)
	}

	return nil
}

// UnmountVolumes 卸载 Pod 的所有卷
func (vm *VolumeManager) UnmountVolumes(pod *object.Pod) error {
	podDir := FormatPodDir(pod.Metadata.Namespace, pod.Metadata.Name)

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		// 处理 PVC 卷
		pvc, err := vm.getPVC(
			pod.Metadata.Namespace,
			volume.PersistentVolumeClaim.ClaimName,
		)
		if err != nil {
			log.Printf(
				"Failed to get PVC %s/%s: %v",
				pod.Metadata.Namespace,
				volume.PersistentVolumeClaim.ClaimName,
				err,
			)

			continue
		}

		pv, err := vm.getPV(pvc.Spec.VolumeName)
		if err != nil {
			log.Printf("Failed to get PV %s: %v", pvc.Spec.VolumeName, err)
			continue
		}

		if pv.Spec.NFS != nil && pv.Spec.NFS.Server != "" &&
			pv.Spec.NFS.Path != "" {
			// 卸载 NFS 类型的 PV
			volumePath := filepath.Join(podDir, volume.Name)

			err = vm.unmountNFSVolume(volumePath)
			if err != nil {
				log.Printf(
					"Failed to unmount NFS volume %s at %s: %v",
					pv.Metadata.Name,
					volumePath,
					err,
				)

				continue
			}
		} // HostPath 类型的 PV 和直接 HostPath 卷无需卸载，软链接在清理 podDir 时移除
	}

	// 清理 Pod 的卷目录，包括软链接和 NFS 挂载点
	err := os.RemoveAll(podDir)
	if err != nil {
		log.Printf("Failed to remove volume directory %s: %v", podDir, err)
	}

	return nil
}

// getPVC 通过 API Server 获取 PVC
func (vm *VolumeManager) getPVC(
	namespace, pvcName string,
) (*object.PersistentVolumeClaim, error) {
	pvc, err := vm.apiClient.GetPVC(namespace, pvcName)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PVC response: %v", err)
	}

	if pvc.Spec.VolumeName == "" {
		return nil, fmt.Errorf(
			"PVC %s/%s is not bound to any PV",
			namespace,
			pvcName,
		)
	}

	return pvc, nil
}

// getPV 通过 API Server 获取 PV
func (vm *VolumeManager) getPV(
	pvName string,
) (*object.PersistentVolume, error) {
	return vm.apiClient.GetPV(pvName)
}

// mountNFSVolume 执行 NFS 卷挂载
func (vm *VolumeManager) mountNFSVolume(
	pv *object.PersistentVolume,
	mountPoint string,
) error {
	// TODO: 将NFS.Path添加前缀，确保存在
	formattedNFSPath := filepath.Join(
		"/nfs_share",
		pv.Spec.NFS.Path,
	)

	cmd := exec.Command(
		"mount",
		"-t",
		"nfs",
		"-o",
		"soft,timeo=100,retry=3",
		pv.Spec.NFS.Server+":"+formattedNFSPath,
		mountPoint,
	)

	log.Printf(
		"Mounting NFS volume %s at %s with command: %s",
		pv.Metadata.Name,
		mountPoint,
		cmd.String(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to mount NFS volume %s: %v, output: %s",
			pv.Metadata.Name,
			err,
			string(output),
		)
	}

	err = os.Chmod(mountPoint, 0777)
	if err != nil {
		return fmt.Errorf(
			"failed to set permissions for %s: %v",
			mountPoint,
			err,
		)
	}

	return nil
}

// unmountNFSVolume 卸载 NFS 卷
func (vm *VolumeManager) unmountNFSVolume(mountPoint string) error {
	cmd := exec.Command("umount", mountPoint)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to unmount NFS volume at %s: %v, output: %s",
			mountPoint,
			err,
			string(output),
		)
	}

	return nil
}
