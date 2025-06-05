package interfaces

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

// CreatePersistentVolumeClaim 创建新的 PersistentVolumeClaim
func CreatePersistentVolumeClaim(c *gin.Context) {
	// 解析请求体中的 JSON 到 PersistentVolumeClaim 对象
	var pvc object.PersistentVolumeClaim
	if err := c.BindJSON(&pvc); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 检查 PVC 配置中的命名空间
	if pvc.Metadata.Namespace == "" {
		pvc.Metadata.Namespace = DefaultNamespace
	}

	// 创建 PersistentVolumeClaimStore
	st, err := object.NewPersistentVolumeClaimStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolumeClaim 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeClaimStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolumeClaim 存储失败: %v\n", closeErr)
		}
	}()

	// 检查 etcd 中是否已存在同名 PVC
	existingPVC, err := st.GetPersistentVolumeClaim(
		c.Request.Context(),
		pvc.Metadata.Namespace,
		pvc.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 获取 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	if existingPVC != nil {
		c.JSON(http.StatusConflict, "创建 PersistentVolumeClaim 失败: 命名空间和名称已存在")
		return
	}

	// 创建 PersistentVolumeStore
	pvSt, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := pvSt.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// TODO: 检查条件，试图和已有的PV绑定，如果不存在，创建PV
	volumes, err := pvSt.ListPersistentVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			"从 etcd 列出 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	requiredCapacity, err := object.StorageToMegabytes(
		pvc.Spec.Capacity.Storage,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, "无效的存储容量格式: "+err.Error())
		return
	}

	found := false

	// 将 PVC 绑定到第一个可用的 PV
	for _, pv := range volumes {
		if pv.Status != object.PersistentVolumeAvailable {
			continue // 跳过不可用的 PV
		}

		pvCapacity, err := object.StorageToMegabytes(pv.Spec.Capacity.Storage)
		if err != nil {
			c.JSON(http.StatusBadRequest, "无效的 PV 存储容量格式: "+err.Error())
			return
		}

		// 检查 PV 的存储容量是否满足 PVC 的要求
		if pvCapacity < requiredCapacity {
			continue // 跳过容量不足的 PV
		}

		found = true // 找到一个合适的 PV

		// 如果 PV 的存储容量大于等于 PVC 的要求，则进行绑定
		// 绑定 PVC 到 PV
		pvc.Spec.VolumeName = pv.Metadata.Name
		pv.Status = object.PersistentVolumeBound // 更新 PV 状态为已绑定

		// 更新 PV 的状态
		if err := pvSt.UpdatePersistentVolume(context.Background(), pv); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"更新 PersistentVolume 状态失败: "+err.Error(),
			)

			return
		}

		break // 找到一个合适的 PV 后退出循环
	}

	// 检查是否有可用的 PV
	if !found {
		// TODO: 自动创建NFS-PV
		nfsPath := "/autogen/" + pvc.Metadata.Namespace + "-" + pvc.Metadata.Name // TODO: 替换为实际的 NFS 导出路径
		pv := object.PersistentVolume{
			Metadata: object.Metadata{
				Name:      pvc.Metadata.Name + "-pv",
				Namespace: pvc.Metadata.Namespace,
			},
			Spec: object.PersistentVolumeSpec{
				Capacity: pvc.Spec.Capacity,
				NFS: &object.NFSVolumeSource{
					Server: os.Getenv("APISERVER_URL"),
					Path:   nfsPath, // 使用自动生成的 NFS 路径
				},
			},
			Status: object.PersistentVolumeAvailable, // 设置 PV 状态为可用
		}

		formattedNFSPath := filepath.Join(
			"/nfs_share",
			pv.Spec.NFS.Path,
		)

		// 分配文件夹
		if err := os.MkdirAll(formattedNFSPath, os.ModePerm); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"创建 NFS 目录失败: "+err.Error(),
			)

			return
		}

		// 将新创建的 PV 写入 etcd
		if err := pvSt.AddPersistentVolume(context.Background(), &pv); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"向 etcd 添加 PersistentVolume 失败: "+err.Error(),
			)

			return
		}

		// 将 PVC 绑定到新创建的 PV
		pvc.Spec.VolumeName = pv.Metadata.Name
		pv.Status = object.PersistentVolumeBound // 更新 PV 状态为已绑定

		if err := pvSt.UpdatePersistentVolume(context.Background(), &pv); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"更新 PersistentVolume 状态失败: "+err.Error(),
			)

			return
		}
	}

	// 将 PVC 写入 etcd
	if err := st.AddPersistentVolumeClaim(context.Background(), &pvc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"向 etcd 添加 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolumeClaim 创建成功: "+pvc.Metadata.Name)
}

