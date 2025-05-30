package allcontroller

import (
	gpujob "github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/gpu-job"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	ReplicasetController *(replicaset.ReplicasetController)
	HpaController        *(hpa.HPAController)
	JobController        *(gpujob.JobController)
	// dnsController        *(dnsController.DnsController)
}

func (c *Controller) StartController() {
	c.JobController = gpujob.NewJobController()
	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	go c.JobController.Start()
	select {}
}
