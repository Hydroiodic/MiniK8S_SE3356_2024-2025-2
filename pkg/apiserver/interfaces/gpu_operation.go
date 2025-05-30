package interfaces

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

const WorkDir = "/home/ubuntu/wang/MiniK8S_SE3356_2024-2025-2"

func GetGpuJobs(c *gin.Context) {
	// Create HpaStore and check for errors.
	st, err := object.NewGPUJOBStore([]string{})

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create GPUJOB store: "+err.Error(),
		)

		return
	}

	// Ensure the HpaStore is closed after use.
	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close GPUJOB store: %v\n", closeErr)
		}
	}()

	// List all HPA in etcd.
	gpujobs, err := st.ListGpuJobs(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to list GPUJOBS from etcd: "+err.Error(),
		)

		return
	}

	// Convert the HPA to a JSON format and return them.
	c.JSON(http.StatusOK, gpujobs)
}

//nolint:dupl
func CreateGpujob(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var job object.Job
	if err := c.BindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	st, err := object.NewGPUJOBStore([]string{})
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to create Gpu Job store: "+err.Error(),
		)

		return
	}

	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			fmt.Printf("Failed to close Job store: %v\n", closeErr)
		}
	}()

	reply, err := st.GetGpuJob(
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
		fmt.Println("Create Job from file failed: same Job namespace & name")
		c.JSON(
			http.StatusConflict,
			"Create Job from file failed: same Job namespace & name",
		)

		return
	}

	if err := st.AddGpuJob(c.Request.Context(), &job); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to add hpa to etcd: "+err.Error(),
		)

		return
	}
}

type RequestBody struct {
	JobID        string `json:"job_id"`
	JobName      string `json:"jobname"`
	JobNamespace string `json:"jobnamespace"`
	Output       string `json:"output,omitempty"`
	Error        string `json:"error,omitempty"`
}

func UpdateResult(c *gin.Context) {
	// Parse the JSON body into a hpa object.
	var body RequestBody
	if err := c.BindJSON(&body); err != nil {
		fmt.Println(err)
		return
	}

	// 定义目标文件夹和文件名
	folderPath := WorkDir + "/jobs/" + body.JobNamespace +
		"/" + body.JobName // 可以是相对路径或绝对路径（如 `/home/user/docs`）
	fileName := "ouput.txt"
	fullPath := filepath.Join(folderPath, fileName) // 跨平台路径拼接

	// 创建文件夹（如果不存在）
	err := os.MkdirAll(folderPath, 0755) // 0755 是目录权限（drwxr-xr-x）
	if err != nil {
		panic(err)
	}

	// 创建文件并写入内容
	file, err := os.Create(fullPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Failed to close file: %v\n", closeErr)
		}
	}()

	_, err = file.WriteString(body.Output)
	if err != nil {
		panic(err)
	}

	println("文件已写入:", fullPath)
}
