package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
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
	// Parse the raw YAML data into a Pod object.
	var pod object.Pod
	if err := yaml.Unmarshal(rawData, &pod); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	// Add the Pod configuration.
	err := apiserver.NewAPIClient("").CreatePod(&pod)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func handleServiceRaw(rawData []byte) error {
	// Parse the raw YAML data into a Service object.
	var s object.Service
	if err := yaml.Unmarshal(rawData, &s); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	// Add the Service configuration.
	err := apiserver.NewAPIClient("").CreateService(&s)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func handleReplicaSetRaw(rawData []byte) error {
	// Parse the raw YAML data into a ReplicaSet object.
	var r object.ReplicaSet
	if err := yaml.Unmarshal(rawData, &r); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	// Add the ReplicaSet configuration.
	err := apiserver.NewAPIClient("").CreateReplicaset(&r)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func handleDNSConfigRaw(rawData []byte) error {
	// Parse the raw YAML data into a DNS object.
	var DNS object.DNS
	if err := yaml.Unmarshal(rawData, &DNS); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	// Add the DNS configuration.
	err := apiserver.NewAPIClient("").AddDNS(&DNS)
	if err != nil {
		log.Printf("error adding DNS: %v", err)
		return err
	}

	return nil
}

func handleHPARaw(rawData []byte) error {
	// Parse the raw YAML data into a HorizontalPodAutoscaler object.
	var hpa object.HorizontalPodAutoscaler
	if err := yaml.Unmarshal(rawData, &hpa); err != nil {
		log.Fatalf("error unmarshaling YAML: %v", err)
	}

	// Add the HorizontalPodAutoscaler configuration.
	err := apiserver.NewAPIClient("").CreateHpa(&hpa)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