// UpdatePersistentVolumeClaim 更新现有的 PersistentVolumeClaim
func UpdatePersistentVolumeClaim(c *gin.Context) { //nolint
	// 解析请求体中的 JSON 到 PersistentVolumeClaim 对象
	var pvc object.PersistentVolumeClaim
	if err := c.BindJSON(&pvc); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 检查 PVC 配置中的命名空间
	if pvc.Metadata.Namespace == "" {
		pvc.Metadata.Namespace = DefaultNamespace
	}

	// 创建 PersistentVolumeClaimStore
	st, err := object.NewPersistentVolumeClaimStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolumeClaim 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeClaimStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolumeClaim 存储失败: %v\n", closeErr)
		}
	}()

	// 更新 etcd 中的 PVC
	if err := st.UpdatePersistentVolumeClaim(c.Request.Context(), &pvc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"更新 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolumeClaim 更新成功: "+pvc.Metadata.Name)
}

// GetPersistentVolumeClaim 获取指定的 PersistentVolumeClaim
func GetPersistentVolumeClaim(c *gin.Context) {
	// 从路径参数获取命名空间和名称
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" {
		namespace = DefaultNamespace
	}

	// 创建 PersistentVolumeClaimStore
	st, err := object.NewPersistentVolumeClaimStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolumeClaim 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeClaimStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolumeClaim 存储失败: %v\n", closeErr)
		}
	}()

	// 从 etcd 获取指定 PVC
	pvc, err := st.GetPersistentVolumeClaim(
		c.Request.Context(),
		namespace,
		name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 获取 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	if pvc == nil {
		c.JSON(http.StatusNotFound, "PersistentVolumeClaim 未找到: "+name)
		return
	}

	c.JSON(http.StatusOK, pvc)
}

// ListPersistentVolumeClaims 列出所有 PersistentVolumeClaim
func GetPersistentVolumeClaims(c *gin.Context) {
	// 创建 PersistentVolumeClaimStore
	st, err := object.NewPersistentVolumeClaimStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolumeClaim 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeClaimStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolumeClaim 存储失败: %v\n", closeErr)
		}
	}()

	// 列出 etcd 中的所有 PVC
	pvcs, err := st.ListPersistentVolumeClaims(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 列出 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, pvcs)
}

// DeletePersistentVolumeClaim 删除指定的 PersistentVolumeClaim
func DeletePersistentVolumeClaim(c *gin.Context) {
	// 解析请求体中的 JSON 到 PersistentVolumeClaim 对象
	var pvc object.PersistentVolumeClaim
	if err := c.BindJSON(&pvc); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 创建 PersistentVolumeClaimStore
	st, err := object.NewPersistentVolumeClaimStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolumeClaim 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeClaimStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolumeClaim 存储失败: %v\n", closeErr)
		}
	}()

	// 解绑PV，保留数据
	// 更新PV的状态为可用
	if pvc.Spec.VolumeName != "" {
		err := changePVState(
			c.Request.Context(),
			object.PersistentVolumeAvailable,
			pvc.Spec.VolumeName,
		)
		if err != nil {
			c.JSON(
				http.StatusInternalServerError,
				"更新 PersistentVolume 状态失败: "+err.Error(),
			)

			return
		} // TODO: 是否删除PV内部包含的数据？
	}

	// 从 etcd 删除 PVC
	if err := st.DeletePersistentVolumeClaim(c.Request.Context(), pvc.Metadata.Namespace, pvc.Metadata.Name); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 删除 PersistentVolumeClaim 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolumeClaim 删除成功: "+pvc.Metadata.Name)
}

// TODO: 这一块逻辑不好，包括了Store的创建和删除，有一定副作用
func changePVState(ctx context.Context, state string, volumeName string) error {
	// 创建 PersistentVolumeStore
	pvSt, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		return fmt.Errorf("创建 PersistentVolume 存储失败: %w", err)
	}
	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := pvSt.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 获取指定的 PV
	pv, err := pvSt.GetPersistentVolume(
		ctx,
		volumeName,
	)

	if err != nil {
		return fmt.Errorf("从 etcd 获取 PersistentVolume 失败: %w", err)
	}

	if pv == nil {
		return fmt.Errorf("PersistentVolume 未找到: %s", volumeName)
	}

	// 更新 PV 的状态为可用
	pv.Status = state
	if err := pvSt.UpdatePersistentVolume(ctx, pv); err != nil {
		return fmt.Errorf("更新 PersistentVolume 状态失败: %w", err)
	}

	return nil
}
