package utils

import "strings"

func FormatContainerName(
	podNs string,
	podName string,
	containerName string,
) string {
	return podNs + "-" + podName + "-" + containerName
}

// 读取docker.io/library/nginx:latest 的尾部 nginx:latest
// ChunkImageName 解析镜像名称，返回仓库、镜像名称和标签
func ChunkImageName(imageName string) (string, string, string) {
	// 解析镜像名称
	parts := strings.Split(imageName, "/")
	if len(parts) == 0 {
		return "", "", ""
	}

	// 获取镜像名称和标签
	image := parts[len(parts)-1]
	tag := "latest"

	// 如果镜像名称中包含冒号，则分离出标签
	if strings.Contains(image, ":") {
		imageParts := strings.Split(image, ":")
		image = imageParts[0]
		tag = imageParts[1]
	}

	// 获取镜像仓库名称
	repo := strings.Join(parts[:len(parts)-1], "/")

	// 如果 repo 为空，设置为默认值（如 "docker.io/library"）
	if repo == "" {
		repo = "docker.io/library"
	}

	return repo, image, tag
}
