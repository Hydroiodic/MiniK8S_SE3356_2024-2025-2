package cmd

import (
	"fmt"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/spf13/cobra"
)

var ci *apiserver.APIClient
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Kubernetes resource",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) < 2 {
			_, _ = fmt.Println(
				"Usage: kubectl delete <resource-type> <resource-name> [resource-namespace]",
			)
			return
		}

		// TODO: 这变量名是啥玩意...
		resourceType := args[0]
		resourceName := args[1]
		resourceNamespace := "default" // Default value

		// TODO: 这样进行可选参数的处理有点丑陋...
		if len(args) >= 3 {
			resourceNamespace = args[2]
		}

		switch resourceType {
		case "pod":
			deletePod(resourceName, resourceNamespace)
		case "service":
			deleteService(resourceName, resourceNamespace)
		case "replicaset":
			deleteReplicaSet(resourceName)
		case "dns":
			deleteDNS(resourceName, resourceNamespace)
		case "hpa":
			deleteHPA(resourceName)
		default:
			_, _ = fmt.Println("Unsupported resource type:", resourceType)
		}
	},
}

// 在 init() 里注册 delete 命令.
func init() {
	// TODO: 能改成面向对象吗...
	rootCmd.AddCommand(deleteCmd)

	ci = apiserver.NewAPIClient("")
}

// 删除 Pod.
func deletePod(name, namespace string) {
	// 这里添加实际删除 Pod 的逻辑
	err := ci.DeletePodByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting pod:", err)
	}
}

// 删除 Service.
func deleteService(name, namespace string) {
	// 这里添加实际删除 Service 的逻辑
	err := ci.DeleteServiceByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting service:", err)
	}
}

// 删除 ReplicaSet.
func deleteReplicaSet(name string) {
	// 这里添加实际删除 ReplicaSet 的逻辑
	err := ci.DeleteReplicaset(name)
	if err != nil {
		fmt.Println("Error deleting replicaset:", err)
	}
}

// 删除 DNS 配置.
func deleteDNS(name, namespace string) {
	// 这里添加实际删除 DNS 配置的逻辑
	err := ci.DeleteDNSByName(name, namespace)
	if err != nil {
		fmt.Println("Error deleting DNS:", err)
		return
	}
}

// 删除 HPA配置.
func deleteHPA(name string) {
	// TODO: 这里添加实际删除 DNS 配置的逻辑
	_, _ = fmt.Printf("Deleting HPA config: %s\n", name)
}
