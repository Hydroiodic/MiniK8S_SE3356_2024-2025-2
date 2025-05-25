package mqtemplate

const (
	RabbitMQUrl = "amqp://guest:guest@localhost:5672/"

	UpdatePodQueueName = "updatePod"
	CreatePodQueueName = "createPod"

	KubeletCreatePodQueue    = "kubeletCreatePodQueue"
	KubeletStopPodQueue      = "kubeletStopPodQueue"
	KubeletDeletePodQueue    = "kubeletDeletePodQueue"
	KubeletCheckNowQueueName = "kubeletCheckNowQueue"

	KubeProxyCreateServiceQueue = "kubeProxyCreateServiceQueue"
	KubeProxyDeleteServiceQueue = "kubeProxyDeleteServiceQueue"

	CreateDnsQueueName  = "createDns"
	UpdateDnsQueueName  = "updateDns"
	DeleteDnsQueueName  = "deleteDns"
	UpdateHostQueueName = "updateHost"

	CreateReplicasetQueueName   = "createReplicasetQueue"
	DeleteReplicasetQueueName   = "deleteReplicasetQueue"
	ReplicasetCheckNowQueueName = "replicasetCheckNowQueue"
)
