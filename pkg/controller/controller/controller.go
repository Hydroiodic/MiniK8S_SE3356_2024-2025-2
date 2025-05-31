package controller

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	ReplicasetController *(replicaset.ReplicasetController)
	HpaController        *(hpa.HPAController)
	JobController        *(gpu.JobController)
}

func (c *Controller) StartController() {
	c.JobController = gpu.NewJobController()
	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	go c.JobController.Start()
	select {}
}
