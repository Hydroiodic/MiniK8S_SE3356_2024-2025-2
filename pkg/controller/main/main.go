package main

import (
	allcontroller "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/all-controller"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

func main() {
	c := &allcontroller.Controller{
		ReplicasetController: &replicaset.ReplicasetController{},
	}
	c.StartController()
}
