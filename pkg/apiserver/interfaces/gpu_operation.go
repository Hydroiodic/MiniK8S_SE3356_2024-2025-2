package interfaces

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GetGPUJobs(c *gin.Context) {
	// Create GPUJobStore and check for errors.
	st, err := object.NewGPUJobStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create GPUJob store: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close GPUJob store: %v\n", closeErr)
		}
	}()

	// List all GPUJobs in etcd.
	GPUJobs, err := st.ListGPUJobs(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list GPUJobs from etcd: "+err.Error(),
		)

		return
	}

	// Convert the GPUJobs to a JSON format and return them.
	c.JSON(http.StatusOK, GPUJobs)
}

//nolint:dupl
func CreateGPUJob(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var job object.Job
	if err := c.BindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	// Create a new GPUJobStore and check for errors.
	st, err := object.NewGPUJobStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create GPUJobStore: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close GPUJobStore: %v\n", closeErr)
		}
	}()

	reply, err := st.GetGPUJob(
		c.Request.Context(),
		job.Metadata.Namespace,
		job.Metadata.Name,
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get Job from etcd: "+err.Error(),
		)

		return
	}

	if reply != nil {
		c.JSON(
			http.StatusConflict,
			"Create Job from file failed: same Job namespace & name",
		)

		return
	}

	if err := st.AddGPUJob(c.Request.Context(), &job); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add GPUJob to etcd: "+err.Error(),
		)

		return
	}
}

func UpdateResult(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var body object.JobRequestBody
	if err := c.BindJSON(&body); err != nil {
		fmt.Println(err)
		return
	}

	// 定义目标文件夹和文件名
	folderPath := gpu.WorkDir + "/jobs/" + body.JobNamespace +
		"/" + body.JobName // 可以是相对路径或绝对路径（如 `/home/user/docs`）
	fileName := "output.txt"
	fullPath := filepath.Join(folderPath, fileName) // 跨平台路径拼接

	// 创建文件夹（如果不存在）
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create directory: "+err.Error(),
		)

		return
	}

	// 创建文件并写入内容
	file, err := os.Create(fullPath)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create file: "+err.Error(),
		)

		return
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Failed to close file: %v\n", closeErr)
		}
	}()

	_, err = file.WriteString(body.Output)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to write to file: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, "File created successfully: "+fullPath)
}

func UpdateStatus(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var body object.JobRequestBody
	if err := c.BindJSON(&body); err != nil {
		fmt.Println(err)
		return
	}

	st, err := object.NewGPUJobStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create GPUJob store: "+err.Error(),
		)

		return
	}

	// Ensure the GPUJobStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			log.Printf("Failed to close GPUJob store: %v\n", closeErr)
		}
	}()

	// List all GPUJobs in etcd.
	GPUJobs, err := st.ListGPUJobs(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list GPUJobs from etcd: "+err.Error(),
		)

		return
	}

	for _, gpujob := range GPUJobs {
		if gpujob.Metadata.Name == body.JobName &&
			gpujob.Metadata.Namespace == body.JobNamespace {
			gpujob.Status = body.Status

			err = st.AddGPUJob(c.Request.Context(), gpujob)
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					"Failed to change status: "+err.Error(),
				)

				return
			}

			break
		}
	}

	c.JSON(http.StatusOK, "status updated successfully: ")
}
