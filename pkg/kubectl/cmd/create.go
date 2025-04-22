package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [api_object] [object_name]",
	Short: "Create Kubernetes resources (e.g., pods, deployments)",
	Args:  cobra.ExactArgs(2), // 必须接收两个参数
	Run: func(cmd *cobra.Command, args []string) {
		apiObject := args[0]  // 获取第一个参数
		objectName := args[1] // 获取第二个参数

		fmt.Printf("Creating %s named '%s'...\n", apiObject, objectName)

		switch apiObject {
		case "pod":
			fmt.Printf("Generating YAML for Pod '%s'...\n", objectName)
			fmt.Println("---")
			fmt.Printf(
				"apiVersion: v1\nkind: Pod\nmetadata:\n  name: %s\nspec:\n  containers:\n  - name: %s\n    image: nginx\n",
				objectName,
				objectName,
			)
		case "deployment":
			fmt.Printf("Generating YAML for Deployment '%s'...\n", objectName)
			fmt.Println("---")
			fmt.Printf(
				"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\nspec:\n  replicas: 1\n  template:\n    metadata:\n      labels:\n        app: %s\n    spec:\n      containers:\n      - name: %s-container\n        image: nginx\n",
				objectName,
				objectName,
				objectName,
			)
		default:
			fmt.Println(
				"Error: Unsupported API object. Use 'pod' or 'deployment'.",
			)
		}
	},
}

// 注册命令
func init() {
	rootCmd.AddCommand(createCmd)
}
