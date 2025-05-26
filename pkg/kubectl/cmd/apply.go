package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	client "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

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
	rootCmd.AddCommand(execCmd)
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
		"Pod": func(rawData []byte) error {
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
	// 获取当前时间
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

	var pod object.Pod
	if err = yaml.Unmarshal(newYAML, &pod); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	ci := client.NewAPIClient("http://localhost:8080")

	fmt.Println(pod)

	err = ci.CreatePod(&pod)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func handleServiceRaw(rawData []byte) error {
	fmt.Println("Raw Service JSON/YAML:", string(rawData))
	// 1. 解析 YAML 到 map
	var s object.Service
	if err := yaml.Unmarshal(rawData, &s); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	ci := client.NewAPIClient("http://localhost:8080")

	fmt.Println(s)

	err := ci.CreateService(&s)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func handleReplicaSetRaw(rawData []byte) error {
	// 1. 解析 YAML 到 map
	var r object.ReplicaSet
	if err := yaml.Unmarshal(rawData, &r); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	ci := client.NewAPIClient("http://localhost:8080")

	fmt.Println(r)

	err := ci.CreateReplicaset(&r)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func handleDNSConfigRaw(rawData []byte) error {
	// 1. 解析 YAML 到 map
	var DNS object.DNS
	if err := yaml.Unmarshal(rawData, &DNS); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
		return err
	}

	// Create a new API client.
	ci := client.NewAPIClient("")

	// Print the DNS configuration.
	fmt.Println(DNS)
	// Add the DNS configuration.
	err := ci.AddDNS(&DNS)
	if err != nil {
		log.Printf("error adding DNS: %v", err)
		return err
	}

	return nil
}

func handleHPARaw(rawData []byte) error {
	var hpa object.HorizontalPodAutoscaler
	if err := yaml.Unmarshal(rawData, &hpa); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	ci := client.NewAPIClient("http://localhost:8080")

	fmt.Println(hpa)

	err := ci.CreateHpa(&hpa)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
