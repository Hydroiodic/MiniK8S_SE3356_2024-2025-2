package cmd

import (
	"fmt"
	"sort"
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
				"Usage: minik8s kubectl get <resource-type> [resource-name] [resource-namespace]",
			)
			return
		}

		resourceType := args[0]
		var resourceName string
		var resourceNamespace string
		if len(args) > 1 {
			resourceName = args[1]
		}
		resourceNamespace = "default"
		if len(args) > 2 {
			resourceNamespace = args[2]
		}
		switch strings.ToLower(resourceType) {
		case "pods", PodResource:
			if resourceName != "" {
				getPod(resourceName, resourceNamespace)
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
		case "hpas":
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
func getPod(name string, resourceNamespace string) string {
	_, _ = fmt.Printf("Getting Pod: %s,%s\n", name, resourceNamespace)
	// 这里添加实际获取单个 Pod 的逻辑
	return emptyReply
}

func printPods(pods []object.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		ageI := time.Since(pods[i].Status.StartTime)
		ageJ := time.Since(pods[j].Status.StartTime)

		return ageI < ageJ // 降序排列
	})
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
func getReplicaSet(name string) {
	rs, err := ci.GetReplicasetyName(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rsArr []object.ReplicaSet
	rsArr = append(rsArr, rs)
	PrintReplicaSetTable(rsArr)
}

func PrintReplicaSetTable(replicaSets []object.ReplicaSet) {
	// 表头
	header := "NAME\t\tNameSpace\tDESIRED\tCURRENT\tREADY\tCONTAINERS\tIMAGES\t\tSELECTOR"
	fmt.Println(header)

	// 每行数据
	for _, rs := range replicaSets {
		// 获取容器信息
		var containerNames []string

		var containerImages []string

		for _, c := range rs.Spec.Template.Spec.Containers {
			containerNames = append(containerNames, c.Name)
			containerImages = append(containerImages, c.Image)
		}

		// 格式化选择器
		var selectorParts []string
		for k, v := range rs.Spec.Selector {
			selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", k, v))
		}

		selector := strings.Join(selectorParts, ",")

		// 打印行数据
		line := fmt.Sprintf("%s\t%s\t\t%d\t%d\t%d\t%s\t%s\t%s",
			rs.Metadata.Name,
			rs.Metadata.Namespace,
			rs.Spec.Replicas,
			rs.Status.AvailableReplicas,
			rs.Status.AvailableReplicas,
			strings.Join(containerNames, ","),
			strings.Join(containerImages, ","),
			selector,
		)

		fmt.Println(line)
	}
}

func getAllReplicaSets() {
	rs, err := ci.GetReplicasets()
	if err != nil {
		fmt.Println(err)
		return
	}

	PrintReplicaSetTable(rs)
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

func getAllHPA() {
	hpas, err := ci.GetHpas()
	if err != nil {
		fmt.Println(err)
		return
	}

	printHPATable(hpas)
}

func printHPATable(hpas []object.HorizontalPodAutoscaler) {
	header := "NAME\t\tNameSpace\tREFERENCE\t\t\tTARGETS\t\t\t\tMINPODS\tMAXPODS"
	fmt.Println(header)

	for _, h := range hpas {
		// 获取目标引用
		ref := fmt.Sprintf(
			"%s/%s",
			h.Spec.ScaleTargetRef.Kind,
			h.Spec.ScaleTargetRef.Name,
		)

		// 格式化指标
		var targets []string

		for _, metric := range h.Spec.Metrics {
			if metric.Resource != nil {
				targets = append(targets, fmt.Sprintf("%s:%f%%",
					metric.Resource.Name,
					*metric.Resource.Target.AverageUtilization))
			}
		}
		// 获取目标引用

		line := fmt.Sprintf("%s\t%s\t\t%s\t%s\t\t%d\t%d\n",
			h.Metadata.Name,
			h.Metadata.Namespace,
			ref,
			strings.Join(targets, ","),
			h.Spec.MinReplicas,
			h.Spec.MaxReplicas,
		)

		fmt.Println(line)
	}
}
