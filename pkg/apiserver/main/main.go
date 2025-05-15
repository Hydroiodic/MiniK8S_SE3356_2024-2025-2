package main

import (
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver"
	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/apiserver/interfaces"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router.
	r := gin.Default()

	// Kubelet/Node operations.
	// NOTE: we do not disdinguish between kubelet and node in this version.
	r.POST(apiserver.KubeletRegisterURL, interfaces.KubeletRegister)
	r.POST(apiserver.KubeletHeartbeatURL, interfaces.KubeletHeartbeat)
	r.GET(apiserver.KubeletGetNodesURL, interfaces.GetNodes)
	r.GET(apiserver.KubeletDeletePodURL, interfaces.DeletePod)

	// Pod operations.
	r.POST(apiserver.PodCreateURL, interfaces.CreatePod)
	r.GET(apiserver.PodGetURL, interfaces.GetPods)
	r.POST(apiserver.PodAssignURL, interfaces.AssignPodToNode)
	r.POST(apiserver.PodDeleteURL, interfaces.DeletePod)

	// r.POST("/getOnePod", interfaces.GetOnePod)

	// r.POST("/updateHost", interfaces.HandleUpdateHost)

	// r.POST("/getObjectByType", interfaces.GetObjectByType)

	// r.POST("/createDnsFromFile", interfaces.HandleDnsCreate)
	// r.POST("/deleteDnsFromFile", interfaces.HandleDnsDelete)

	// r.POST("/createServiceFromFile", interfaces.CreateService)
	// r.POST("/deleteServiceFromFile", interfaces.DeleteService)
	// r.POST("/serviceCheckNow", interfaces.ServiceCheckNow)

	// r.POST("/createReplicasetFromFile", interfaces.CreateReplicaset)
	// r.POST("/deleteReplicasetFromFile", interfaces.DeleteReplicaset)
	// r.POST("/getOneReplicaset", interfaces.GetOneReplicaset)
	// r.POST("/changeReplicasetNum", interfaces.ChangeReplicasetNum)

	// r.POST("/createHPAFromFile", interfaces.CreateHPA)
	// r.POST("/deleteHPAFromFile", interfaces.DeleteHPA)

	// r.POST("/createCRFromFile", interfaces.CreateCR)
	// r.POST("/deleteCRFromFile", interfaces.DeleteCR)
	// r.POST("/getOneCR", interfaces.GetOneCR)

	// r.POST("/createFunctionFromFile", interfaces.CreateFunction)
	// r.POST("/deleteFunctionFromFile", interfaces.DeleteFunction)

	// r.POST("/createPVFromFile", interfaces.CreatePV)
	// r.POST("/deletePVFromFile", interfaces.DeletePV)
	// r.POST("/createPVCFromFile", interfaces.CreatePVC)
	// r.POST("/deletePVCFromFile", interfaces.DeletePVC)
	// r.POST("/getPVC", interfaces.GetPVC)

	// r.POST("/createJobFromFile", interfaces.CreateJob)

	// r.POST("/uploadJobOutputResult", interfaces.UploadJobOutputResult)
	// r.POST("/uploadJobErrorResult", interfaces.UploadJobErrorResult)
	// // r.POST("/applyFromFile", interfaces.ApplyFromFile)

	// go interfaces.CheckNodesHealthy()
	// go interfaces.CalculateServiceAndEps()

	// Start the server on port 8080.
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
