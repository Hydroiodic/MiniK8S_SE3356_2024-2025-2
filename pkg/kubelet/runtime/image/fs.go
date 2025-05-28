package image

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

func ReadFileFromImage(imageRef string, targetPath string) (string, error) {
	// Create a context for the operation.
	ctx := context.Background()

	// Parse source image.
	// NOTE: We use "docker-daemon:" transport to read from the local Docker daemon.
	//       This assumes that the image is available locally.
	srcRef, err := alltransports.ParseImageName("docker-daemon:" + imageRef)
	if err != nil {
		return "", err
	}

	// Create a system context for the image operations.
	sysCtx := &types.SystemContext{}

	// Set the system context to allow reading from the local Docker daemon.
	srcImg, err := srcRef.NewImageSource(ctx, sysCtx)
	if err != nil {
		return "", err
	}

	// Ensure the image source is closed after use.
	defer func() {
		if cerr := srcImg.Close(); cerr != nil {
			fmt.Printf("failed to close image source: %v\n", cerr)
		}
	}()

	// Create an image instance from the unparsed image.
	img, err := image.FromUnparsedImage(
		ctx,
		sysCtx,
		image.UnparsedInstance(srcImg, nil),
	)
	if err != nil {
		return "", err
	}

	// Traverse layers in order (oldest first).
	for _, layer := range img.LayerInfos() {
		// Get the layer reader.
		layerReader, _, err := srcImg.GetBlob(ctx, layer, nil)
		if err != nil {
			return "", err
		}

		// Ensure the layer reader is closed after use.
		defer func() {
			if cerr := layerReader.Close(); cerr != nil {
				fmt.Printf("failed to close layer reader: %v\n", cerr)
			}
		}()

		// Tar extract.
		tr := tar.NewReader(layerReader)

		for {
			// Read the next header from the tar archive.
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			// Other errors should be handled.
			if err != nil {
				return "", err
			}

			cleanPath := "/" + strings.TrimPrefix(hdr.Name, "./")
			if cleanPath == targetPath && hdr.Typeflag == tar.TypeReg {
				var buf bytes.Buffer
				_, err = io.Copy(&buf, tr)

				return buf.String(), err
			}
		}
	}

	// If we reach here, the target file was not found in any layer.
	// NOTE: We do not return an error here, as it may be valid for the file to not exist.
	return "", nil
}
