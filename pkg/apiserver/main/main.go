package main

import (
	"log"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver/interfaces"
	"github.com/gin-gonic/gin"
)

func declareGinServer() *gin.Engine {
	// Create a new Gin router.
	r := gin.Default()

	// Kubelet/Node operations.
	// NOTE: we do not disdinguish between kubelet and node in this version.
	r.POST(apiserver.KubeletRegisterURL, interfaces.KubeletRegister)
	r.POST(apiserver.KubeletHeartbeatURL, interfaces.KubeletHeartbeat)
	r.GET(apiserver.KubeletGetNodesURL, interfaces.GetNodes)

	// Pod operations.
	r.POST(apiserver.PodCreateURL, interfaces.CreatePod)
	r.GET(apiserver.PodGetURL, interfaces.GetPods)
	r.POST(apiserver.PodAssignURL, interfaces.AssignPodToNode)
	r.POST(apiserver.PodDeleteURL, interfaces.DeletePod)

	// DNS operations.
	r.POST(apiserver.DNSAddURL, interfaces.AddDNS)
	r.POST(apiserver.DNSDeleteURL, interfaces.DeleteDNS)
	r.GET(apiserver.DNSGetForwardingInfoURL, interfaces.GetForwardingInfo)

	// Replicaset operations.
	r.GET(apiserver.ReplicasetGetURL, interfaces.GetReplicasets)
	r.POST(apiserver.ReplicasetCreateURL, interfaces.CreateReplicaset)
	r.POST(apiserver.ReplicasetDeleteURL, interfaces.DeleteReplicasetFromEtcd)
	r.POST(apiserver.ReplicasetUpdateURL, interfaces.UpdateReplicaset)

	// HPA operations.
	r.GET(apiserver.HpaGetURL, interfaces.GetHpas)
	r.POST(apiserver.HpaCreateURL, interfaces.CreateHpa)

	// Service operations.
	r.POST(apiserver.ServiceCreateURL, interfaces.CreateService)
	r.POST(apiserver.ServiceDeleteURL, interfaces.DeleteService)
	r.GET(apiserver.ServiceGetURL, interfaces.GetAllService)

	// r.POST("/createCRFromFile", interfaces.CreateCR)
	// r.POST("/deleteCRFromFile", interfaces.DeleteCR)
	// r.POST("/getOneCR", interfaces.GetOneCR)

	// r.POST("/createFunctionFromFile", interfaces.CreateFunction)
	// r.POST("/deleteFunctionFromFile", interfaces.DeleteFunction)

	// r.POST("/createPVFromFile", interfaces.CreatePV)
	// r.POST("/deletePVFromFile", interfaces.DeletePV)
	// r.POST("/createPVCFromFile", interfaces.CreatePVC)
	// r.POST("/deletePVCFromFile", interfaces.DeletePVC)

	r.GET("/pvc/:namespace/:name", interfaces.GetPersistentVolumeClaim)
	r.GET("/pv/:name", interfaces.GetPersistentVolume)

	// r.POST("/createJobFromFile", interfaces.CreateJob)

	// r.POST("/uploadJobOutputResult", interfaces.UploadJobOutputResult)
	// r.POST("/uploadJobErrorResult", interfaces.UploadJobErrorResult)

	return r
}

func checkTimeout() {
	// An infinite loop to check for kubelet timeouts.
	for {
		// Check if the kubelet has timed out.
		if err := interfaces.CheckKubeletTimeout(); err != nil {
			// If there is an error, print it.
			log.Println(err)
		}

		// Sleep for 10 seconds before checking again.
		// NOTE: In `kubelet/runtime/status_controller.go`, the interval of heartbeat is 10s.
		//	     Here we use the same as the interval.
		time.Sleep(10 * time.Second)
	}
}

func checkService() {
	// An infinite loop to check for service status.
	for {
		// Check if the service has timed out.
		if err := interfaces.SyncEtcdServices(); err != nil {
			// If there is an error, print it.
			log.Println(err)
		}

		// Sleep for 3 seconds before checking again.
		time.Sleep(3 * time.Second)
	}
}

func main() {
	// Initialize internal services.
	interfaces.InitializeInternalServices()

	// Create a new Gin router.
	r := declareGinServer()

	// Go routine to check for some status.
	go checkTimeout()
	go checkService()

	// Start the Gin server on port 8080.
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
