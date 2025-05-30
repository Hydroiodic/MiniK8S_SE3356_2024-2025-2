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
			_, _ = fmt.Println("Usage: kubectl top <pod|node> [<name>]")
			return
		}

		resourceType := args[0]
		var resourceName string
		if len(args) > 1 {
			resourceName = args[1]
		}

		switch strings.ToLower(resourceType) {
		case PodCmdName, PodCmdName + "s":
			if resourceName != "" {
				getPodMetrics(resourceName)
			} else {
				getAllPodMetrics()
			}
		default:
			fmt.Println("Unsupported resource type:", resourceType)
		}
	},
}

func init() {
	rootCmd.AddCommand(topCmd)
}

// 获取单个 Pod 的资源指标.
func getPodMetrics(name string) string {
	fmt.Printf("Metrics for Pod: %s\n", name)
	// 调用 cAdvisor API 获取 CPU/Memory 数据
	return ""
}

// 获取所有 Pod 的资源指标.
func getAllPodMetrics() string {
	fmt.Println("Metrics for all Pods")
	// 批量获取指标并格式化输出
	return ""
}
