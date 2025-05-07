package cmd

import (
	"fmt"
	"os"

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
	kubectlCmd.AddCommand(execCmd)
}

func parseYaml(fileAddr string) {
	//读取文件内容
	data, err := os.ReadFile(fileAddr) // #nosec G304
	if err != nil {
		_, _ = fmt.Println("Error reading file:", err)
		return
	}

	//把文件解析成结构体
	// 定义一个临时结构体，只包含 Kind 字段
	var kindStruct struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal([]byte(data), &kindStruct); err != nil {
		_, _ = fmt.Println("error decoding YAML kind:", err)
		return
	}

	//判断yaml文件的类型
	switch kindStruct.Kind {
	case "Pod":
		var pod object.Pod
		if err := yaml.Unmarshal(data, &pod); err != nil {
			_, _ = fmt.Println("error parsing Pod YAML:", err)
			return
		}
		handlePod(&pod)
	case "Service":
		var svc object.Service
		if err := yaml.Unmarshal(data, &svc); err != nil {
			_, _ = fmt.Println("error parsing Pod YAML:", err)
			return
		}
		handleService(&svc)
	case "ReplicaSet":
		var rs object.ReplicaSet
		if err := yaml.Unmarshal(data, &rs); err != nil {
			_, _ = fmt.Println("error parsing ReplicaSet YAML:", err)
			return
		}
		handleReplicaSet(&rs)
	case "DNS":
		var dns object.DNS
		if err := yaml.Unmarshal(data, &dns); err != nil {
			_, _ = fmt.Println("error parsing DNS YAML:", err)
			return
		}
		handleDNSConfig(&dns)
	case "HorizontalPodAutoscaler":
		var hpa object.HorizontalPodAutoscaler
		if err := yaml.Unmarshal(data, &hpa); err != nil {
			_, _ = fmt.Println("error parsing DNS YAML:", err)
			return
		}
		handleHPA(&hpa)
	default:
		_, _ = fmt.Println("unsupported kind:", err)
		return
	}
	return
}

func handlePod(pod *object.Pod) {
	_, _ = fmt.Println("pod apply")
}

func handleService(pod *object.Service) {
	_, _ = fmt.Println("service apply")
}

func handleReplicaSet(pod *object.ReplicaSet) {
	_, _ = fmt.Println("replicaset apply")
}

func handleDNSConfig(pod *object.DNS) {
	_, _ = fmt.Println("dns apply")
}
func handleHPA(pod *object.HorizontalPodAutoscaler) {
	_, _ = fmt.Println("hpa apply")
}
