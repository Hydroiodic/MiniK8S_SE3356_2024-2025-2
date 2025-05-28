package mqtemplate

const (
	UpdatePodQueueName = "updatePod"
	CreatePodQueueName = "createPod"

	KubeletCreatePodQueue    = "kubeletCreatePodQueue"
	KubeletStopPodQueue      = "kubeletStopPodQueue"
	KubeletDeletePodQueue    = "kubeletDeletePodQueue"
	KubeletCheckNowQueueName = "kubeletCheckNowQueue"

	KubeProxyCreateServiceQueue = "kubeProxyCreateServiceQueue"
	KubeProxyDeleteServiceQueue = "kubeProxyDeleteServiceQueue"

	DnsCreatePod        = "DnsCreatePodQueue"
	DnsDeletePod        = "DnsDeletePodQueue"
	CreateDnsQueueName  = "createDns"
	UpdateDnsQueueName  = "updateDns"
	DeleteDnsQueueName  = "deleteDns"
	UpdateHostQueueName = "updateHost"

	CreateReplicasetQueueName   = "createReplicasetQueue"
	DeleteReplicasetQueueName   = "deleteReplicasetQueue"
	ReplicasetCheckNowQueueName = "replicasetCheckNowQueue"
)
