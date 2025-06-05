package controller

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/function"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	ReplicasetController  *(replicaset.ReplicasetController)
	HpaController         *(hpa.HPAController)
	JobController         *(gpu.JobController)
	FunctionController    *(function.FucntionController)
	Serverless_controller *(function.Serverless_controller)
	eventController       *(function.EventTriggerController)
}

func (c *Controller) StartController() {
	c.JobController = gpu.NewJobController()
	c.FunctionController = function.NewFucntionController()
	c.Serverless_controller = function.NewServerlessController()
	c.eventController = function.NewEventController()
	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	go c.JobController.Start()
	go c.FunctionController.Start()
	go c.Serverless_controller.Start()
	go c.eventController.Start()
	select {}
}
