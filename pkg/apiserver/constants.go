package apiserver

const (
	// The URL of the APIServer.
	APIServerURL        = "192.168.1.6"
	APIServerPort       = "8080"
	KubeletRegisterURL  = "/kubelet/register"
	KubeletHeartbeatURL = "/kubelet/heartbeat"
	KubeletGetNodesURL  = "/kubelet/getNodes"
	KubeletDeletePodURL = "/kubelet/deletePod"

	// Pod operations.
	PodCreateURL = "/pod/createPod"
	PodGetURL    = "/pod/getPods"
	PodAssignURL = "/pod/assignPodToNode"
	PodDeleteURL = "/pod/deletePod"

	// TODO: Service operations.
	ServiceCreateURL = "/service/createService"
	ServiceGetURL    = "/service/getServices" // GetAllService
	ServiceDeleteURL = "/service/deleteService"

	// DNS operations.
	DNSAddURL               = "/dns/addDNS"
	DNSGetURL               = "/dns/getDNS"
	DNSDeleteURL            = "/dns/deleteDNS"
	DNSGetForwardingInfoURL = "/dns/getForwardingInfo"

	// Replicaset operations.
	ReplicasetCreateURL = "/replicaset/createReplicasets"
	ReplicasetGetURL    = "/replicaset/getReplicasets"
	ReplicasetUpdateURL = "/replicaset/updateReplicaset"
	ReplicasetDeleteURL = "/replicaset/deleteReplicaset"

	// HPA operations.
	HpaGetURL    = "/hpa/getHpas"
	HpaCreateURL = "/hpa/createHpa"

	PVGetURL    = "/pv"
	PVCreateURL = "/pv/createPV"
	PVDeleteURL = "/pv/deletePV"

	PVCGetURL        = "/pvc"
	PVClaimCreateURL = "/pvc/createPVC"
	PVClaimDeleteURL = "/pvc/deletePVC"

	// GPUJob operations
	GPUJobsCreateURL = "/gpujob/createJobs"
	GPUJobsGetURL    = "/gpujob/getJobs"
	UploadResultURL  = "/gpujob/updateResult"
)
