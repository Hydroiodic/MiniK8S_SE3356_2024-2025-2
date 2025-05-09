package utils

func FormatContainerName(
	podNs string,
	podName string,
	containerName string,
) string {
	return podNs + "-" + podName + "-" + containerName
}
