package image

import (
	"context"
	"fmt"
	"log"

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

// TODO: What if Image Already Exists?
func (is *ImageService) PullImage(imageName string) error {
	ctx := context.Background()
	_, err := is.Cli.ImagePull(ctx, imageName, image.PullOptions{})

	if err != nil {
		return fmt.Errorf("无法拉取镜像 %s: %v", imageName, err)
	}

	log.Printf("镜像 %s 拉取成功", imageName)

	return nil
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
