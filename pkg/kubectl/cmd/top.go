package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Display resource usage metrics",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) < 1 {
			_, _ = fmt.Println("Usage: minik8s kubectl top <pod|node> [<name>]")
			return
		}

		resourceType := args[0]
		var resourceName string
		if len(args) > 1 {
			resourceName = args[1]
		}

		switch strings.ToLower(resourceType) {
		case "pods", PodResource:
			if resourceName != "" {
				getPodMetrics(resourceName)
			} else {
				getAllPodMetrics()
			}
		default:
			_, _ = fmt.Println("Unsupported resource type:", resourceType)
		}
	},
}

func init() {
	kubectlCmd.AddCommand(topCmd)
}

// 获取单个 Pod 的资源指标.
func getPodMetrics(name string) string {
	_, _ = fmt.Printf("Metrics for Pod: %s\n", name)
	// 调用 cAdvisor API 获取 CPU/Memory 数据
	return ""
}

// 获取所有 Pod 的资源指标.
func getAllPodMetrics() string {
	_, _ = fmt.Println("Metrics for all Pods")
	// 批量获取指标并格式化输出
	return ""
}
