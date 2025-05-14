package utils

import (
	"maps"
	"strings"
)

const PodNsNameLabelKey = "pod"
const IsPauseLabelKey = "pause"

// PodNsNameLabelKey 是用于标识 Pod 的命名空间和名称的标签键
// IsPauseLabelKey 是用于标识 Pause 容器的标签键
func GeneratePodNsNameLabel(
	podNs string,
	podName string,
) string {
	return podNs + "=" + podName
}

func ParsePodNsNameLabel(
	podNsNameLabel string,
) (string, string) {
	// 解析 Pod 的命名空间和名称
	parts := strings.Split(podNsNameLabel, "=")
	if len(parts) != 2 {
		return "", ""
	}

	podNs := parts[0]
	podName := parts[1]

	return podNs, podName
}

// Generate Labels for Container
// 打上 Namespace 和 Name 的 Label
// 打上 Pause 的 Tag
// 添加 Pod 的 Labels
func NewLabelForPauseContainer(
	podNs string,
	podName string,
	labels map[string]string,
) map[string]string {
	mergedLabels := map[string]string{
		PodNsNameLabelKey: podNs + "=" + podName,
		IsPauseLabelKey:   "true",
	}

	maps.Copy(mergedLabels, labels)

	return mergedLabels
}

func NewLabelForOtherContainer(
	podNs string,
	podName string,
	labels map[string]string,
) map[string]string {
	mergedLabels := map[string]string{
		PodNsNameLabelKey: podNs + "=" + podName,
		IsPauseLabelKey:   "false",
	}

	maps.Copy(mergedLabels, labels)

	return mergedLabels
}
