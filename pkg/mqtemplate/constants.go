package mqtemplate

const (
	RabbitMQUrl = "amqp://guest:guest@localhost:5672/"

	UpdatePodQueueName       = "updatePod"
	CreatePodQueueName       = "createPod"
	KubeletCreatePodQueue    = "kubeletCreatePodQueue"
	KubeletStopPodQueue      = "kubeletStopPodQueue"
	KubeletDeletePodQueue    = "kubeletDeletePodQueue"
	KubeletCheckNowQueueName = "kubeletCheckNowQueue"

	DnsCreatePod        = "DnsCreatePodQueue"
	DnsDeletePod        = "DnsDeletePodQueue"
	CreateDnsQueueName  = "createDns"
	UpdateDnsQueueName  = "updateDns"
	DeleteDnsQueueName  = "deleteDns"
	UpdateHostQueueName = "updateHost"

	CreateServiceQueueName   = "createServiceQueue"
	DeleteServiceQueueName   = "deleteServiceQueue"
	ServiceCheckNowQueueName = "serviceCheckNowQueue"

	CreateReplicasetQueueName   = "createReplicasetQueue"
	DeleteReplicasetQueueName   = "deleteReplicasetQueue"
	ReplicasetCheckNowQueueName = "replicasetCheckNowQueue"
)
