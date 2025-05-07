// cmd/autoscale.go
package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// minik8s kubectl autoscale rs my-rs 2 10
var autoscaleCmd = &cobra.Command{
	Use:   "autoscale",
	Short: "Manually trigger scaling for a resource",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			_, _ = fmt.Println(
				"Usage: minik8s kubectl autoscale <target-type> <target-name> <min> <max> [--cpu-percent=50]",
			)
			return
		}

		targetType := args[0]
		targetName := args[1]
		min, _ := strconv.Atoi(args[2])
		max, _ := strconv.Atoi(args[3])

		// 解析 Flags（如 CPU 目标利用率）
		cpuPercent, _ := cmd.Flags().GetInt("cpu-percent")

		// 调用 HPA 控制器逻辑
		triggerAutoscale(targetType, targetName, min, max, cpuPercent)
	},
}

func init() {
	autoscaleCmd.Flags().
		Int("cpu-percent", 80, "Target CPU utilization percentage")
	kubectlCmd.AddCommand(autoscaleCmd)
}

func triggerAutoscale(targetType, targetName string, min, max, cpuPercent int) {
	_, _ = fmt.Printf(
		"Trigger autoscale for %s/%s (min=%d, max=%d, cpu=%d%%)\n",
		targetType,
		targetName,
		min,
		max,
		cpuPercent,
	)
	// 调用 HPA 控制器的扩缩容逻辑
}
