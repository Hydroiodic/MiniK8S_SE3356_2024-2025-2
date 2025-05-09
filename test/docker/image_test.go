package docker_test

import (
	"testing"

	// "github.com/docker/docker/api/types/image"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/docker"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/image"
	"github.com/stretchr/testify/assert"
)

func TestPullImage(t *testing.T) {

	imgService := &image.ImageService{
		Cli: docker.GetDockerClient(),
	}

	err := imgService.PullImage("nginx:latest")

	// Check if the image is pulled successfully
	assert.NoError(t, err)

	// List images to verify the image is present
	images, err := imgService.ListImages()

	// Check if the image is listed
	for _, img := range images {
		t.Logf("Image ID: %s, RepoTags: %v", img.ID, img.RepoTags)

		if len(img.RepoTags) > 0 && img.RepoTags[0] == "nginx:latest" {
			assert.NoError(t, err)
			return
		}
	}

	assert.Fail(t, "Image not found in the list")
}

func TestDeleteImage(t *testing.T) {
	imgService := &image.ImageService{
		Cli: docker.GetDockerClient(),
	}

	err := imgService.DeleteImage("nginx:latest")

	// Check if the image is deleted successfully
	assert.NoError(t, err)

	// List images to verify the image is deleted
	images, err := imgService.ListImages()

	// Check if the image is not listed
	for _, img := range images {
		if img.RepoTags[0] == "nginx:latest" {
			assert.Fail(t, "Image still exists after deletion")
			return
		}
	}

	assert.NoError(t, err)
}

func TestListImages(t *testing.T) {
	imgService := &image.ImageService{
		Cli: docker.GetDockerClient(),
	}

	images, err := imgService.ListImages()

	// Check if the images are listed successfully
	assert.NoError(t, err)

	// Check if the list is not empty
	assert.NotEmpty(t, images)

	// Print the image IDs and RepoTags
	for _, img := range images {
		t.Logf("Image ID: %s, RepoTags: %v", img.ID, img.RepoTags)
	}
}
