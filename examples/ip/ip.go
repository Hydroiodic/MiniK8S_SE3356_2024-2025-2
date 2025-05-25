package main

import "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/utils"

func main() {
	ip, err := utils.GetEnInterfaceIP()
	if err != nil {
		panic(err)
	}

	println("IP:", ip)
}
