package allcontroller

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/hpa"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/controller/replicaset"
)

type Controller struct {
	// endpointsController *(endpointsController.EndpointsController)
	// servicesController *(servicesController.ServicesController)
	ReplicasetController *(replicaset.ReplicasetController)
	HpaController        *(hpa.HPAController)
	// dnsController        *(dnsController.DnsController)
}

func (c *Controller) Init() {

}

// NOTE: 由于DNS/反向代理需要，nginx必须部署在和controller相同的
func (c *Controller) StartController() {

	go c.ReplicasetController.Start()
	go c.HpaController.Start()
	select {}

}
