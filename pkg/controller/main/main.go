package main

import (
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	allcontroller "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/all-controller"
	hpa "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

func main() {

	c := &allcontroller.Controller{
		ReplicasetController: &replicaset.ReplicasetController{},
		HpaController: &hpa.HPAController{
			HpasMap: make(map[string]*object.HorizontalPodAutoscaler),
			Tickers: make(map[string]*time.Ticker),
			QuitChs: make(map[string]chan struct{}),
			Ci:      apiserver.NewAPIClient(""), // 假设有一个NewAPIClient函数
		},
	}
	c.StartController()
}
