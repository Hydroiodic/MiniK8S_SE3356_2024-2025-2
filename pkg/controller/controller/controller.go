package controller

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/function"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	ReplicasetController *(replicaset.ReplicasetController)
	HpaController        *(hpa.HPAController)
	JobController        *(gpu.JobController)
	FunctionController   *(function.FucntionController)
}

func (c *Controller) StartController() {
	c.JobController = gpu.NewJobController()
	c.FunctionController = function.NewFucntionController()

	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	go c.JobController.Start()
	go c.FunctionController.Start()
	select {}
}
