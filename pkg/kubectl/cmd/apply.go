package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	yaml2 "sigs.k8s.io/yaml"
)

// PodResource 代表 Pod 资源类型.
const PodResource = "Pod"

var execCmd = &cobra.Command{
	Use:   "apply",
	Short: "Interactive exec into a resource",
	Run: func(cmd *cobra.Command, _ []string) {
		fileFlag, _ := cmd.Flags().GetString("file") // 获取 `-f` 参数的值
		if fileFlag != "" {
			_, _ = fmt.Println("Using file:", fileFlag)
			parseYaml(fileFlag)
		} else {
			_, _ = fmt.Println("No file provided.")
		}
		// **手动重置 flag**
		if err := cmd.Flags().Lookup("file").Value.Set(""); err != nil {
			// 处理错误或打印日志
			_, _ = fmt.Println("Failed to set flag value:", err)
		}
	},
}

// 在 `init()` 里添加 `-f` flag.
func init() {
	execCmd.Flags().StringP("file", "f", "", "Specify the configuration file")
	kubectlCmd.AddCommand(execCmd)
}

// 统一解析并处理 YAML 资源的通用函数.
func parseAndHandle[T any](data []byte, handler func(*T)) error {
	var obj T
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("解析 YAML 失败: %w", err)
	}

	handler(&obj)

	return nil
}

func parseYaml(fileAddr string) {
	data, err := os.ReadFile(fileAddr) // #nosec G304
	if err != nil {
		_, _ = fmt.Println("读取文件失败:", err)
		return
	}

	var kindStruct struct {
		Kind string `yaml:"kind"`
	}

	marshalErr := yaml.Unmarshal(data, &kindStruct)
	if marshalErr != nil {
		_, _ = fmt.Println("解析 YAML 结构失败:", marshalErr)
		return
	}

	resourceHandlers := map[string]func([]byte) error{
		PodResource: func(rawData []byte) error {
			// 直接传递原始 JSON/YAML 数据，不解析成结构体
			return handlePodRaw(rawData)
		},
		"Service": func(rawData []byte) error {
			return handleServiceRaw(rawData)
		},
		"ReplicaSet": func(rawData []byte) error {
			return handleReplicaSetRaw(rawData)
		},
		"DNS": func(rawData []byte) error {
			return handleDNSConfigRaw(rawData)
		},
		"HorizontalPodAutoscaler": func(rawData []byte) error {
			return handleHPARaw(rawData)
		},
		// PodResource: func(data []byte) error {
		// 	return parseAndHandle[object.Pod](data, handlePod)
		// },
		// "Service": func(data []byte) error {
		// 	return parseAndHandle[object.Service](data, handleService)
		// },
		// "ReplicaSet": func(data []byte) error {
		// 	return parseAndHandle[object.ReplicaSet](data, handleReplicaSet)
		// },
		// "DNS": func(data []byte) error {
		// 	return parseAndHandle[object.DNS](data, handleDNSConfig)
		// },
		// "HorizontalPodAutoscaler": func(data []byte) error {
		// 	return parseAndHandle[object.HorizontalPodAutoscaler](
		// 		data,
		// 		handleHPA,
		// 	)
		// },
	}

	// 根据 kind 处理相应的资源
	if handler, exists := resourceHandlers[kindStruct.Kind]; exists {
		if err := handler(data); err != nil {
			_, _ = fmt.Println(err)
		}
	} else {
		_, _ = fmt.Println("不支持的资源类型:", kindStruct.Kind)
	}
}

func handlePodRaw(rawData []byte) error {
	// 1. 解析 YAML 到 map
	var data map[string]interface{}
	if err := yaml.Unmarshal(rawData, &data); err != nil {
		panic(err)
	}
	//获取当前时间
	currentTime := time.Now().UTC().Format(time.RFC3339)
	// 2. 添加 status 字段
	data["status"] = map[string]string{
		"phase":     "Running",
		"startTime": currentTime,
	}
	// 3. 重新生成 YAML
	newYAML, err := yaml.Marshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(newYAML))
	jsonData, err := yaml2.YAMLToJSON([]byte(newYAML))
	resp, err := http.Post( //nolint:gosec
		"http://localhost:8080/pod/createPod",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // 读取错误响应体
		return fmt.Errorf(
			"server returned %d: %s",
			resp.StatusCode,
			string(body),
		)
	}
	return nil
}

func handleServiceRaw(rawData []byte) error {
	fmt.Println("Raw Service JSON/YAML:", string(rawData))
	return nil
}

func handleReplicaSetRaw(rawData []byte) error {
	fmt.Println("Raw ReplicaSet JSON/YAML:", string(rawData))
	// 在这里可以直接操作 JSON，如提取字段、修改内容等
	return nil
}

func handleDNSConfigRaw(rawData []byte) error {
	fmt.Println("Raw DNS JSON/YAML:", string(rawData))
	return nil
}
func handleHPARaw(rawData []byte) error {
	fmt.Println("Raw HPA JSON/YAML:", string(rawData))
	return nil
}

func handlePod(pod *object.Pod) {
	_, _ = fmt.Println("pod apply" + pod.Kind)
}

func handleService(pod *object.Service) {
	_, _ = fmt.Println("service apply" + pod.Kind)
}

func handleReplicaSet(pod *object.ReplicaSet) {
	_, _ = fmt.Println("replicaset apply" + pod.Kind)
}

func handleDNSConfig(pod *object.DNS) {
	_, _ = fmt.Println("dns apply" + pod.Kind)
}
func handleHPA(pod *object.HorizontalPodAutoscaler) {
	_, _ = fmt.Println("hpa apply" + pod.Kind)
}
