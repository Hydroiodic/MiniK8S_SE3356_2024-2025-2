package main

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/kubeproxy/ipvs_ops"
)

func main() {
	ipvsOps := ipvs_ops.NewIpvsOps(ipvs_ops.CLUSTER_CIDR_DEFAULT)
	ipvsOps.Clear()
	ipvsOps.Init()

	defer ipvsOps.Close()

	apiClient := apiserver.NewAPIClient("")

	kubeProxy := kubeproxy.NewKubeProxy(
		ipvsOps,
		apiClient,
		10*time.Second,
	)

	stopCh := make(chan struct{})

	log.Printf("KubeProxy is starting...")
	kubeProxy.Run(stopCh)
}
