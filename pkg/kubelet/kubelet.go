package kubelet

import (
	"context"
	"fmt"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

// Kubelet持有一个全局唯一的Containerd Client实例

// TODO: 以Pod为单位进行管理？
// 一个Pod包含多个Container

func CreateContainer(
	ctx context.Context,
	client *containerd.Client,
	pod *object.Pod,
	containerSpec object.Container,
) error {
	// 设置 Containerd namespace（对应 Pod 的 namespace）
	ctx = namespaces.WithNamespace(ctx, pod.Metadata.Namespace)

	// 1. 拉取镜像
	image, err := client.Pull(
		ctx,
		containerSpec.Image,
		containerd.WithPullUnpack,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to pull image %s: %v",
			containerSpec.Image,
			err,
		)
	}

	// 2. 创建容器
	container, err := client.NewContainer(
		ctx,
		fmt.Sprintf("%s-%s", pod.Metadata.Name, containerSpec.Name), // 容器名称
		containerd.WithImage(image),
		containerd.WithNewSnapshot(
			fmt.Sprintf("snapshot-%s", containerSpec.Name),
			image,
		),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			// oci.WithHostNamespace(oci.NetworkNamespace), // 共享网络
			oci.WithEnv(
				[]string{
					"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				},
			),
			oci.WithProcessArgs(containerSpec.Command...), // 设置命令和参数
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	// 3. 启动容器
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %v", err)
	}

	err = task.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start task: %v", err)
	}

	return nil
}
