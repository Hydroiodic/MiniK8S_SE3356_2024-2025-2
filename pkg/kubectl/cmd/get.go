package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// emptyReply default.
const emptyReply = ""

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many Kubernetes resources",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) < 1 {
			_, _ = fmt.Println(
				"Usage: minik8s kubectl get <resource-type> [<resource-name>]",
			)
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
				getPod(resourceName)
			} else {
				getAllPods()
			}
		case "services", "service":
			if resourceName != "" {
				getService(resourceName)
			} else {
				getAllServices()
			}
		case "replicasets", "replicaset":
			if resourceName != "" {
				getReplicaSet(resourceName)
			} else {
				getAllReplicaSets()
			}
		case "dns":
			if resourceName != "" {
				getDNS(resourceName)
			} else {
				getAllDNS()
			}
		case "hpa":
			if resourceName != "" {
				getHPA(resourceName)
			} else {
				getAllHPA()
			}
		default:
			_, _ = fmt.Println("Unsupported resource type:", resourceType)
		}
	},
}

func init() {
	kubectlCmd.AddCommand(getCmd)
}

// Pod 相关操作.
func getPod(name string) string {
	_, _ = fmt.Printf("Getting Pod: %s\n", name)
	// 这里添加实际获取单个 Pod 的逻辑
	return emptyReply
}

func getAllPods() string {
	_, _ = fmt.Println("Listing all Pods")
	// 这里添加实际获取所有 Pod 的逻辑
	return emptyReply
}

// Service 相关操作.
func getService(name string) string {
	_, _ = fmt.Printf("Getting Service: %s\n", name)
	// 这里添加实际获取单个 Service 的逻辑
	return emptyReply
}

func getAllServices() string {
	_, _ = fmt.Println("Listing all Services")
	// 这里添加实际获取所有 Service 的逻辑
	return emptyReply
}

// ReplicaSet 相关操作.
func getReplicaSet(name string) string {
	_, _ = fmt.Printf("Getting ReplicaSet: %s\n", name)
	// 这里添加实际获取单个 ReplicaSet 的逻辑
	return emptyReply
}

func getAllReplicaSets() string {
	_, _ = fmt.Println("Listing all ReplicaSets")
	// 这里添加实际获取所有 ReplicaSet 的逻辑
	return emptyReply
}

// DNS 相关操作.
func getDNS(name string) string {
	_, _ = fmt.Printf("Getting DNS config: %s\n", name)
	// 这里添加实际获取单个 DNS 配置的逻辑
	return emptyReply
}

func getAllDNS() string {
	_, _ = fmt.Println("Listing all DNS configs")
	// 这里添加实际获取所有 DNS 配置的逻辑
	return emptyReply
}

// HPA 相关操作.
func getHPA(name string) string {
	_, _ = fmt.Printf("Getting HPA config: %s\n", name)
	// 这里添加实际获取单个 HPA 配置的逻辑
	return emptyReply
}

func getAllHPA() string {
	_, _ = fmt.Println("Listing all HPA configs")
	// 这里添加实际获取所有 HPA 配置的逻辑
	return emptyReply
}
