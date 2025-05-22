package apiserver

const (
	// The URL of the APIServer.
	APIServerUrl        = "http://localhost:8080"
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
	DNSAddURL        = "/dns/addDNS"
	DNSDeleteURL     = "/dns/deleteDNS"
	DNSGetResolveURL = "/dns/getDNSResolve"

	ReplicasetCreateURL = "/replicaset/createReplicasets"
	ReplicasetGetURL    = "/replicaset/getReplicasets"
	ReplicasetUpdateURL = "/replicaset/updateReplicaset"
	ReplicasetDeleteURL = "/replicaset/deleteReplicaset"

	HpaGetURL    = "/hpa/getHpas"
	HpaCreateURL = "hpa/createHpa"
)
