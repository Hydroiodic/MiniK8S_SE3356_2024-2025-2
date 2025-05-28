package interfaces

import (
	"fmt"
	"net/http"

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

	// 将 PVC 写入 etcd
	if err := st.AddPersistentVolumeClaim(c.Request.Context(), &pvc); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"向 etcd 添加 PersistentVolumeClaim 失败: "+err.Error(),
		)
		return
	}

	c.JSON(http.StatusOK, "PersistentVolumeClaim 创建成功: "+pvc.Metadata.Name)
}

// UpdatePersistentVolumeClaim 更新现有的 PersistentVolumeClaim
func UpdatePersistentVolumeClaim(c *gin.Context) {
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
func ListPersistentVolumeClaims(c *gin.Context) {
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
