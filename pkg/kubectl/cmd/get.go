package cmd

import (
	"fmt"
	"strings"
	"time"

	client "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
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
	rootCmd.AddCommand(getCmd)
}

// Pod 相关操作.
func getPod(name string) string {
	_, _ = fmt.Printf("Getting Pod: %s\n", name)
	// 这里添加实际获取单个 Pod 的逻辑
	return emptyReply
}

func printPods(pods []object.Pod) {
	// 打印表头，增加了NAMESPACE和LABELS列
	fmt.Printf("%-30s %-15s %-10s %-10s %-10s %-30s\n",
		"NAME", "NAMESPACE", "READY", "STATUS", "AGE", "LABELS")

	for _, pod := range pods {
		ready := fmt.Sprintf(
			"%d/%d",
			len(pod.Spec.Containers),
			len(pod.Spec.Containers),
		)
		age := time.Since(pod.Status.StartTime).Truncate(time.Second)

		// 将labels map转换为字符串
		labels := ""
		for k, v := range pod.Metadata.Labels {
			if labels != "" {
				labels += ","
			}
			labels += fmt.Sprintf("%s=%s", k, v)
		}

		fmt.Printf(
			"%-30s %-15s %-10s %-10s %-10s %-30s\n",
			pod.Metadata.Name,
			pod.Metadata.Namespace, // 添加namespace
			ready,
			pod.Status.Phase,
			age,
			labels, // 添加labels
		)
	}
}

func getAllPods() {
	ci := client.NewAPIClient("http://localhost:8080")
	results, err := ci.GetPods()

	if err != nil {
		fmt.Println(err)
		return
	}

	printPods(results)
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
