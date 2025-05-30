package main

import (
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/container"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubelet/runtime/pod"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/utils"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

// NOTE: Currently, we use USERNAME_HOSTNAME as the Kubelet name.
func getKubeletName() string {
	// Let's get the hostname of the current machine.
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}

	// Get the username of the current user.
	user, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get current user: %v", err)
	}

	// Combine the username and hostname to create a unique Kubelet name.
	return user.Username + "_" + hostname
}

func main() {
	controlPlaneIP := os.Getenv("APISERVER_URL")
	if controlPlaneIP == "" {
		log.Fatalf("APISERVER_URL environment variable must be set")
	}

	config := object.KubeletConfig{
		Name: getKubeletName(),
	}

	// 获取NodeIP
	if config.NodeIP == "" {
		ip, err := utils.GetEnInterfaceIP()
		if err != nil {
			log.Fatalf("failed to get NodeIP: %v", err)
		}

		config.NodeIP = ip
	}

	// 初始化 PodServiceInterface 的实际实现
	containerService, err := container.NewContainerService()
	if err != nil {
		log.Panicf("failed to create container service: %v", err)
	}

	podService := pod.NewPodService(containerService)

	// 使用环境变量创建 APIServerClient
	client := apiserver.NewAPIClient("http://" + controlPlaneIP + ":8080")

	kubeletService := kubelet.NewKubeletService(config, podService, client)

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
