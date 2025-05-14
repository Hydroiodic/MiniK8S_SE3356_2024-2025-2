package apiserver

const (
	// The URL of the APIServer.
	APIServerUrl        = "http://localhost:8080"
	KubeletRegisterURL  = "/kubelet/register"
	KubeletHeartbeatURL = "/kubelet/heartbeat"
	KubeletGetNodesURL  = "/kubelet/getNodes"

	// Pod operations.
	PodCreationURL = "/pod/createPod"
	PodGetURL      = "/pod/getPods"
	PodAssignURL   = "/pod/assignPodToNode"
)
