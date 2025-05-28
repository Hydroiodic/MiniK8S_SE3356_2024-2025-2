package allcontroller

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	ReplicasetController *(replicaset.ReplicasetController)
	HpaController        *(hpa.HPAController)
}

func (c *Controller) StartController() {
	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	select {}
}
