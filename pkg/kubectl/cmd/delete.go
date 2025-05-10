package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Kubernetes resource",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) < 2 {
			_, _ = fmt.Println(
				"Usage: minik8s kubectl delete <resource-type> <resource-name>",
			)
			return
		}

		resourceType := args[0]
		resourceName := args[1]

		switch resourceType {
		case PodResource:
			deletePod(resourceName)
		case "service":
			deleteService(resourceName)
		case "replicaset":
			deleteReplicaSet(resourceName)
		case "dns":
			deleteDNS(resourceName)
		case "hpa":
			deleteHPA(resourceName)
		default:
			_, _ = fmt.Println("Unsupported resource type:", resourceType)
		}
	},
}

// 在 init() 里注册 delete 命令.
func init() {
	kubectlCmd.AddCommand(deleteCmd)
}

// 删除 Pod.
func deletePod(name string) {
	// 这里添加实际删除 Pod 的逻辑
	_, _ = fmt.Printf("Deleting Pod: %s\n", name)
}

// 删除 Service.
func deleteService(name string) {
	// 这里添加实际删除 Service 的逻辑
	_, _ = fmt.Printf("Deleting Service: %s\n", name)
}

// 删除 ReplicaSet.
func deleteReplicaSet(name string) {
	// 这里添加实际删除 ReplicaSet 的逻辑
	_, _ = fmt.Printf("Deleting ReplicaSet: %s\n", name)
}

// 删除 DNS 配置.
func deleteDNS(name string) {
	// 这里添加实际删除 DNS 配置的逻辑
	_, _ = fmt.Printf("Deleting DNS config: %s\n", name)
}

// 删除 HPA配置.
func deleteHPA(name string) {
	// 这里添加实际删除 DNS 配置的逻辑
	_, _ = fmt.Printf("Deleting HPA config: %s\n", name)
}
