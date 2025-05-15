package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {
	// 初始化 KubeletConfig，根据需要修改
	config := object.KubeletConfig{
		Name: "myNode",
		// 其他字段...
	}

	// 初始化 PodServiceInterface 的实际实现
	containerService, err := container.NewContainerService()
	if err != nil {
		log.Panicf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	// 初始化 APIServerClient 的实际实现
	var apiClient kubelet.APIServerClient

	// 创建 KubeletService
	kubeletService := kubelet.NewKubeletService(config, podService, apiClient)

	// 创建一个 stopCh，用于优雅关闭服务
	stopCh := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Received stop signal, shutting down...")
		close(stopCh)
	}()

	// 运行 KubeletService
	log.Println("KubeletService is starting...")
	kubeletService.Run(stopCh)
	log.Println("KubeletService has stopped.")
}
