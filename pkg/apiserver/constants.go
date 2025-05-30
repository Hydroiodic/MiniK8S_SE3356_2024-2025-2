package apiserver

const (
	// The URL of the APIServer.
	APIServerUrl        = "http://192.168.1.14:8080"
	APIServerIp         = "localhost"
	APIServerPort       = "5000"
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

	//gpujob operations
	GpuJobsCreateUrl      = "/gpujob/createJobs"
	GpuJobsGetUrl         = "/gpujob/getJobs"
	UploadJobOutputResult = "/gpujob/updateResult"
)
