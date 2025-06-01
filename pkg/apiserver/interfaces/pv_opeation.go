package interfaces

import (
	"fmt"
	"net/http"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

// CreatePersistentVolume 创建新的 PersistentVolume
func CreatePersistentVolume(c *gin.Context) {
	// 解析请求体中的 JSON 到 PersistentVolume 对象
	var pv object.PersistentVolume
	if err := c.BindJSON(&pv); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 检查 PV 配置中的命名空间
	if pv.Metadata.Namespace == "" {
		pv.Metadata.Namespace = DefaultNamespace
	}

	// 创建 PersistentVolumeStore
	st, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 检查 etcd 中是否已存在同名 PV
	existingPV, err := st.GetPersistentVolume(
		c.Request.Context(),
		pv.Metadata.Name,
	)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 获取 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	if existingPV != nil {
		c.JSON(http.StatusConflict, "创建 PersistentVolume 失败: 名称已存在")
		return
	}

	// 设置 PV 的状态为可用
	pv.Status = object.PersistentVolumeAvailable

	// 将 PV 写入 etcd
	if err := st.AddPersistentVolume(c.Request.Context(), &pv); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"向 etcd 添加 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolume 创建成功: "+pv.Metadata.Name)
}

// UpdatePersistentVolume 更新现有的 PersistentVolume
func UpdatePersistentVolume(c *gin.Context) { //nolint
	// 解析请求体中的 JSON 到 PersistentVolume 对象
	var pv object.PersistentVolume
	if err := c.BindJSON(&pv); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 检查 PV 配置中的命名空间
	if pv.Metadata.Namespace == "" {
		pv.Metadata.Namespace = DefaultNamespace
	}

	// 创建 PersistentVolumeStore
	st, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 更新 etcd 中的 PV
	if err := st.UpdatePersistentVolume(c.Request.Context(), &pv); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"更新 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolume 更新成功: "+pv.Metadata.Name)
}

// GetPersistentVolume 获取指定的 PersistentVolume
func GetPersistentVolume(c *gin.Context) {
	// 从路径参数获取名称
	name := c.Param("name")

	// 创建 PersistentVolumeStore
	st, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 从 etcd 获取指定 PV
	pv, err := st.GetPersistentVolume(c.Request.Context(), name)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 获取 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	if pv == nil {
		c.JSON(http.StatusNotFound, "PersistentVolume 未找到: "+name)
		return
	}

	c.JSON(http.StatusOK, pv)
}

// ListPersistentVolumes 列出所有 PersistentVolume
func ListPersistentVolumes(c *gin.Context) {
	// 创建 PersistentVolumeStore
	st, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 列出 etcd 中的所有 PV
	pvs, err := st.ListPersistentVolumes(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 列出 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, pvs)
}

// DeletePersistentVolume 删除指定的 PersistentVolume
func DeletePersistentVolume(c *gin.Context) {
	// 解析请求体中的 JSON 到 PersistentVolume 对象
	var pv object.PersistentVolume
	if err := c.BindJSON(&pv); err != nil {
		c.JSON(http.StatusBadRequest, "无效的 JSON: "+err.Error())
		return
	}

	// 创建 PersistentVolumeStore
	st, err := object.NewPersistentVolumeStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"创建 PersistentVolume 存储失败: "+err.Error(),
		)

		return
	}

	// 确保使用后关闭 PersistentVolumeStore
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("关闭 PersistentVolume 存储失败: %v\n", closeErr)
		}
	}()

	// 从 etcd 删除 PV
	if err := st.DeletePersistentVolume(c.Request.Context(), pv.Metadata.Name); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"从 etcd 删除 PersistentVolume 失败: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "PersistentVolume 删除成功: "+pv.Metadata.Name)
}
