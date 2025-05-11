package image

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/docker"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/utils"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type ImageServiceInterface interface {
	PullImage(imageName string) error
	DeleteImage(imageName string) error
	ListImages() ([]image.Summary, error)
}

type ImageService struct {
	// Docker 客户端
	Cli *client.Client
}

func NewImageService() (*ImageService, error) {
	cli := docker.GetDockerClient()
	if cli == nil {
		return nil, fmt.Errorf("无法创建 Docker 客户端")
	}

	return &ImageService{
		Cli: cli,
	}, nil
}

// TODO: What if Image Already Exists?
// BUG: 镜像名称似乎不支持域名
func (is *ImageService) PullImage(imageName string) error {
	ctx := context.Background()
	out, err := is.Cli.ImagePull(ctx, imageName, image.PullOptions{})

	if err != nil {
		return fmt.Errorf("无法拉取镜像 %s: %v", imageName, err)
	}

	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("Failed to close image pull output: %v", err)
		}
	}()

	_, err = io.Copy(os.Stderr, out)

	if err != nil {
		return fmt.Errorf("读取镜像 %s 拉取输出失败: %v", imageName, err)
	}

	// 验证镜像存在
	images, err := is.Cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("列出镜像失败: %v", err)
	}

	repo, img, version := utils.ChunkImageName(imageName)

	// 对于 "docker.io/library/nginx:latest"，似乎并不会有前缀
	if repo == "docker.io/library" {
		imageName = img + ":" + version
	} else {
		imageName = fmt.Sprintf("%s/%s:%s", repo, img, version)
	}

	for _, img := range images {
		if slices.Contains(img.RepoTags, imageName) {
			log.Printf("镜像 %s 拉取成功", imageName)
			return nil
		}
	}

	return fmt.Errorf("镜像 %s 拉取后未找到", imageName)
}

// FIXME: Image Chain Dependency
func (is *ImageService) DeleteImage(imageName string) error {
	ctx := context.Background()
	_, err := is.Cli.ImageRemove(ctx, imageName, image.RemoveOptions{})

	if err != nil {
		return fmt.Errorf("无法删除镜像 %s: %v", imageName, err)
	}

	log.Printf("镜像 %s 删除成功", imageName)

	return nil
}

func (is *ImageService) ListImages() ([]image.Summary, error) {
	ctx := context.Background()
	images, err := is.Cli.ImageList(ctx, image.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("无法列出镜像: %v", err)
	}

	log.Printf("镜像列表: %v", images)

	return images, nil
}
