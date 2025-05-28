package mqtemplate

const (
	RabbitMQUrl = "amqp://guest:guest@localhost:5672/"

	CreatePodQueueName = "createPod"

	KubeletCreatePodQueue = "kubeletCreatePodQueue"
	KubeletDeletePodQueue = "kubeletDeletePodQueue"

	KubeProxyCreateServiceQueue = "kubeProxyCreateServiceQueue"
	KubeProxyDeleteServiceQueue = "kubeProxyDeleteServiceQueue"
)
