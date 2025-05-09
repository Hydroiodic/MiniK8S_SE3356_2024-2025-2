// pkg/cadvisor.go
package hpa

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type MetricsClient struct {
	Endpoint string // cAdvisor 地址（如 http://localhost:8080）
}

func NewClient() *MetricsClient {
	return &MetricsClient{Endpoint: "http://localhost:8080"}
}

// 获取 Pod 的 CPU/Memory 使用率
func (c *MetricsClient) GetPodMetric(podName, resource string) float64 {
	resp, err := http.Get(
		fmt.Sprintf("%s/api/v1.3/subcontainers/%s", c.Endpoint, podName),
	)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0
	}

	// 解析 CPU 使用率（单位：核）
	if resource == "cpu" {
		return data["cpu"].(map[string]interface{})["usage"].(float64) / 1e9 // 转换为秒
	}

	// 解析 Memory 使用量（单位：字节）
	if resource == "memory" {
		return data["memory"].(map[string]interface{})["usage"].(float64)
	}

	return 0
}
